package database

import (
	"database/sql"
	"documentapi/pkg/common"
	"time"
)

type SQLite struct {
	*sql.DB
}

type Document struct {
	Id            int       `json:"id"`
	Name          string    `json:"name"`
	LatestVersion int       `json:"latestVersion"`
	CreatedAt     time.Time `json:"createdAt"`
}

type Draft struct {
	Id            int       `json:"id"`
	DocumentId    int       `json:"documentId"`
	Content       string    `json:"content"`
	VersionNumber int       `json:"versionNumber"`
	CreatedAt     time.Time `json:"createdAt"`
}

type Comment struct {
	Id              int       `json:"id"`
	DraftId         int       `json:"draftId"`
	UserId          int       `json:"userId"`
	Text            string    `json:"text"`
	ParentCommentId *int      `json:"parentCommentId"`
	CreatedAt       time.Time `json:"createdAt"`
}

type CommentWithReactions struct {
	Id              int               `json:"id"`
	UserId          int               `json:"userId"`
	Text            string            `json:"text"`
	ParentCommentId *int              `json:"parentCommentId,omitempty"`
	CreatedAt       time.Time         `json:"createdAt"`
	Reactions       []common.Reaction `json:"reactions"`
}
