# Document and Comment Management API

## Overview
The Document and Comment Management API is a RESTful service designed for managing documents and their drafts, along with facilitating user interactions through comments and reactions. This system allows for storing multiple versions of a document, commenting on specific drafts, and reacting to comments, making it an ideal solution for collaborative document editing and review.

## Features
Document Management: Create and store multiple drafts of a document.
Commenting System: Users can leave comments on specific drafts, enabling feedback and discussions.
Nested Comments: Support for commenting on existing comments, allowing threaded discussions.
Reactions: Users can react to comments with emojis, enhancing interaction.
Search Functionality: Search within draft contents for specific text strings.
RESTful API: Easy to use API endpoints for managing documents, drafts, comments, and reactions.

## Installation
```bash
git clone https://github.com/MudDev/document-comment-api.git
cd document-comment-api
go build .\cmd\
```

## Usage
To start the API server, simply run the binary.


## API Endpoints
POST /api/drafts - Add a new draft.
GET /api/drafts - Get the most recent drafts.
GET /api/drafts/search - Search within drafts.
POST /api/comments - Add a comment to a draft.
GET /api/drafts/{draftId}/comments - Get comments for a draft.
POST /api/comment/{commentId}/reaction - Add a reaction to a comment.
GET /api/documents/latest - Get the most recent version of all documents.

## Postman
A postman collection is included, use the import to utilize this collection
|- Custodia Document Drafts.postman_collection.json

## Testing
```bash
go test ./...
```

