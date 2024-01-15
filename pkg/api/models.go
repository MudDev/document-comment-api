package api

import (
	"documentapi/pkg/database"

	"github.com/gorilla/mux"
)

type API struct {
	Router *mux.Router
	SQL    *database.SQLite
}

type NewCommentResult struct {
	Id      int64  `json:"id"`
	Message string `json:"message"`
}
