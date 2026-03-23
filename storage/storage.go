package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"blog-api/models"
)

var (
	ErrEmailExists    = errors.New("email already exists")
	ErrUsernameExists = errors.New("username already exists")
	ErrUserNotFound   = errors.New("user not found")
	ErrPostNotFound   = errors.New("post not found")
)

type FileStorage struct {
	dataDir  string
	mu       sync.Mutex
	users    []models.User
	posts    []models.Post
	comments []models.Comment
}

func NewFileStorage(dataDir string) *FileStorage {
	os.MkdirAll(dataDir, 0755)
	fs := &FileStorage{dataDir: dataDir}
	fs.load()
	return fs
}

func (fs *FileStorage) load() {
	fs.users = loadSlice[models.User](filepath.Join(fs.dataDir, "users.json"))
	fs.posts = loadSlice[models.Post](filepath.Join(fs.dataDir, "posts.json"))
	fs.comments = loadSlice[models.Comment](filepath.Join(fs.dataDir, "comments.json"))
}

func loadSlice[T any](path string) []T {
	data, err := os.ReadFile(path)
	if err != nil {
		return []T{}
	}
	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return []T{}
	}
	return items
}

func saveSlice[T any](path string, items []T) error {
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (fs *FileStorage) CreateUser(user *models.User) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for _, u := range fs.users {
		if u.Email == user.Email {
			return ErrEmailExists
		}
		if u.Username == user.Username {
			return ErrUsernameExists
		}
	}

	maxID := 0
	for _, u := range fs.users {
		if u.ID > maxID {
			maxID = u.ID
		}
	}
	user.ID = maxID + 1
	fs.users = append(fs.users, *user)
	return saveSlice(filepath.Join(fs.dataDir, "users.json"), fs.users)
}

func (fs *FileStorage) GetUserByEmail(email string) (*models.User, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for i := range fs.users {
		if fs.users[i].Email == email {
			u := fs.users[i]
			return &u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (fs *FileStorage) CreatePost(post *models.Post) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	maxID := 0
	for _, p := range fs.posts {
		if p.ID > maxID {
			maxID = p.ID
		}
	}
	post.ID = maxID + 1
	fs.posts = append(fs.posts, *post)
	return saveSlice(filepath.Join(fs.dataDir, "posts.json"), fs.posts)
}

func (fs *FileStorage) GetAllPosts() []models.Post {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	result := make([]models.Post, len(fs.posts))
	copy(result, fs.posts)
	return result
}

func (fs *FileStorage) GetPostByID(id int) (*models.Post, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for i := range fs.posts {
		if fs.posts[i].ID == id {
			p := fs.posts[i]
			return &p, nil
		}
	}
	return nil, ErrPostNotFound
}

func (fs *FileStorage) CreateComment(comment *models.Comment) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	postExists := false
	for _, p := range fs.posts {
		if p.ID == comment.PostID {
			postExists = true
			break
		}
	}
	if !postExists {
		return ErrPostNotFound
	}

	maxID := 0
	for _, c := range fs.comments {
		if c.ID > maxID {
			maxID = c.ID
		}
	}
	comment.ID = maxID + 1
	fs.comments = append(fs.comments, *comment)
	return saveSlice(filepath.Join(fs.dataDir, "comments.json"), fs.comments)
}

func (fs *FileStorage) GetCommentsByPostID(postID int) ([]models.Comment, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	postExists := false
	for _, p := range fs.posts {
		if p.ID == postID {
			postExists = true
			break
		}
	}
	if !postExists {
		return nil, ErrPostNotFound
	}

	result := []models.Comment{}
	for _, c := range fs.comments {
		if c.PostID == postID {
			result = append(result, c)
		}
	}
	return result, nil
}
