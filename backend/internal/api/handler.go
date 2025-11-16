package api

import (
	"alfa-hack-backend/internal/ai"
	"alfa-hack-backend/internal/models"
	"database/sql"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// Register - регистрация нового пользователя
func (h *Handler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверка существования пользователя
	var existingID string
	err := h.db.QueryRow("SELECT id FROM users WHERE username = ?", req.Username).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	} else if err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Создание пользователя
	userID := uuid.New().String()
	_, err = h.db.Exec(
		"INSERT INTO users (id, username, password_hash, business_name, specialization) VALUES (?, ?, ?, ?, ?)",
		userID, req.Username, string(hashedPassword), req.BusinessName, req.Specialization,
	)
	if err != nil {
		// Логируем детальную ошибку для отладки
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user",
			"details": err.Error(),
		})
		return
	}

	// Генерация JWT токена
	token := generateToken(userID, req.Username)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":              userID,
			"username":        req.Username,
			"business_name":  req.BusinessName,
			"specialization": req.Specialization,
		},
	})
}

// Login - вход пользователя
func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	// Используем COALESCE для обработки NULL значений (для старых пользователей)
	err := h.db.QueryRow(
		"SELECT id, username, password_hash, COALESCE(business_name, '') as business_name, specialization FROM users WHERE username = ?",
		req.Username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.BusinessName, &user.Specialization)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Генерация JWT токена
	token := generateToken(user.ID, user.Username)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":              user.ID,
			"username":        user.Username,
			"business_name":   user.BusinessName,
			"specialization":  user.Specialization,
		},
	})
}

// UploadFile - загрузка файла
func (h *Handler) UploadFile(c *gin.Context) {
	userID := c.GetString("user_id")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}
	defer file.Close()

	// Создание директории для файлов пользователя
	// Используем переменную окружения или путь по умолчанию
	baseUploadDir := os.Getenv("UPLOADS_DIR")
	if baseUploadDir == "" {
		// Для локального запуска используем относительный путь
		baseUploadDir = filepath.Join("..", "uploads")
	}
	uploadDir := filepath.Join(baseUploadDir, userID)
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Сохранение файла
	fileID := uuid.New().String()
	fileExt := filepath.Ext(header.Filename)
	filePath := filepath.Join(uploadDir, fileID+fileExt)

	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Сохранение информации о файле в БД
	fileSize := header.Size
	fileType := strings.ToLower(fileExt[1:]) // убираем точку

	_, err = h.db.Exec(
		"INSERT INTO files (id, user_id, filename, file_path, file_type, file_size) VALUES (?, ?, ?, ?, ?, ?)",
		fileID, userID, header.Filename, filePath, fileType, fileSize,
	)
	if err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         fileID,
		"filename":   header.Filename,
		"file_type":  fileType,
		"file_size":  fileSize,
		"uploaded_at": time.Now(),
	})
}

// GetFiles - получение списка файлов пользователя
func (h *Handler) GetFiles(c *gin.Context) {
	userID := c.GetString("user_id")

	rows, err := h.db.Query(
		"SELECT id, filename, file_type, file_size, uploaded_at FROM files WHERE user_id = ? ORDER BY uploaded_at DESC",
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get files", "files": []interface{}{}})
		return
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var f models.File
		if err := rows.Scan(&f.ID, &f.Filename, &f.FileType, &f.FileSize, &f.UploadedAt); err != nil {
			continue
		}
		files = append(files, f)
	}

	// Всегда возвращаем массив, даже если он пустой
	if files == nil {
		files = []models.File{}
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}

// DeleteFile - удаление файла
func (h *Handler) DeleteFile(c *gin.Context) {
	userID := c.GetString("user_id")
	fileID := c.Param("id")

	var filePath string
	err := h.db.QueryRow(
		"SELECT file_path FROM files WHERE id = ? AND user_id = ?",
		fileID, userID,
	).Scan(&filePath)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Удаление из БД
	_, err = h.db.Exec("DELETE FROM files WHERE id = ? AND user_id = ?", fileID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}

	// Удаление файла с диска
	os.Remove(filePath)

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

// CreateChat - создание нового чата
func (h *Handler) CreateChat(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.CreateChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Если title не указан, используем дефолтное название
		req.Title = "Новый чат"
	}

	chatID := uuid.New().String()
	now := time.Now()

	_, err := h.db.Exec(
		"INSERT INTO chats (id, user_id, title, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		chatID, userID, req.Title, now, now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         chatID,
		"title":      req.Title,
		"created_at": now,
		"updated_at": now,
	})
}

// GetChats - получение списка чатов пользователя
func (h *Handler) GetChats(c *gin.Context) {
	userID := c.GetString("user_id")

	rows, err := h.db.Query(
		"SELECT id, title, created_at, updated_at FROM chats WHERE user_id = ? ORDER BY updated_at DESC",
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get chats", "chats": []interface{}{}})
		return
	}
	defer rows.Close()

	var chats []models.Chat
	for rows.Next() {
		var chat models.Chat
		if err := rows.Scan(&chat.ID, &chat.Title, &chat.CreatedAt, &chat.UpdatedAt); err != nil {
			continue
		}
		chat.UserID = userID
		chats = append(chats, chat)
	}

	if chats == nil {
		chats = []models.Chat{}
	}

	c.JSON(http.StatusOK, gin.H{"chats": chats})
}

