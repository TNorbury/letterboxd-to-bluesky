package letterboxd

import (
	"fmt"
	"os"
	"strings"
	"time"
	"tnorbury/letterboxd-bluesky/database"
	"tnorbury/letterboxd-bluesky/models"

	"github.com/gocolly/colly"
)

func ScrapeLetterboxDiary(maxEntries int, db *database.Db) []models.DiaryEntry {
	diaryPageCollector := colly.NewCollector()

	var entries []models.DiaryEntry

	doneReadingEntries := false

	diaryPageCollector.OnHTML(".diary-entry-row", func(e *colly.HTMLElement) {
		if doneReadingEntries {
			return
		}
		entry := models.DiaryEntry{}

		entry.Name = e.ChildText(".name")

		urls := e.ChildAttrs("a", "href")
		for _, u := range urls {
			if strings.Contains(u, "/film/") {
				entry.Url = u
				break
			}
		}

		entries = append(entries, entry)

		if maxEntries > 0 && len(entries) >= maxEntries {
			doneReadingEntries = true
		}

	})

	letterboxdUser := os.Getenv("LETTERBOXD_USERNAME")
	err := diaryPageCollector.Visit(fmt.Sprintf("https://letterboxd.com/%s/diary/", letterboxdUser))
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d entries\n", len(entries))
	// time.Parse(time.DateOnly, "")
	foundMatchingEntry := false
	matchingEntryIdx := -1

	for i, entry := range entries {
		entryPageCollector := colly.NewCollector()

		entryPageCollector.OnHTML(".js-review-body", func(e *colly.HTMLElement) {
			// first character, 1 or 0, is the rating
			rating := e.ChildText("p")[0:1]

			entry.Rating = rating
			entries[i] = entry
		})

		entryPageCollector.OnHTML("meta", func(e *colly.HTMLElement) {
			date := e.Attr("content")
			if date == "" {
				return
			}

			tt, err := time.Parse(time.DateOnly, date)
			if err != nil {
				return
			}

			entry.Date = tt
			entries[i] = entry
		})

		entryPageCollector.Visit(fmt.Sprintf("https://letterboxd.com/%s", entry.Url))

		// check to see if this entry is already logged
		if db.HasMatchingEntry(entries[i]) {
			foundMatchingEntry = true
			matchingEntryIdx = i
			break
		}
	}

	if foundMatchingEntry {
		if matchingEntryIdx == 0 {
			entries = []models.DiaryEntry{}
		} else {
			entries = entries[0 : matchingEntryIdx-1]
		}
	}

	for _, entry := range entries {
		fmt.Printf("I give %s a %s\n", entry.Name, entry.Rating)
	}

	var entriesToReturn []models.DiaryEntry

	noEntriesBeforeStr := os.Getenv("NO_ENTRIES_BEFORE")
	if noEntriesBeforeStr != "" {
		noEntriesBefore, _ := time.Parse(time.DateOnly, noEntriesBeforeStr)

		for _, entry := range entries {

			if entry.Date.After(noEntriesBefore) {
				entriesToReturn = append(entriesToReturn, entry)
			}
		}

	} else {
		entriesToReturn = entries
	}

	return entriesToReturn
}
