package database

import (
	"database/sql"
	"documentapi/pkg/common"
	"time"
)

// createDocument - Creates a new document or increments the version.
func (s *SQLite) createDocument(name string) (*Document, error) {
	existingDocument, err := s.GetDocumentByName(name)
	if err != nil {
		return nil, err
	}

	if existingDocument != nil {
		existingDocument.LatestVersion += 1
		if err := s.incrementDocumentVersion(existingDocument); err != nil {
			return nil, err
		}
		return existingDocument, nil
	}

	query := `INSERT INTO documents (Name, CreatedAt, LatestVersion) VALUES (?, ?, ?)`
	res, err := s.Exec(query, name, time.Now(), 1)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.GetDocumentById(int(id))
}

// incrementDocumentVersion - Updates the version of an existing document.
func (s *SQLite) incrementDocumentVersion(document *Document) error {
	query := `UPDATE documents SET LatestVersion = ? WHERE Name = ?`
	_, err := s.Exec(query, document.LatestVersion, document.Name)
	return err
}

// GetDocumentById - Retrieves a document by its ID.
func (s *SQLite) GetDocumentById(id int) (*Document, error) {
	query := `SELECT Id, Name, CreatedAt, LatestVersion FROM documents WHERE Id = ?`
	row := s.QueryRow(query, id)

	var document Document
	if err := row.Scan(&document.Id, &document.Name, &document.CreatedAt, &document.LatestVersion); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	return &document, nil
}

// GetDocumentByName - Retrieves a document by its name.
func (s *SQLite) GetDocumentByName(name string) (*Document, error) {
	query := `SELECT Id, Name, CreatedAt, LatestVersion FROM documents WHERE Name = ?`
	row := s.QueryRow(query, name)

	var document Document
	if err := row.Scan(&document.Id, &document.Name, &document.CreatedAt, &document.LatestVersion); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	return &document, nil
}

