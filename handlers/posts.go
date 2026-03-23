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

type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type PostResponse struct {
	ID        int       `json:"id"`
	AuthorID  int       `json:"author_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *Handler) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromRequest(r)
	if err != nil {
		respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" {
		respondError(w, http.StatusBadRequest, "title is required")
		return
	}
	if len([]rune(req.Title)) < 3 {
		respondError(w, http.StatusBadRequest, "title must be at least 3 characters")
		return
	}
	if req.Content == "" {
		respondError(w, http.StatusBadRequest, "content is required")
		return
	}

	post := &models.Post{
		AuthorID:  userID,
		Title:     req.Title,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	if err := h.store.CreatePost(post); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create post")
		return
	}

	h.logger.Log(fmt.Sprintf("user %d created post %d", userID, post.ID))

	respondJSON(w, http.StatusCreated, PostResponse{
		ID:        post.ID,
		AuthorID:  post.AuthorID,
		Title:     post.Title,
		Content:   post.Content,
		CreatedAt: post.CreatedAt,
	})
}

func (h *Handler) GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	posts := h.store.GetAllPosts()
	response := make([]PostResponse, len(posts))
	for i, p := range posts {
		response[i] = PostResponse{
			ID:        p.ID,
			AuthorID:  p.AuthorID,
			Title:     p.Title,
			Content:   p.Content,
			CreatedAt: p.CreatedAt,
		}
	}
	respondJSON(w, http.StatusOK, response)
}

func (h *Handler) GetPostHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid post id")
		return
	}

	post, err := h.store.GetPostByID(id)
	if err != nil {
		if err == storage.ErrPostNotFound {
			respondError(w, http.StatusNotFound, "post not found")
		} else {
			respondError(w, http.StatusInternalServerError, "failed to get post")
		}
		return
	}

	respondJSON(w, http.StatusOK, PostResponse{
		ID:        post.ID,
		AuthorID:  post.AuthorID,
		Title:     post.Title,
		Content:   post.Content,
		CreatedAt: post.CreatedAt,
	})
}
