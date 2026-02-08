package main

import (
	"fmt"
	"time"

	"tnorbury/letterboxd-bluesky/database"
	"tnorbury/letterboxd-bluesky/letterboxd"

	"tnorbury/letterboxd-bluesky/bluesky"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(fmt.Sprintf("Error loading .env file. %e", err))
	}

	db := database.InitDb()

	checkForNewEntries(db)

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for t := range ticker.C {
		fmt.Printf("Load Letterbox @ %v\n", t)
		ctrl := checkForNewEntries(db)
		switch ctrl {
		case 1:
			continue
		}
	}

}

func checkForNewEntries(db *database.Db) int {
	entries, err := letterboxd.ScrapeLetterboxDiary(0, db)
	if err != nil {
		fmt.Printf("Error scrapping letterbox: %e\n", err)
		return 1
	}

	if len(entries) >= 1 {
		client := bluesky.ConnectToBluesky()

		for i := len(entries) - 1; i >= 0; i = i - 1 {
			entry := entries[i]

			postErr := bluesky.PostEntry(client, entry, db)
			if postErr != nil {
				panic(postErr)
			}

			// wait 1/2 sec before making next post
			if i < len(entries)-1 {
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
	return 0
}