// CreateDraft - Creates a new draft for a document.
func (s *SQLite) CreateDraft(draft common.Draft) error {
	tx, err := s.Begin()
	if err != nil {
		return err
	}

	doc, err := s.createDocument(draft.Name)
	if err != nil {
		tx.Rollback()
		return err
	}

	query := `INSERT INTO drafts (DocumentId, Content, VersionNumber, CreatedAt) VALUES (?, ?, ?, ?)`
	if _, err = tx.Exec(query, doc.Id, draft.Content, doc.LatestVersion, time.Now()); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetLatestDrafts - Gets the latest drafts, and if limit is 0, it will return all drafts.
func (s *SQLite) GetLatestDrafts(limit int) ([]Draft, error) {
	var query string
	if limit > 0 {
		// Query to get the latest 'limit' drafts for each DocumentId
		query = `
            SELECT d.Id, d.DocumentId, d.Content, d.VersionNumber, d.CreatedAt
            FROM drafts d
            WHERE (
                SELECT COUNT(*)
                FROM drafts d2
                WHERE d2.DocumentId = d.DocumentId AND d2.Id > d.Id
            ) < ?
            ORDER BY d.VersionNumber DESC`
	} else {
		// Query to get all drafts
		query = `
            SELECT Id, DocumentId, Content, VersionNumber, CreatedAt
            FROM drafts
            ORDER BY VersionNumber DESC`
	}

	rows, err := s.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drafts []Draft
	for rows.Next() {
		var draft Draft
		if err := rows.Scan(&draft.Id, &draft.DocumentId, &draft.Content, &draft.VersionNumber, &draft.CreatedAt); err != nil {
			return nil, err
		}
		drafts = append(drafts, draft)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return drafts, nil
}

// SearchDrafts - Searching drafts will search the content of a draft.
func (s *SQLite) SearchDrafts(query string) ([]Draft, error) {
	sqlQuery := `
        SELECT Id, DocumentId, Content, VersionNumber, CreatedAt
        FROM drafts
        WHERE Content LIKE ?
        ORDER BY Id DESC`

	searchQuery := "%" + query + "%"
	rows, err := s.Query(sqlQuery, searchQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drafts []Draft
	for rows.Next() {
		var draft Draft
		if err := rows.Scan(&draft.Id, &draft.DocumentId, &draft.Content, &draft.VersionNumber, &draft.CreatedAt); err != nil {
			return nil, err
		}
		drafts = append(drafts, draft)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return drafts, nil
}

// GetAllDocumentsLatestVersions - Retrieves a list of document Id's with the latest draft versions.
func (s *SQLite) GetAllDocumentsLatestVersions() ([]Document, error) {
	query := `SELECT Id, Name, LatestVersion, CreatedAt FROM documents ORDER BY Id`
	rows, err := s.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var doc Document
		if err := rows.Scan(&doc.Id, &doc.Name, &doc.LatestVersion, &doc.CreatedAt); err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return documents, nil
}

// AddCommentToDraft - Create a comments to drafts. A successful result will return the comment Id.
func (s *SQLite) AddCommentToDraft(comment Comment) (int64, error) {
	query := `INSERT INTO comments (DraftId, UserId, Text, ParentCommentId, CreatedAt) VALUES (?, ?, ?, ?, ?)`
	result, err := s.Exec(query, comment.DraftId, comment.UserId, comment.Text, comment.ParentCommentId, time.Now())
	if err != nil {
		return 0, err
	}

	commentId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return commentId, nil
}

// GetCommentsAndReactionsByDraftId - Retrieves all of a drafts comments, with the the comment reactions.
func (s *SQLite) GetCommentsAndReactionsByDraftId(draftId int) ([]CommentWithReactions, error) {
	query := `
        SELECT c.Id, c.UserId, c.Text, c.ParentCommentId, c.CreatedAt, 
               r.Id, r.UserId, r.Emoji, r.CreatedAt
        FROM comments c
        LEFT JOIN reactions r ON c.Id = r.CommentId
        WHERE c.DraftId = ?
        ORDER BY c.CreatedAt DESC`

	rows, err := s.Query(query, draftId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	commentsMap := make(map[int]*CommentWithReactions)
	for rows.Next() {
		var commentId int
		var reactionId sql.NullInt64
		var reactionUserId sql.NullInt64
		var reactionEmoji sql.NullString
		var reactionCreatedAt sql.NullTime
		var comment CommentWithReactions
		var reaction common.Reaction

		err := rows.Scan(
			&commentId, &comment.UserId, &comment.Text, &comment.ParentCommentId, &comment.CreatedAt,
			&reactionId, &reactionUserId, &reactionEmoji, &reactionCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Check if the comment is already in the map
		if storedComment, found := commentsMap[commentId]; found {
			if reactionId.Valid { // Check if reactionId is not NULL
				reaction.Id = int(reactionId.Int64) // Convert to int
				reaction.UserId = int(reactionUserId.Int64)
				reaction.Emoji = reactionEmoji.String
				reaction.CreatedAt = reactionCreatedAt.Time
				storedComment.Reactions = append(storedComment.Reactions, reaction)
			}
		} else {
			// If not found, create a new comment with the reaction
			newComment := CommentWithReactions{
				Id:              commentId,
				UserId:          comment.UserId,
				Text:            comment.Text,
				ParentCommentId: comment.ParentCommentId,
				CreatedAt:       comment.CreatedAt,
				Reactions:       []common.Reaction{},
			}
			if reactionId.Valid { // Check if reactionId is not NULL
				reaction.Id = int(reactionId.Int64)
				reaction.UserId = int(reactionUserId.Int64)
				reaction.Emoji = reactionEmoji.String
				reaction.CreatedAt = reactionCreatedAt.Time
				newComment.Reactions = append(newComment.Reactions, reaction)
			}
			commentsMap[commentId] = &newComment
		}
	}

	// Convert the map to a slice
	var commentsWithReactions []CommentWithReactions
	for _, comment := range commentsMap {
		commentsWithReactions = append(commentsWithReactions, *comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return commentsWithReactions, nil
}

// AddReactionToComment - Creats a reaction to a comment.
func (s *SQLite) AddReactionToComment(reaction common.Reaction) error {
	query := `INSERT INTO reactions (CommentId, UserId, Emoji, CreatedAt) VALUES (?, ?, ?, ?)`
	_, err := s.Exec(query, reaction.Id, reaction.UserId, reaction.Emoji, time.Now())
	return err
}
