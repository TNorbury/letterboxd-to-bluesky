package main

import (
	"fmt"

	"tnorbury/letterboxd-bluesky/letterboxd"

	"tnorbury/letterboxd-bluesky/bluesky"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(fmt.Sprintf("Error loading .env file. %e", err))
	}

	entries := letterboxd.ScrapeLetterboxDiary(1)
	entry := entries[0]

	client := bluesky.ConnectToBluesky()

	postErr := bluesky.PostEntry(client, entry)
	if postErr != nil {
		panic(postErr)
	}

}
