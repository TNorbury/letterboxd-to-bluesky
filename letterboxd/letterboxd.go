package letterboxd

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"tnorbury/letterboxd-bluesky/database"
	"tnorbury/letterboxd-bluesky/models"

	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

func ScrapeLetterboxDiary(maxEntries int, db *database.Db) ([]models.DiaryEntry, error) {
	fakeChrome := req.DefaultClient().ImpersonateChrome()
	diaryPageCollector := colly.NewCollector(
		colly.UserAgent(fakeChrome.Headers.Get("user-agent")),
	)

	diaryPageCollector.SetClient(&http.Client{
		Transport: fakeChrome.Transport,
	})

	var entries []models.DiaryEntry

	doneReadingEntries := false

	diaryPageCollector.OnHTML(".diary-entry-row", func(e *colly.HTMLElement) {
		if doneReadingEntries {
			return
		}
		entry := models.DiaryEntry{}

		name := e.ChildText(".name")
		year := e.ChildText(".col-releaseyear")

		entry.Name = fmt.Sprintf("%s (%s)", name, year)

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
		return nil, err
	}

	fmt.Printf("%d entries\n", len(entries))
	// time.Parse(time.DateOnly, "")
	foundMatchingEntry := false
	matchingEntryIdx := -1

	for i, entry := range entries {
		entryPageCollector := colly.NewCollector(
			colly.UserAgent(fakeChrome.Headers.Get("user-agent")),
		)
		entryPageCollector.SetClient(&http.Client{
			Transport: fakeChrome.Transport,
		})

		entryPageCollector.OnHTML(".js-review-body", func(e *colly.HTMLElement) {
			// first character, 1 or 0, is the rating
			rating := e.ChildText("p")[0:1]

			entry.Rating = rating
			entries[i] = entry
		})

		entryPageCollector.OnHTML(".view-date", func(e *colly.HTMLElement) {
			watchedDateFull := e.ChildTexts("a")
			if len(watchedDateFull) != 3 {
				return
			}

			day := watchedDateFull[0]
			year := watchedDateFull[2]
			var month string

			switch watchedDateFull[1] {
			case "Jan":
				month = "01"
			case "Feb":
				month = "02"
			case "Mar":
				month = "03"
			case "Apr":
				month = "04"
			case "May":
				month = "05"
			case "Jun":
				month = "06"
			case "Jul":
				month = "07"
			case "Aug":
				month = "08"
			case "Sep":
				month = "09"
			case "Oct":
				month = "10"
			case "Nov":
				month = "11"
			case "Dec":
				month = "12"
			}

			dateStr := fmt.Sprintf("%s-%s-%s", year, month, day)

			tt, err := time.Parse(time.DateOnly, dateStr)
			if err != nil {
				return
			}

			entry.Date = tt
			entries[i] = entry
		})

		entryPageCollector.Visit(fmt.Sprintf("https://letterboxd.com/%s", entry.Url))

		// check to see if this entry is already logged
		if db.HasMatchingEntry(entries[i]) {
			name := entries[1].Name
			fmt.Printf("Matching Entry: %s\n", name)
			foundMatchingEntry = true
			matchingEntryIdx = i
			break
		}
	}

	if foundMatchingEntry {
		if matchingEntryIdx == 0 {
			entries = []models.DiaryEntry{}
		} else {
			entries = entries[0:matchingEntryIdx]
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

	return entriesToReturn, nil
}
