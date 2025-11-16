package models

import "time"

type User struct {
	ID             string    `json:"id"`
	Username       string    `json:"username"`
	PasswordHash   string    `json:"-"`
	BusinessName  string    `json:"business_name"`
	Specialization string    `json:"specialization"`
	CreatedAt      time.Time `json:"created_at"`
}

type File struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Filename   string    `json:"filename"`
	FilePath   string    `json:"file_path"`
	FileType   string    `json:"file_type"`
	FileSize   int64     `json:"file_size"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Chat struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Message struct {
	ID        string    `json:"id"`
	ChatID    string    `json:"chat_id"`
	UserID    string    `json:"user_id"`
	Message   string    `json:"message"`
	Response  string    `json:"response"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Username       string `json:"username" binding:"required"`
	Password       string `json:"password" binding:"required,min=6"`
	BusinessName   string `json:"business_name" binding:"required"`
	Specialization string `json:"specialization" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChatRequest struct {
	Message  string `json:"message" binding:"required"`
	Category string `json:"category"`
	ChatID   string `json:"chat_id"`
}

type CreateChatRequest struct {
	Title string `json:"title"`
}

