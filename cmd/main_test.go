package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"documentapi/pkg/api"
	"documentapi/pkg/common"
	"documentapi/pkg/database"
)

func setup() (*database.SQLite, *api.API, string) {
	sqlService, dbName := setupTestDB()
	apiService := &api.API{}

	apiService.Initialize(sqlService)

	return sqlService, apiService, dbName
}

func setupTestDB() (*database.SQLite, string) {
	// Generate a unique filename for the test database
	dbName := fmt.Sprintf("testdb_%v.db", time.Now().UnixNano())

	sqlService := &database.SQLite{}
	err := sqlService.Initialize(dbName) // Adjust the Initialize method to accept a db name
	if err != nil {
		log.Fatalf("Failed to initialize test database: %v", err)
	}

	return sqlService, dbName
}

func teardown(sqlService *database.SQLite, dbName string) {
	// Close the database connection
	if sqlService.DB != nil {
		err := sqlService.DB.Close()
		if err != nil {
			log.Fatalf("Failed to close database connection: %v", err)
		}
	}

	// Delete the test database file
	err := os.Remove(dbName)
	if err != nil {
		log.Fatalf("Failed to delete test database file: %v", err)
	}
}

func createDraft(serverURL, draftName, draftContent string) ([]byte, error) {
	requestBody := strings.NewReader(fmt.Sprintf(`{"name": "%s", "content": "%s"}`, draftName, draftContent))
	resp, err := http.Post(serverURL+"/api/drafts", "application/json", requestBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create draft, status: %s, body: %s", resp.Status, string(body))
	}

	return body, nil
}

func getJSON(t *testing.T, url string, target interface{}) {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
}

func createComment(serverURL string, draftId, userId int, text string) (string, int, error) {
	commentData := database.Comment{
		DraftId: draftId,
		UserId:  userId,
		Text:    text,
	}

	commentDataBytes, err := json.Marshal(commentData)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal comment data: %v", err)
	}

	// Make a POST request to add a comment
	resp, err := http.Post(serverURL+"/api/comments", "application/json", bytes.NewReader(commentDataBytes))
	if err != nil {
		return "", 0, fmt.Errorf("failed to make POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", 0, fmt.Errorf("expected status Created; got %v", resp.Status)
	}

	// Assuming the API returns the created comment ID in the response body
	var result = api.NewCommentResult{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, fmt.Errorf("failed to decode response body: %v", err)
	}

	return result.Message, int(result.Id), nil
}

func TestAddDraft(t *testing.T) {
	sqlService, apiService, dbName := setup()
	defer teardown(sqlService, dbName)

	server := httptest.NewServer(apiService.Router)
	defer server.Close()

	body, err := createDraft(server.URL, "Test Draft", "Draft content")
	if err != nil {
		t.Fatalf("Failed to create draft: %v", err)
	}

	var respBody map[string]string
	json.Unmarshal(body, &respBody)
	if msg, ok := respBody["message"]; !ok || msg != "Draft added successfully" {
		t.Errorf("Unexpected response body: %s", body)
	}
}

func TestGetMostRecentDrafts(t *testing.T) {
	sqlService, apiService, dbName := setup()
	defer teardown(sqlService, dbName)

	server := httptest.NewServer(apiService.Router)
	defer server.Close()

	// Create a few drafts for testing
	_, err := createDraft(server.URL, "Test Draft 1", "Draft content 1")
	if err != nil {
		t.Fatalf("Failed to create draft: %v", err)
	}
	_, err = createDraft(server.URL, "Test Draft 2", "Draft content 2")
	if err != nil {
		t.Fatalf("Failed to create draft: %v", err)
	}

	var drafts []database.Draft
	getJSON(t, server.URL+"/api/drafts", &drafts)

	if len(drafts) < 2 {
		t.Errorf("Expected at least 2 drafts, got %d", len(drafts))
	}

	if drafts[1].Content == "Draft content 1" || drafts[0].Content == "Draft content 2" {
		t.Errorf("Drafts not in expected order or missing. Received: %v", drafts)
	}
}

