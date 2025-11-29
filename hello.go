package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/TNorbury/go-bluesky"
	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
)

// Unique diary entry
// TODO: add date...
type DiaryEntry struct {
	name   string
	url    string
	rating string
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(fmt.Sprintf("Error loading .env file. %e", err))
	}

	entries := scrapeLetterboxDiary()
	entry := entries[0]

	client := connectToBluesky()
	client.CustomCall(func(api *xrpc.Client) error {

		post := bsky.FeedPost{}
		// TODO: Update to I give ... a 1/0
		post.Text = fmt.Sprintf("I give %s a %s", entry.name, entry.rating)
		post.LexiconTypeID = "app.bsky.feed.post"
		post.CreatedAt = time.Now().UTC().Format(util.ISO8601)

		postInput := &atproto.RepoCreateRecord_Input{
			Collection: "app.bsky.feed.post",
			Repo:       api.Auth.Did,
			Record:     &lexutil.LexiconTypeDecoder{Val: &post},
		}

		out, err := atproto.RepoCreateRecord(context.Background(), api, postInput)

		if err != nil {
			panic(fmt.Errorf("Unable to post, %v", err))
		}

		fmt.Printf("Posted: %v, %v", out.Cid, out.Uri)

		return err
	})

}

func scrapeLetterboxDiary() []DiaryEntry {
	diaryPageCollector := colly.NewCollector()

	var entries []DiaryEntry

	diaryPageCollector.OnHTML(".diary-entry-row", func(e *colly.HTMLElement) {
		entry := DiaryEntry{}

		entry.name = e.ChildText(".name")

		urls := e.ChildAttrs("a", "href")
		// for (u :)
		for _, u := range urls {
			if strings.Contains(u, "/film/") {
				entry.url = u
				break
			}
		}

		entries = append(entries, entry)

	})

	diaryPageCollector.Visit("https://letterboxd.com/tylernorbury/diary/")
	// fmt.Println("Visit done!")
	fmt.Printf("%d entries\n", len(entries))

	for i, entry := range entries {

		entryPageCollector := colly.NewCollector()

		entryPageCollector.OnHTML(".js-review-body", func(e *colly.HTMLElement) {
			rating := e.ChildText("p")[0:1]

			entry.rating = rating
			entries[i] = entry
		})

		entryPageCollector.Visit(fmt.Sprintf("https://letterboxd.com/%s", entry.url))
	}

	for _, entry := range entries {
		fmt.Printf("I give %s a %s\n", entry.name, entry.rating)
	}

	return entries
}

func connectToBluesky() *bluesky.Client {
	ctx := context.Background()
	client, err := bluesky.Dial(ctx, bluesky.ServerBskySocial)
	if err != nil {
		panic(err)
	}

	defer client.Close()

	err = client.Login(ctx, os.Getenv("BLUESKY_USERNAME"), os.Getenv("BLUESKY_APPKEY"))
	switch {
	case errors.Is(err, bluesky.ErrMasterCredentials):
		panic("You're not allowed to use your full-access credentials, please create an appkey")
	case errors.Is(err, bluesky.ErrLoginUnauthorized):
		panic("Username of application password seems incorrect, please double check")
	case err != nil:
		panic("Something else went wrong, please look at the returned error")
	}

	return client
}
