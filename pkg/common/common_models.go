package common

import "time"

type Draft struct {
	Name          string `json:"name"`
	Content       string `json:"content"`
	VersionNumber int    `json:"versionNumber"`
}

type Reaction struct {
	Id        int       `json:"id"`
	UserId    int       `json:"userId"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"createdAt"`
}
