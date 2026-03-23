package storage

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"blog-api/models"
)

func newTestStorage(t *testing.T) *FileStorage {
	t.Helper()
	dir := t.TempDir()
	return NewFileStorage(dir)
}

// --- CreateUser ---

func TestCreateUser_Success(t *testing.T) {
	store := newTestStorage(t)
	user := &models.User{
		Username:     "alice",
		Email:        "alice@example.com",
		PasswordHash: "hash123",
		CreatedAt:    time.Now(),
	}

	err := store.CreateUser(user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != 1 {
		t.Fatalf("expected ID=1, got %d", user.ID)
	}
}

func TestCreateUser_AutoIncrementID(t *testing.T) {
	store := newTestStorage(t)

	u1 := &models.User{Username: "u1", Email: "u1@mail.com", PasswordHash: "h", CreatedAt: time.Now()}
	u2 := &models.User{Username: "u2", Email: "u2@mail.com", PasswordHash: "h", CreatedAt: time.Now()}
	u3 := &models.User{Username: "u3", Email: "u3@mail.com", PasswordHash: "h", CreatedAt: time.Now()}

	store.CreateUser(u1)
	store.CreateUser(u2)
	store.CreateUser(u3)

	if u1.ID != 1 || u2.ID != 2 || u3.ID != 3 {
		t.Fatalf("expected IDs 1,2,3, got %d,%d,%d", u1.ID, u2.ID, u3.ID)
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	store := newTestStorage(t)

	u1 := &models.User{Username: "alice", Email: "same@mail.com", PasswordHash: "h", CreatedAt: time.Now()}
	u2 := &models.User{Username: "bob", Email: "same@mail.com", PasswordHash: "h", CreatedAt: time.Now()}

	store.CreateUser(u1)
	err := store.CreateUser(u2)

	if err != ErrEmailExists {
		t.Fatalf("expected ErrEmailExists, got %v", err)
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	store := newTestStorage(t)

	u1 := &models.User{Username: "same", Email: "a@mail.com", PasswordHash: "h", CreatedAt: time.Now()}
	u2 := &models.User{Username: "same", Email: "b@mail.com", PasswordHash: "h", CreatedAt: time.Now()}

	store.CreateUser(u1)
	err := store.CreateUser(u2)

	if err != ErrUsernameExists {
		t.Fatalf("expected ErrUsernameExists, got %v", err)
	}
}

// --- GetUserByEmail ---

func TestGetUserByEmail_Found(t *testing.T) {
	store := newTestStorage(t)

	store.CreateUser(&models.User{Username: "alice", Email: "alice@mail.com", PasswordHash: "h", CreatedAt: time.Now()})

	user, err := store.GetUserByEmail("alice@mail.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Username != "alice" {
		t.Fatalf("expected username alice, got %s", user.Username)
	}
	if user.ID != 1 {
		t.Fatalf("expected ID=1, got %d", user.ID)
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	store := newTestStorage(t)

	_, err := store.GetUserByEmail("nobody@mail.com")
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

// --- CreatePost ---

func TestCreatePost_Success(t *testing.T) {
	store := newTestStorage(t)

	post := &models.Post{AuthorID: 1, Title: "Title", Content: "Body", CreatedAt: time.Now()}
	err := store.CreatePost(post)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if post.ID != 1 {
		t.Fatalf("expected ID=1, got %d", post.ID)
	}
}

func TestCreatePost_AutoIncrementID(t *testing.T) {
	store := newTestStorage(t)

	p1 := &models.Post{AuthorID: 1, Title: "T1", Content: "C1", CreatedAt: time.Now()}
	p2 := &models.Post{AuthorID: 1, Title: "T2", Content: "C2", CreatedAt: time.Now()}

	store.CreatePost(p1)
	store.CreatePost(p2)

	if p1.ID != 1 || p2.ID != 2 {
		t.Fatalf("expected IDs 1,2, got %d,%d", p1.ID, p2.ID)
	}
}

// --- GetAllPosts ---

func TestGetAllPosts_Empty(t *testing.T) {
	store := newTestStorage(t)

	posts := store.GetAllPosts()
	if len(posts) != 0 {
		t.Fatalf("expected 0 posts, got %d", len(posts))
	}
}

func TestGetAllPosts_WithPosts(t *testing.T) {
	store := newTestStorage(t)

	store.CreatePost(&models.Post{AuthorID: 1, Title: "T1", Content: "C1", CreatedAt: time.Now()})
	store.CreatePost(&models.Post{AuthorID: 2, Title: "T2", Content: "C2", CreatedAt: time.Now()})

	posts := store.GetAllPosts()
	if len(posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(posts))
	}
}

func TestGetAllPosts_ReturnsCopy(t *testing.T) {
	store := newTestStorage(t)
	store.CreatePost(&models.Post{AuthorID: 1, Title: "T1", Content: "C1", CreatedAt: time.Now()})

	posts := store.GetAllPosts()
	posts[0].Title = "MODIFIED"

	original := store.GetAllPosts()
	if original[0].Title == "MODIFIED" {
		t.Fatal("GetAllPosts should return a copy, not a reference to internal data")
	}
}

// --- GetPostByID ---

func TestGetPostByID_Found(t *testing.T) {
	store := newTestStorage(t)
	store.CreatePost(&models.Post{AuthorID: 1, Title: "MyPost", Content: "Body", CreatedAt: time.Now()})

	post, err := store.GetPostByID(1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if post.Title != "MyPost" {
		t.Fatalf("expected title MyPost, got %s", post.Title)
	}
}

func TestGetPostByID_NotFound(t *testing.T) {
	store := newTestStorage(t)

	_, err := store.GetPostByID(999)
	if err != ErrPostNotFound {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
}

// --- CreateComment ---

func TestCreateComment_Success(t *testing.T) {
	store := newTestStorage(t)
	store.CreatePost(&models.Post{AuthorID: 1, Title: "T", Content: "C", CreatedAt: time.Now()})

	comment := &models.Comment{PostID: 1, AuthorID: 1, Text: "Nice!", CreatedAt: time.Now()}
	err := store.CreateComment(comment)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if comment.ID != 1 {
		t.Fatalf("expected ID=1, got %d", comment.ID)
	}
}

func TestCreateComment_PostNotFound(t *testing.T) {
	store := newTestStorage(t)

	comment := &models.Comment{PostID: 999, AuthorID: 1, Text: "Nice!", CreatedAt: time.Now()}
	err := store.CreateComment(comment)

	if err != ErrPostNotFound {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
}

func TestCreateComment_AutoIncrementID(t *testing.T) {
	store := newTestStorage(t)
	store.CreatePost(&models.Post{AuthorID: 1, Title: "T", Content: "C", CreatedAt: time.Now()})

	c1 := &models.Comment{PostID: 1, AuthorID: 1, Text: "First", CreatedAt: time.Now()}
	c2 := &models.Comment{PostID: 1, AuthorID: 2, Text: "Second", CreatedAt: time.Now()}

	store.CreateComment(c1)
	store.CreateComment(c2)

	if c1.ID != 1 || c2.ID != 2 {
		t.Fatalf("expected IDs 1,2, got %d,%d", c1.ID, c2.ID)
	}
}

// --- GetCommentsByPostID ---

func TestGetCommentsByPostID_WithComments(t *testing.T) {
	store := newTestStorage(t)
	store.CreatePost(&models.Post{AuthorID: 1, Title: "T", Content: "C", CreatedAt: time.Now()})
	store.CreateComment(&models.Comment{PostID: 1, AuthorID: 1, Text: "A", CreatedAt: time.Now()})
	store.CreateComment(&models.Comment{PostID: 1, AuthorID: 2, Text: "B", CreatedAt: time.Now()})

	comments, err := store.GetCommentsByPostID(1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments, got %d", len(comments))
	}
}

func TestGetCommentsByPostID_Empty(t *testing.T) {
	store := newTestStorage(t)
	store.CreatePost(&models.Post{AuthorID: 1, Title: "T", Content: "C", CreatedAt: time.Now()})

	comments, err := store.GetCommentsByPostID(1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(comments) != 0 {
		t.Fatalf("expected 0 comments, got %d", len(comments))
	}
}

func TestGetCommentsByPostID_PostNotFound(t *testing.T) {
	store := newTestStorage(t)

	_, err := store.GetCommentsByPostID(999)
	if err != ErrPostNotFound {
		t.Fatalf("expected ErrPostNotFound, got %v", err)
	}
}

func TestGetCommentsByPostID_FiltersCorrectly(t *testing.T) {
	store := newTestStorage(t)
	store.CreatePost(&models.Post{AuthorID: 1, Title: "P1", Content: "C", CreatedAt: time.Now()})
	store.CreatePost(&models.Post{AuthorID: 1, Title: "P2", Content: "C", CreatedAt: time.Now()})

	store.CreateComment(&models.Comment{PostID: 1, AuthorID: 1, Text: "For post 1", CreatedAt: time.Now()})
	store.CreateComment(&models.Comment{PostID: 2, AuthorID: 1, Text: "For post 2", CreatedAt: time.Now()})
	store.CreateComment(&models.Comment{PostID: 1, AuthorID: 2, Text: "Also for post 1", CreatedAt: time.Now()})

	comments, _ := store.GetCommentsByPostID(1)
	if len(comments) != 2 {
		t.Fatalf("expected 2 comments for post 1, got %d", len(comments))
	}

	comments2, _ := store.GetCommentsByPostID(2)
	if len(comments2) != 1 {
		t.Fatalf("expected 1 comment for post 2, got %d", len(comments2))
	}
}

// --- Persistence ---

func TestPersistence_DataSurvivesReload(t *testing.T) {
	dir := t.TempDir()

	store1 := NewFileStorage(dir)
	store1.CreateUser(&models.User{Username: "alice", Email: "a@mail.com", PasswordHash: "h", CreatedAt: time.Now()})
	store1.CreatePost(&models.Post{AuthorID: 1, Title: "T", Content: "C", CreatedAt: time.Now()})
	store1.CreateComment(&models.Comment{PostID: 1, AuthorID: 1, Text: "Hi", CreatedAt: time.Now()})

	store2 := NewFileStorage(dir)

	user, err := store2.GetUserByEmail("a@mail.com")
	if err != nil {
		t.Fatalf("user not persisted: %v", err)
	}
	if user.Username != "alice" {
		t.Fatalf("expected alice, got %s", user.Username)
	}

	posts := store2.GetAllPosts()
	if len(posts) != 1 {
		t.Fatalf("expected 1 post after reload, got %d", len(posts))
	}

	comments, _ := store2.GetCommentsByPostID(1)
	if len(comments) != 1 {
		t.Fatalf("expected 1 comment after reload, got %d", len(comments))
	}
}

func TestPersistence_FilesCreated(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStorage(dir)

	store.CreateUser(&models.User{Username: "u", Email: "u@m.com", PasswordHash: "h", CreatedAt: time.Now()})

	if _, err := os.Stat(filepath.Join(dir, "users.json")); os.IsNotExist(err) {
		t.Fatal("users.json was not created")
	}
}

func TestLoadSlice_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "users.json"), []byte("not json"), 0644)

	store := NewFileStorage(dir)
	users := store.GetAllPosts()
	if len(users) != 0 {
		t.Fatal("expected empty slice for invalid JSON")
	}
}

func TestLoadSlice_MissingFile(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStorage(dir)

	posts := store.GetAllPosts()
	if posts == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(posts) != 0 {
		t.Fatalf("expected 0 posts, got %d", len(posts))
	}
}

// --- Concurrency ---

func TestConcurrentAccess(t *testing.T) {
	store := newTestStorage(t)
	store.CreatePost(&models.Post{AuthorID: 1, Title: "T", Content: "C", CreatedAt: time.Now()})

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			store.CreateComment(&models.Comment{
				PostID:    1,
				AuthorID:  i,
				Text:      "concurrent comment",
				CreatedAt: time.Now(),
			})
		}(i)
	}
	wg.Wait()

	comments, _ := store.GetCommentsByPostID(1)
	if len(comments) != 20 {
		t.Fatalf("expected 20 comments, got %d", len(comments))
	}
}
