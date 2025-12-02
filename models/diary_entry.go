package models

import (
	"time"

	"gorm.io/gorm"
)

// Unique diary entry
type DiaryEntry struct {
	gorm.Model
	Name   string
	Url    string
	Rating string
	Date   time.Time
}
