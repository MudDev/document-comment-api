package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func (s *SQLite) Initialize(dbNames ...string) error {
	dbName := "document-drafts.db" // Default database name
	if len(dbNames) > 0 {
		dbName = dbNames[0] // If a name is provided, use it instead
	}
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		log.Fatal(err)
		panic("Failed to initialize document service")
	} else {
		log.Println("Document Service initialized with database:", dbName)
	}

	setupTables(db)

	s.DB = db

	return nil
}

func setupTables(db *sql.DB) {
	createTableStatements := []string{
		`CREATE TABLE IF NOT EXISTS documents (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			Name TEXT NOT NULL,
			LatestVersion INTEGER NOT NULL DEFAULT 1,
			CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS drafts (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			DocumentId INTEGER NOT NULL,
			Content TEXT,
			VersionNumber INTEGER NOT NULL,
			CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (DocumentId) REFERENCES documents(Id)
		);`,
		`CREATE TABLE IF NOT EXISTS comments (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			DraftId INTEGER NOT NULL,
			UserId INTEGER NOT NULL,
			Text TEXT NOT NULL,
			ParentCommentId INTEGER,
			CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (DraftId) REFERENCES drafts(Id),
			FOREIGN KEY (ParentCommentId) REFERENCES comments(Id)
		);`,
		`CREATE TABLE IF NOT EXISTS reactions (
			Id INTEGER PRIMARY KEY AUTOINCREMENT,
			CommentId INTEGER NOT NULL,
			UserId INTEGER NOT NULL,
			Emoji TEXT NOT NULL,
			CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (CommentId) REFERENCES comments(Id)
		);`,
	}

	for _, stmt := range createTableStatements {
		_, err := db.Exec(stmt)
		if err != nil {
			log.Fatalf("Failed to execute statement: %v, error: %v", stmt, err)
		}
	}
}
