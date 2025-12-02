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
		entries := letterboxd.ScrapeLetterboxDiary(0, db)

		if len(entries) >= 1 {
	client := bluesky.ConnectToBluesky()

	for i, entry := range entries {
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

}
