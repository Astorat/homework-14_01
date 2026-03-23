package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"blog-api/models"
	"blog-api/storage"
)

type CreateCommentRequest struct {
	Text string `json:"text"`
}

type CommentResponse struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	AuthorID  int       `json:"author_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *Handler) CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	idStr := r.PathValue("id")
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Text == "" {
		respondError(w, http.StatusBadRequest, "text is required")
		return
	}

	comment := &models.Comment{
		PostID:    postID,
		AuthorID:  userID,
		Text:      req.Text,
		CreatedAt: time.Now(),
	}

	if err := h.store.CreateComment(comment); err != nil {
		if err == storage.ErrPostNotFound {
			respondError(w, http.StatusNotFound, "post not found")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to create comment")
		}
		return
	}

	h.logger.Log(fmt.Sprintf("user %d created comment %d", userID, comment.ID))

	respondJSON(w, http.StatusCreated, CommentResponse{
		ID:        comment.ID,
		PostID:    comment.PostID,
		AuthorID:  comment.AuthorID,
		Text:      comment.Text,
		CreatedAt: comment.CreatedAt,
	})
}

func (h *Handler) GetCommentsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	comments, err := h.store.GetCommentsByPostID(postID)
	if err != nil {
		if err == storage.ErrPostNotFound {
			respondError(w, http.StatusNotFound, "post not found")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to get comments")
		}
		return
	}

	response := make([]CommentResponse, len(comments))
	for i, c := range comments {
		response[i] = CommentResponse{
			ID:        c.ID,
			PostID:    c.PostID,
			AuthorID:  c.AuthorID,
			Text:      c.Text,
			CreatedAt: c.CreatedAt,
		}
	}

	respondJSON(w, http.StatusOK, response)
}
