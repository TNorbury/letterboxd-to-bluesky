package database

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"tnorbury/letterboxd-bluesky/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Db struct {
	Gorm *gorm.DB
	Sql  *sql.DB
}

func InitDb() *Db {
	dbHost := os.Getenv("DB_HOST")
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		panic(err)
	}

	dbUser := os.Getenv("DB_USER")
	dbPw := os.Getenv("DB_PW")
	dbName := os.Getenv("DB_NAME")

	var psqlConn string

	if dbPw != "" {
		psqlConn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPw, dbName)
	} else {
		psqlConn = fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbName)
	}

	sqldb, err := sql.Open("pgx", psqlConn)
	if err != nil {
		panic(err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqldb,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db := &Db{gormDB, sqldb}

	err = db.Sql.Ping()
	if err != nil {
		panic(err)
	}

	// done connecting, set up table
	err = db.Gorm.AutoMigrate(&models.DiaryEntry{})
	if err != nil {
		panic(err)
	}
	return db
}

func (db *Db) GetEntriesWithName(name string) []models.DiaryEntry {
	var entries []models.DiaryEntry
	db.Gorm.Where("name = ?", name).Find(&entries)

	fmt.Printf("found %d entries in db", len(entries))

	return entries
}

func (db *Db) HasMatchingEntry(entry models.DiaryEntry) bool {
	var entryMatch []models.DiaryEntry
	db.Gorm.Where("name = ?", entry.Name).Where("date = ?", entry.Date).Limit(1).Find(&entryMatch)

	return len(entryMatch) >= 1
}

func (db *Db) AddEntry(entry models.DiaryEntry) {
	db.Gorm.Create(&entry)
}
