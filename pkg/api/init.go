package api

import (
	"documentapi/pkg/database"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (a *API) Initialize(sql *database.SQLite) {
	a.SQL = sql
	a.Router = mux.NewRouter()

	a.Router.HandleFunc("/api/drafts", a.addDraft).Methods("POST")
	a.Router.HandleFunc("/api/drafts", a.getMostRecentDrafts).Methods("GET")
	a.Router.HandleFunc("/api/drafts/search", a.searchDrafts).Methods("GET")
	a.Router.HandleFunc("/api/drafts/comments-reactions", a.getCommentsAndReactions).Methods("GET")
	a.Router.HandleFunc("/api/documents/latest", a.getDocumentsLatestVersions).Methods("GET")
	a.Router.HandleFunc("/api/comments", a.addComment).Methods("POST")
	a.Router.HandleFunc("/api/comment/{commentId}/reaction", a.addReaction).Methods("POST")
}

func (a *API) StartAPI() {
	address := ":8080"

	log.Printf("Starting API server on %s", address)

	err := http.ListenAndServe(address, a.Router)
	if err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}
}
