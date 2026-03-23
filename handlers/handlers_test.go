package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"blog-api/auth"
	"blog-api/logger"
	"blog-api/storage"
)

type testEnv struct {
	handler *Handler
	server  *httptest.Server
	auth    *auth.TokenAuth
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()
	dir := t.TempDir()
	store := storage.NewFileStorage(filepath.Join(dir, "data"))
	tokenAuth := auth.NewTokenAuth("test-secret")

	logFile := filepath.Join(dir, "test.log")
	eventLogger := logger.NewEventLogger(logFile)
	eventLogger.Start()

	h := NewHandler(store, tokenAuth, eventLogger)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.HealthHandler)
	mux.HandleFunc("POST /register", h.RegisterHandler)
	mux.HandleFunc("POST /login", h.LoginHandler)
	mux.HandleFunc("POST /posts", h.CreatePostHandler)
	mux.HandleFunc("GET /posts", h.GetPostsHandler)
	mux.HandleFunc("GET /posts/{id}", h.GetPostHandler)
	mux.HandleFunc("POST /posts/{id}/comments", h.CreateCommentHandler)
	mux.HandleFunc("GET /posts/{id}/comments", h.GetCommentsHandler)

	server := httptest.NewServer(mux)

	t.Cleanup(func() {
		server.Close()
		eventLogger.Stop()
	})

	return &testEnv{handler: h, server: server, auth: tokenAuth}
}

func doRequest(t *testing.T, method, url string, body interface{}, token string) (*http.Response, map[string]interface{}) {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	return resp, result
}

func doRequestArray(t *testing.T, method, url string, token string) (*http.Response, []map[string]interface{}) {
	t.Helper()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	var result []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	return resp, result
}

func registerUser(t *testing.T, env *testEnv, username, email, password string) string {
	t.Helper()
	body := map[string]string{"username": username, "email": email, "password": password}
	resp, result := doRequest(t, "POST", env.server.URL+"/register", body, "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register failed: %d, %v", resp.StatusCode, result)
	}
	return result["token"].(string)
}

// ==================== Health ====================

func TestHealthHandler(t *testing.T) {
	env := setupTestEnv(t)
	resp, result := doRequest(t, "GET", env.server.URL+"/health", nil, "")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if result["status"] != "ok" {
		t.Fatalf("expected status ok, got %v", result["status"])
	}
}

func TestHealthHandler_ContentType(t *testing.T) {
	env := setupTestEnv(t)
	resp, _ := doRequest(t, "GET", env.server.URL+"/health", nil, "")

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", ct)
	}
}

// ==================== Register ====================

func TestRegister_Success(t *testing.T) {
	env := setupTestEnv(t)
	body := map[string]string{"username": "alice", "email": "alice@mail.com", "password": "secret123"}
	resp, result := doRequest(t, "POST", env.server.URL+"/register", body, "")

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	if result["token"] == nil || result["token"] == "" {
		t.Fatal("expected token in response")
	}
	user := result["user"].(map[string]interface{})
	if user["username"] != "alice" {
		t.Fatalf("expected username alice, got %v", user["username"])
	}
	if user["email"] != "alice@mail.com" {
		t.Fatalf("expected email alice@mail.com, got %v", user["email"])
	}
	if user["id"].(float64) != 1 {
		t.Fatalf("expected id 1, got %v", user["id"])
	}
}

func TestRegister_MissingUsername(t *testing.T) {
	env := setupTestEnv(t)
	body := map[string]string{"email": "a@m.com", "password": "secret123"}
	resp, result := doRequest(t, "POST", env.server.URL+"/register", body, "")

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if result["error"] != "username is required" {
		t.Fatalf("expected 'username is required', got %v", result["error"])
	}
}

func TestRegister_MissingEmail(t *testing.T) {
	env := setupTestEnv(t)
	body := map[string]string{"username": "alice", "password": "secret123"}
	resp, result := doRequest(t, "POST", env.server.URL+"/register", body, "")

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if result["error"] != "email is required" {
		t.Fatalf("expected 'email is required', got %v", result["error"])
	}
}