func TestSearchDrafts(t *testing.T) {
	sqlService, apiService, dbName := setup()
	defer teardown(sqlService, dbName)

	server := httptest.NewServer(apiService.Router)
	defer server.Close()

	// Create drafts for testing
	_, err := createDraft(server.URL, "Custodia rocks", "Content about custodia bank")
	if err != nil {
		t.Fatalf("Failed to create draft: %v", err)
	}
	_, err = createDraft(server.URL, "Search Test Draft", "Content about API testing")
	if err != nil {
		t.Fatalf("Failed to create draft: %v", err)
	}

	searchQuery := "custodia bank"

	resp, err := http.Get(server.URL + "/api/drafts/search?text=" + url.QueryEscape(searchQuery))
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	var drafts []database.Draft
	if err := json.NewDecoder(resp.Body).Decode(&drafts); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if len(drafts) == 0 {
		t.Errorf("Expected to find drafts, but none were returned")
	}

	if !strings.Contains(drafts[0].Content, searchQuery) {
		t.Errorf("The returned drafts do not match the search query. Received: %v", drafts)
	}
}

func TestGetDocumentsLatestVersions(t *testing.T) {
	sqlService, apiService, dbName := setup()
	defer teardown(sqlService, dbName)

	server := httptest.NewServer(apiService.Router)
	defer server.Close()

	// Create drafts with diffrent names for testing
	_, err := createDraft(server.URL, "Custodia rocks", "Content about custodia bank")
	if err != nil {
		t.Fatalf("Failed to create draft: %v", err)
	}
	_, err = createDraft(server.URL, "Custodia rocks", "Content about how custodia bank is the best bitcoin vault")
	if err != nil {
		t.Fatalf("Failed to create draft: %v", err)
	}
	_, err = createDraft(server.URL, "Rough Draft", "Satoshi's Draft Whitepaper")
	if err != nil {
		t.Fatalf("Failed to create draft: %v", err)
	}

	// Make a GET request to the /api/documents/latest endpoint
	resp, err := http.Get(server.URL + "/api/documents/latest")
	if err != nil {
		t.Fatalf("Failed to make GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	var documents []database.Document
	if err := json.NewDecoder(resp.Body).Decode(&documents); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if len(documents) < 2 {
		t.Errorf("Expected at least 2 documents, got %d", len(documents))
	}

	for _, doc := range documents {
		if doc.Name == "Custodia rocks" && doc.LatestVersion != 2 {
			t.Errorf("Expected latest version of 'Custodia rocks' to be 2, got %d", doc.LatestVersion)
		}
	}
}

func TestAddComment(t *testing.T) {
	sqlService, apiService, dbName := setup()
	defer teardown(sqlService, dbName)

	server := httptest.NewServer(apiService.Router)
	defer server.Close()

	// Create a draft for testing
	_, err := createDraft(server.URL, "Test Draft for Comment", "Draft content for comment")
	if err != nil {
		t.Fatalf("Failed to create draft: %v", err)
	}

	var drafts []database.Draft
	getJSON(t, server.URL+"/api/drafts", &drafts)

	respMessage, _, err := createComment(server.URL, drafts[0].Id, 1, "This is a test comment")
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	if respMessage != "Comment added successfully" {
		t.Errorf("Unexpected response message: %s", respMessage)
	}
}

func TestAddReaction(t *testing.T) {
	sqlService, apiService, dbName := setup()
	defer teardown(sqlService, dbName)

	server := httptest.NewServer(apiService.Router)
	defer server.Close()

	// Create a draft and a comment for testing
	_, err := createDraft(server.URL, "Gensis Block", "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks")
	if err != nil {
		t.Fatalf("Failed to create draft: %v", err)
	}

	var drafts []database.Draft
	getJSON(t, server.URL+"/api/drafts", &drafts)

	_, commentId, err := createComment(server.URL, drafts[0].Id, 1, "This is the way")
	if err != nil {
		t.Fatalf("Failed to create comment: %v", err)
	}

	// Prepare reaction data
	reactionData := common.Reaction{
		Id:     commentId,
		Emoji:  "👍",
		UserId: 1,
	}

	reactionDataBytes, err := json.Marshal(reactionData)
	if err != nil {
		t.Fatalf("Failed to marshal reaction data: %v", err)
	}

	reactionURL := fmt.Sprintf("%s/api/comment/%d/reaction", server.URL, commentId)
	resp, err := http.Post(reactionURL, "application/json", bytes.NewReader(reactionDataBytes))
	if err != nil {
		t.Fatalf("Failed to make POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status Created; got %v", resp.Status)
	}

	// Check the response body
	var respBody map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if msg, ok := respBody["message"]; !ok || msg != "Reaction added successfully" {
		t.Errorf("Unexpected response body: %s", respBody)
	}
}
