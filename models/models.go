package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

type Post struct {
	ID        int       `json:"id"`
	AuthorID  int       `json:"author_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
	ID        int       `json:"id"`
	PostID    int       `json:"post_id"`
	AuthorID  int       `json:"author_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}
