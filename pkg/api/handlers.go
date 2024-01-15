package api

import (
	"documentapi/pkg/common"
	"documentapi/pkg/database"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (a *API) addDraft(w http.ResponseWriter, r *http.Request) {
	var draft common.Draft
	if err := json.NewDecoder(r.Body).Decode(&draft); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	newDraft := common.Draft{
		Name:    draft.Name,
		Content: draft.Content,
	}
	if err := a.SQL.CreateDraft(newDraft); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Draft added successfully"})
}

func (a *API) getMostRecentDrafts(w http.ResponseWriter, r *http.Request) {
	limitParam := r.URL.Query().Get("limit")
	limit := 1 // Default limit
	if limitParam != "" {
		var err error
		limit, err = strconv.Atoi(limitParam)
		if err != nil {
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
	}

	recentDrafts, err := a.SQL.GetLatestDrafts(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(recentDrafts)
}

func (a *API) searchDrafts(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("text")
	if searchQuery == "" {
		http.Error(w, "text parameter is required", http.StatusBadRequest)
		return
	}

	drafts, err := a.SQL.SearchDrafts(searchQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(drafts)
}

func (a *API) getDocumentsLatestVersions(w http.ResponseWriter, r *http.Request) {
	documents, err := a.SQL.GetAllDocumentsLatestVersions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(documents)
}

func (a *API) addComment(w http.ResponseWriter, r *http.Request) {
	var comment database.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	commentId, err := a.SQL.AddCommentToDraft(comment)
	if err != nil {
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	response := NewCommentResult{
		Message: "Comment added successfully",
		Id:      commentId,
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (a *API) getCommentsAndReactions(w http.ResponseWriter, r *http.Request) {
	draftIdStr := r.URL.Query().Get("draftId")
	if draftIdStr == "" {
		http.Error(w, "draftId query parameter is required", http.StatusBadRequest)
		return
	}

	draftId, err := strconv.Atoi(draftIdStr)
	if err != nil {
		http.Error(w, "Invalid draftId", http.StatusBadRequest)
		return
	}

	comments, err := a.SQL.GetCommentsAndReactionsByDraftId(draftId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(comments)
}

func (a *API) addReaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentIdStr, ok := vars["commentId"]
	if !ok {
		http.Error(w, "Comment ID is required", http.StatusBadRequest)
		return
	}
	commentId, err := strconv.Atoi(commentIdStr)
	if err != nil {
		http.Error(w, "Invalid Comment ID", http.StatusBadRequest)
		return
	}

	newReaction := common.Reaction{}
	if err := json.NewDecoder(r.Body).Decode(&newReaction); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !isOnlySupportedEmojis(newReaction.Emoji) {
		http.Error(w, "Invalid emoji", http.StatusBadRequest)
		return
	}

	reaction := common.Reaction{
		Id:     commentId,
		Emoji:  newReaction.Emoji,
		UserId: newReaction.UserId,
	}

	if err := a.SQL.AddReactionToComment(reaction); err != nil {
		http.Error(w, "Failed to add reaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Reaction added successfully"})
}

func isOnlySupportedEmojis(s string) bool {
	for _, r := range s {
		if !isSupportedEmoji(r) {
			return false
		}
	}
	return true
}

func isSupportedEmoji(r rune) bool {
	return (r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
		(r >= 0x1F300 && r <= 0x1F5FF) || // Misc Symbols and Pictographs
		(r >= 0x1F680 && r <= 0x1F6FF) || // Transport and Map
		(r >= 0x2600 && r <= 0x26FF) || // Misc symbols
		(r >= 0x2700 && r <= 0x27BF) || // Dingbats
		(r >= 0xFE00 && r <= 0xFE0F) || // Variation Selectors
		(r >= 0x1F900 && r <= 0x1F9FF) || // Supplemental Symbols and Pictographs
		(r >= 0x1F1E6 && r <= 0x1F1FF) // Regional indicator symbols (for flag emojis)
}