func TestRegister_MissingPassword(t *testing.T) {
	env := setupTestEnv(t)
	body := map[string]string{"username": "alice", "email": "a@m.com"}
	resp, result := doRequest(t, "POST", env.server.URL+"/register", body, "")

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if result["error"] != "password is required" {
		t.Fatalf("expected 'password is required', got %v", result["error"])
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	env := setupTestEnv(t)
	body := map[string]string{"username": "alice", "email": "a@m.com", "password": "123"}
	resp, result := doRequest(t, "POST", env.server.URL+"/register", body, "")

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if result["error"] != "password must be at least 6 characters" {
		t.Fatalf("unexpected error: %v", result["error"])
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	env := setupTestEnv(t)
	registerUser(t, env, "alice", "same@mail.com", "secret123")

	body := map[string]string{"username": "bob", "email": "same@mail.com", "password": "secret123"}
	resp, result := doRequest(t, "POST", env.server.URL+"/register", body, "")

	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	if result["error"] != "email already exists" {
		t.Fatalf("expected 'email already exists', got %v", result["error"])
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	env := setupTestEnv(t)
	registerUser(t, env, "same", "a@mail.com", "secret123")

	body := map[string]string{"username": "same", "email": "b@mail.com", "password": "secret123"}
	resp, result := doRequest(t, "POST", env.server.URL+"/register", body, "")

	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	if result["error"] != "username already exists" {
		t.Fatalf("expected 'username already exists', got %v", result["error"])
	}
}

func TestRegister_InvalidJSON(t *testing.T) {
	env := setupTestEnv(t)

	req, _ := http.NewRequest("POST", env.server.URL+"/register", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// ==================== Login ====================

func TestLogin_Success(t *testing.T) {
	env := setupTestEnv(t)
	registerUser(t, env, "alice", "alice@mail.com", "secret123")

	body := map[string]string{"email": "alice@mail.com", "password": "secret123"}
	resp, result := doRequest(t, "POST", env.server.URL+"/login", body, "")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if result["token"] == nil || result["token"] == "" {
		t.Fatal("expected token in response")
	}
	user := result["user"].(map[string]interface{})
	if user["username"] != "alice" {
		t.Fatalf("expected username alice, got %v", user["username"])
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	env := setupTestEnv(t)
	registerUser(t, env, "alice", "alice@mail.com", "secret123")

	body := map[string]string{"email": "alice@mail.com", "password": "wrongpass"}
	resp, result := doRequest(t, "POST", env.server.URL+"/login", body, "")

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	if result["error"] != "invalid email or password" {
		t.Fatalf("expected 'invalid email or password', got %v", result["error"])
	}
}

func TestLogin_WrongEmail(t *testing.T) {
	env := setupTestEnv(t)

	body := map[string]string{"email": "nobody@mail.com", "password": "secret123"}
	resp, result := doRequest(t, "POST", env.server.URL+"/login", body, "")

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	if result["error"] != "invalid email or password" {
		t.Fatalf("expected 'invalid email or password', got %v", result["error"])
	}
}

func TestLogin_MissingEmail(t *testing.T) {
	env := setupTestEnv(t)

	body := map[string]string{"password": "secret123"}
	resp, result := doRequest(t, "POST", env.server.URL+"/login", body, "")

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if result["error"] != "email is required" {
		t.Fatalf("expected 'email is required', got %v", result["error"])
	}
}

func TestLogin_MissingPassword(t *testing.T) {
	env := setupTestEnv(t)

	body := map[string]string{"email": "a@m.com"}
	resp, result := doRequest(t, "POST", env.server.URL+"/login", body, "")

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if result["error"] != "password is required" {
		t.Fatalf("expected 'password is required', got %v", result["error"])
	}
}

func TestLogin_InvalidJSON(t *testing.T) {
	env := setupTestEnv(t)

	req, _ := http.NewRequest("POST", env.server.URL+"/login", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// ==================== Create Post ====================

func TestCreatePost_Success(t *testing.T) {
	env := setupTestEnv(t)
	token := registerUser(t, env, "alice", "alice@mail.com", "secret123")

	body := map[string]string{"title": "My Post", "content": "Post content"}
	resp, result := doRequest(t, "POST", env.server.URL+"/posts", body, token)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	if result["id"].(float64) != 1 {
		t.Fatalf("expected id 1, got %v", result["id"])
	}
	if result["title"] != "My Post" {
		t.Fatalf("expected title 'My Post', got %v", result["title"])
	}
	if result["author_id"].(float64) != 1 {
		t.Fatalf("expected author_id 1, got %v", result["author_id"])
	}
}

func TestCreatePost_NoAuth(t *testing.T) {
	env := setupTestEnv(t)

	body := map[string]string{"title": "My Post", "content": "Content"}
	resp, result := doRequest(t, "POST", env.server.URL+"/posts", body, "")

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	if result["error"] != "authorization header required" {
		t.Fatalf("expected 'authorization header required', got %v", result["error"])
	}
}

func TestCreatePost_InvalidToken(t *testing.T) {
	env := setupTestEnv(t)

	body := map[string]string{"title": "My Post", "content": "Content"}
	resp, _ := doRequest(t, "POST", env.server.URL+"/posts", body, "invalid-token")

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestCreatePost_MissingTitle(t *testing.T) {
	env := setupTestEnv(t)
	token := registerUser(t, env, "alice", "alice@mail.com", "secret123")

	body := map[string]string{"content": "Content"}
	resp, result := doRequest(t, "POST", env.server.URL+"/posts", body, token)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if result["error"] != "title is required" {
		t.Fatalf("expected 'title is required', got %v", result["error"])
	}
}

func TestCreatePost_MissingContent(t *testing.T) {
	env := setupTestEnv(t)
	token := registerUser(t, env, "alice", "alice@mail.com", "secret123")

	body := map[string]string{"title": "Title"}
	resp, result := doRequest(t, "POST", env.server.URL+"/posts", body, token)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if result["error"] != "content is required" {
		t.Fatalf("expected 'content is required', got %v", result["error"])
	}
}

func TestCreatePost_InvalidJSON(t *testing.T) {
	env := setupTestEnv(t)
	token := registerUser(t, env, "alice", "alice@mail.com", "secret123")

	req, _ := http.NewRequest("POST", env.server.URL+"/posts", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

// ==================== Get Posts ====================

func TestGetPosts_Empty(t *testing.T) {
	env := setupTestEnv(t)

	resp, result := doRequestArray(t, "GET", env.server.URL+"/posts", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 posts, got %d", len(result))
	}
}

func TestGetPosts_WithPosts(t *testing.T) {
	env := setupTestEnv(t)
	token := registerUser(t, env, "alice", "alice@mail.com", "secret123")

	doRequest(t, "POST", env.server.URL+"/posts", map[string]string{"title": "P1", "content": "C1"}, token)
	doRequest(t, "POST", env.server.URL+"/posts", map[string]string{"title": "P2", "content": "C2"}, token)

	resp, result := doRequestArray(t, "GET", env.server.URL+"/posts", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(result))
	}
}

func TestGetPosts_NoAuthRequired(t *testing.T) {
	env := setupTestEnv(t)

	resp, _ := doRequestArray(t, "GET", env.server.URL+"/posts", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /posts should not require auth, got %d", resp.StatusCode)
	}
}

// ==================== Get Post by ID ====================

func TestGetPost_Success(t *testing.T) {
	env := setupTestEnv(t)
	token := registerUser(t, env, "alice", "alice@mail.com", "secret123")
	doRequest(t, "POST", env.server.URL+"/posts", map[string]string{"title": "MyPost", "content": "Body"}, token)

	resp, result := doRequest(t, "GET", env.server.URL+"/posts/1", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if result["title"] != "MyPost" {
		t.Fatalf("expected title MyPost, got %v", result["title"])
	}
}

func TestGetPost_NotFound(t *testing.T) {
	env := setupTestEnv(t)

	resp, result := doRequest(t, "GET", env.server.URL+"/posts/999", nil, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	if result["error"] != "post not found" {
		t.Fatalf("expected 'post not found', got %v", result["error"])
	}
}

func TestGetPost_InvalidID(t *testing.T) {
	env := setupTestEnv(t)

	resp, result := doRequest(t, "GET", env.server.URL+"/posts/abc", nil, "")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if result["error"] != "invalid post id" {
		t.Fatalf("expected 'invalid post id', got %v", result["error"])
	}
}

func TestGetPost_NoAuthRequired(t *testing.T) {
	env := setupTestEnv(t)
	token := registerUser(t, env, "alice", "alice@mail.com", "secret123")
	doRequest(t, "POST", env.server.URL+"/posts", map[string]string{"title": "P", "content": "C"}, token)

	resp, _ := doRequest(t, "GET", env.server.URL+"/posts/1", nil, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /posts/{id} should not require auth, got %d", resp.StatusCode)
	}
}

// ==================== Create Comment ====================

func TestCreateComment_Success(t *testing.T) {
	env := setupTestEnv(t)
	token := registerUser(t, env, "alice", "alice@mail.com", "secret123")
	doRequest(t, "POST", env.server.URL+"/posts", map[string]string{"title": "P", "content": "C"}, token)

	body := map[string]string{"text": "Great post!"}
	resp, result := doRequest(t, "POST", env.server.URL+"/posts/1/comments", body, token)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	if result["id"].(float64) != 1 {
		t.Fatalf("expected id 1, got %v", result["id"])
	}
	if result["text"] != "Great post!" {
		t.Fatalf("expected text 'Great post!', got %v", result["text"])
	}
	if result["post_id"].(float64) != 1 {
		t.Fatalf("expected post_id 1, got %v", result["post_id"])
	}
}

func TestCreateComment_NoAuth(t *testing.T) {
	env := setupTestEnv(t)

	body := map[string]string{"text": "comment"}
	resp, result := doRequest(t, "POST", env.server.URL+"/posts/1/comments", body, "")

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	if result["error"] != "authorization header required" {
		t.Fatalf("expected 'authorization header required', got %v", result["error"])
	}
}

func TestCreateComment_PostNotFound(t *testing.T) {
	env := setupTestEnv(t)
	token := registerUser(t, env, "alice", "alice@mail.com", "secret123")

	body := map[string]string{"text": "comment"}
	resp, result := doRequest(t, "POST", env.server.URL+"/posts/999/comments", body, token)

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	if result["error"] != "post not found" {
		t.Fatalf("expected 'post not found', got %v", result["error"])
	}
}

func TestCreateComment_MissingText(t *testing.T) {
	env := setupTestEnv(t)
	token := registerUser(t, env, "alice", "alice@mail.com", "secret123")
	doRequest(t, "POST", env.server.URL+"/posts", map[string]string{"title": "P", "content": "C"}, token)

	body := map[string]string{}
	resp, result := doRequest(t, "POST", env.server.URL+"/posts/1/comments", body, token)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if result["error"] != "text is required" {
		t.Fatalf("expected 'text is required', got %v", result["error"])
	}
}

func TestCreateComment_InvalidPostID(t *testing.T) {
	env := setupTestEnv(t)
	token := registerUser(t, env, "alice", "alice@mail.com", "secret123")

	body := map[string]string{"text": "comment"}
	resp, result := doRequest(t, "POST", env.server.URL+"/posts/abc/comments", body, token)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if result["error"] != "invalid post id" {
		t.Fatalf("expected 'invalid post id', got %v", result["error"])
	}
}

func TestCreateComment_InvalidJSON(t *testing.T) {
	env := setupTestEnv(t)
	token := registerUser(t, env, "alice", "alice@mail.com", "secret123")
	doRequest(t, "POST", env.server.URL+"/posts", map[string]string{"title": "P", "content": "C"}, token)

	req, _ := http.NewRequest("POST", env.server.URL+"/posts/1/comments", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateComment_InvalidToken(t *testing.T) {
	env := setupTestEnv(t)

	body := map[string]string{"text": "comment"}
	resp, _ := doRequest(t, "POST", env.server.URL+"/posts/1/comments", body, "bad-token")

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

// ==================== Get Comments ====================

func TestGetComments_WithComments(t *testing.T) {
	env := setupTestEnv(t)
	token := registerUser(t, env, "alice", "alice@mail.com", "secret123")
	doRequest(t, "POST", env.server.URL+"/posts", map[string]string{"title": "P", "content": "C"}, token)
	doRequest(t, "POST", env.server.URL+"/posts/1/comments", map[string]string{"text": "C1"}, token)
	doRequest(t, "POST", env.server.URL+"/posts/1/comments", map[string]string{"text": "C2"}, token)

	resp, result := doRequestArray(t, "GET", env.server.URL+"/posts/1/comments", "")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(result))
	}
}

func TestGetComments_Empty(t *testing.T) {
	env := setupTestEnv(t)
	token := registerUser(t, env, "alice", "alice@mail.com", "secret123")
	doRequest(t, "POST", env.server.URL+"/posts", map[string]string{"title": "P", "content": "C"}, token)

	resp, result := doRequestArray(t, "GET", env.server.URL+"/posts/1/comments", "")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 comments, got %d", len(result))
	}
}

func TestGetComments_PostNotFound(t *testing.T) {
	env := setupTestEnv(t)

	resp, result := doRequest(t, "GET", env.server.URL+"/posts/999/comments", nil, "")

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	if result["error"] != "post not found" {
		t.Fatalf("expected 'post not found', got %v", result["error"])
	}
}

func TestGetComments_InvalidPostID(t *testing.T) {
	env := setupTestEnv(t)

	resp, result := doRequest(t, "GET", env.server.URL+"/posts/abc/comments", nil, "")

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if result["error"] != "invalid post id" {
		t.Fatalf("expected 'invalid post id', got %v", result["error"])
	}
}

func TestGetComments_NoAuthRequired(t *testing.T) {
	env := setupTestEnv(t)
	token := registerUser(t, env, "alice", "alice@mail.com", "secret123")
	doRequest(t, "POST", env.server.URL+"/posts", map[string]string{"title": "P", "content": "C"}, token)

	resp, _ := doRequestArray(t, "GET", env.server.URL+"/posts/1/comments", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /posts/{id}/comments should not require auth, got %d", resp.StatusCode)
	}
}

// ==================== Auth Header Parsing ====================

func TestInvalidAuthHeaderFormat(t *testing.T) {
	env := setupTestEnv(t)

	req, _ := http.NewRequest("POST", env.server.URL+"/posts", bytes.NewReader([]byte(`{"title":"T","content":"C"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token abc123")
	resp, _ := http.DefaultClient.Do(req)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for non-Bearer auth, got %d", resp.StatusCode)
	}
}

// ==================== Full Flow ====================

func TestFullFlow_RegisterLoginPostComment(t *testing.T) {
	env := setupTestEnv(t)

	// Register
	regBody := map[string]string{"username": "alice", "email": "alice@mail.com", "password": "secret123"}
	regResp, regResult := doRequest(t, "POST", env.server.URL+"/register", regBody, "")
	if regResp.StatusCode != http.StatusCreated {
		t.Fatalf("register failed: %d", regResp.StatusCode)
	}

	// Login
	loginBody := map[string]string{"email": "alice@mail.com", "password": "secret123"}
	loginResp, loginResult := doRequest(t, "POST", env.server.URL+"/login", loginBody, "")
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login failed: %d", loginResp.StatusCode)
	}
	token := loginResult["token"].(string)

	// Tokens from register and login should both be valid
	regToken := regResult["token"].(string)
	if regToken == "" || token == "" {
		t.Fatal("both register and login should return valid tokens")
	}

	// Create post
	postBody := map[string]string{"title": "Hello World", "content": "My first post"}
	postResp, postResult := doRequest(t, "POST", env.server.URL+"/posts", postBody, token)
	if postResp.StatusCode != http.StatusCreated {
		t.Fatalf("create post failed: %d", postResp.StatusCode)
	}
	postID := postResult["id"].(float64)

	// Get all posts
	_, posts := doRequestArray(t, "GET", env.server.URL+"/posts", "")
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}

	// Get post by ID
	getResp, getResult := doRequest(t, "GET", env.server.URL+"/posts/1", nil, "")
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("get post failed: %d", getResp.StatusCode)
	}
	if getResult["id"].(float64) != postID {
		t.Fatal("post ID mismatch")
	}

	// Create comment
	commentBody := map[string]string{"text": "Great post!"}
	commentResp, commentResult := doRequest(t, "POST", env.server.URL+"/posts/1/comments", commentBody, token)
	if commentResp.StatusCode != http.StatusCreated {
		t.Fatalf("create comment failed: %d", commentResp.StatusCode)
	}
	if commentResult["post_id"].(float64) != 1 {
		t.Fatal("comment post_id mismatch")
	}

	// Get comments
	_, comments := doRequestArray(t, "GET", env.server.URL+"/posts/1/comments", "")
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(comments))
	}
	if comments[0]["text"] != "Great post!" {
		t.Fatalf("expected 'Great post!', got %v", comments[0]["text"])
	}
}

func TestMultipleUsers_PostsAndComments(t *testing.T) {
	env := setupTestEnv(t)

	token1 := registerUser(t, env, "alice", "alice@m.com", "secret123")
	token2 := registerUser(t, env, "bob", "bob@m.com", "secret456")

	// Alice creates a post
	doRequest(t, "POST", env.server.URL+"/posts", map[string]string{"title": "Alice's Post", "content": "By Alice"}, token1)

	// Bob comments on Alice's post
	commentResp, commentResult := doRequest(t, "POST", env.server.URL+"/posts/1/comments",
		map[string]string{"text": "Hi Alice!"}, token2)
	if commentResp.StatusCode != http.StatusCreated {
		t.Fatalf("Bob should be able to comment, got %d", commentResp.StatusCode)
	}
	if commentResult["author_id"].(float64) != 2 {
		t.Fatalf("comment author_id should be 2 (Bob), got %v", commentResult["author_id"])
	}

	// Alice comments on her own post
	doRequest(t, "POST", env.server.URL+"/posts/1/comments",
		map[string]string{"text": "Thanks Bob!"}, token1)

	_, comments := doRequestArray(t, "GET", env.server.URL+"/posts/1/comments", "")
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}
}
