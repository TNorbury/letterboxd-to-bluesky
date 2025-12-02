package bluesky

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
	"tnorbury/letterboxd-bluesky/letterboxd"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/util"
	"github.com/bluesky-social/indigo/xrpc"

	"github.com/TNorbury/go-bluesky"
)

// Connect and login to bluesky using credentials provided in the .env
// Returns client for further interactions
func ConnectToBluesky() *bluesky.Client {
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

func PostEntry(client *bluesky.Client, entry letterboxd.DiaryEntry) error {
	return client.CustomCall(func(api *xrpc.Client) error {

		post := bsky.FeedPost{}
		post.Text = fmt.Sprintf("I give %s a %s", entry.Name, entry.Rating)
		post.LexiconTypeID = "app.bsky.feed.post"
		post.CreatedAt = time.Now().UTC().Format(util.ISO8601)
		postInput := &atproto.RepoCreateRecord_Input{
			Collection: "app.bsky.feed.post",
			Repo:       api.Auth.Did,
			Record:     &lexutil.LexiconTypeDecoder{Val: &post},
		}

		out, err := atproto.RepoCreateRecord(context.Background(), api, postInput)

		if err != nil {
			panic(fmt.Errorf("unable to post, %e", err))
		}

		fmt.Printf("Posted: %s -- %v, %v", entry.Name, out.Cid, out.Uri)

		return err
	})
}