// DeleteChat - удаление чата
func (h *Handler) DeleteChat(c *gin.Context) {
	userID := c.GetString("user_id")
	chatID := c.Param("id")

	// Проверяем, что чат принадлежит пользователю
	var exists bool
	err := h.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM chats WHERE id = ? AND user_id = ?)",
		chatID, userID,
	).Scan(&exists)

	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat not found"})
		return
	}

	// Удаление чата (сообщения удалятся каскадно)
	_, err = h.db.Exec("DELETE FROM chats WHERE id = ? AND user_id = ?", chatID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete chat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chat deleted successfully"})
}

// SendMessage - отправка сообщения в чат
func (h *Handler) SendMessage(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Если chat_id не указан, создаем новый чат
	var chatID string
	if req.ChatID == "" {
		chatID = uuid.New().String()
		now := time.Now()
		// Генерируем название чата из первого сообщения
		title := req.Message
		if len(title) > 50 {
			title = title[:50] + "..."
		}
		_, err := h.db.Exec(
			"INSERT INTO chats (id, user_id, title, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
			chatID, userID, title, now, now,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chat"})
			return
		}
	} else {
		chatID = req.ChatID
		// Проверяем, что чат принадлежит пользователю
		var exists bool
		err := h.db.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM chats WHERE id = ? AND user_id = ?)",
			chatID, userID,
		).Scan(&exists)
		if err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Chat not found"})
			return
		}
		// Обновляем updated_at
		h.db.Exec("UPDATE chats SET updated_at = ? WHERE id = ?", time.Now(), chatID)
	}

	// Получение username, названия бизнеса и специализации пользователя
	var username, businessName, specialization string
	h.db.QueryRow("SELECT username, COALESCE(business_name, '') as business_name, specialization FROM users WHERE id = ?", userID).Scan(&username, &businessName, &specialization)

	// Получение всех файлов пользователя
	files, err := h.getUserFiles(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user files"})
		return
	}

	// Генерация ответа через AI
	response, err := ai.GenerateResponse(req.Message, req.Category, username, businessName, specialization, files)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate response"})
		return
	}

	// Сохранение сообщения в БД
	messageID := uuid.New().String()
	_, err = h.db.Exec(
		"INSERT INTO messages (id, chat_id, user_id, message, response, category) VALUES (?, ?, ?, ?, ?, ?)",
		messageID, chatID, userID, req.Message, response, req.Category,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         messageID,
		"chat_id":    chatID,
		"message":    req.Message,
		"response":   response,
		"category":   req.Category,
		"created_at": time.Now(),
	})
}

// GetUser - получение информации о текущем пользователе
func (h *Handler) GetUser(c *gin.Context) {
	userID := c.GetString("user_id")

	var user models.User
	err := h.db.QueryRow(
		"SELECT id, username, COALESCE(business_name, '') as business_name, specialization, created_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.BusinessName, &user.Specialization, &user.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Получение статистики
	var fileCount int
	h.db.QueryRow("SELECT COUNT(*) FROM files WHERE user_id = ?", userID).Scan(&fileCount)

	var messageCount int
	h.db.QueryRow("SELECT COUNT(*) FROM messages WHERE user_id = ?", userID).Scan(&messageCount)

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":             user.ID,
			"username":       user.Username,
			"specialization": user.Specialization,
			"created_at":     user.CreatedAt,
		},
		"stats": gin.H{
			"files_count":    fileCount,
			"messages_count": messageCount,
		},
	})
}

// GetChatHistory - получение истории конкретного чата
func (h *Handler) GetChatHistory(c *gin.Context) {
	userID := c.GetString("user_id")
	chatID := c.Param("chatId")

	// Проверяем, что чат принадлежит пользователю
	var exists bool
	err := h.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM chats WHERE id = ? AND user_id = ?)",
		chatID, userID,
	).Scan(&exists)

	if err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat not found", "messages": []interface{}{}})
		return
	}

	rows, err := h.db.Query(
		"SELECT id, message, response, category, created_at FROM messages WHERE chat_id = ? ORDER BY created_at ASC",
		chatID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get chat history", "messages": []interface{}{}})
		return
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.Message, &m.Response, &m.Category, &m.CreatedAt); err != nil {
			continue
		}
		m.ChatID = chatID
		messages = append(messages, m)
	}

	// Всегда возвращаем массив, даже если он пустой
	if messages == nil {
		messages = []models.Message{}
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// Вспомогательные функции

func (h *Handler) getUserFiles(userID string) ([]models.File, error) {
	rows, err := h.db.Query(
		"SELECT id, filename, file_path, file_type FROM files WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var f models.File
		if err := rows.Scan(&f.ID, &f.Filename, &f.FilePath, &f.FileType); err != nil {
			continue
		}
		files = append(files, f)
	}
	return files, nil
}

func generateToken(userID, username string) string {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 дней
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("alfa-hack-secret-key")) // В продакшене использовать переменную окружения
	return tokenString
}

