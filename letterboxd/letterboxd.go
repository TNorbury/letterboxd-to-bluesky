package letterboxd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gocolly/colly"
)

// Unique diary entry
// TODO: add date...
type DiaryEntry struct {
	Name   string
	Url    string
	Rating string
}

func ScrapeLetterboxDiary() []DiaryEntry {
	diaryPageCollector := colly.NewCollector()

	var entries []DiaryEntry

	diaryPageCollector.OnHTML(".diary-entry-row", func(e *colly.HTMLElement) {
		entry := DiaryEntry{}

		entry.Name = e.ChildText(".name")

		urls := e.ChildAttrs("a", "href")
		for _, u := range urls {
			if strings.Contains(u, "/film/") {
				entry.Url = u
				break
			}
		}

		entries = append(entries, entry)

	})

	letterboxdUser := os.Getenv("LETTERBOXD_USERNAME")
	err := diaryPageCollector.Visit(fmt.Sprintf("https://letterboxd.com/%s/diary/", letterboxdUser))
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d entries\n", len(entries))

	for i, entry := range entries {

		entryPageCollector := colly.NewCollector()

		entryPageCollector.OnHTML(".js-review-body", func(e *colly.HTMLElement) {
			// first character, 1 or 0, is the rating
			rating := e.ChildText("p")[0:1]

			entry.Rating = rating
			entries[i] = entry
		})

		entryPageCollector.Visit(fmt.Sprintf("https://letterboxd.com/%s", entry.Url))
	}

	for _, entry := range entries {
		fmt.Printf("I give %s a %s\n", entry.Name, entry.Rating)
	}

	return entries
}
