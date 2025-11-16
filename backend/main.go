package main

import (
	"alfa-hack-backend/internal/api"
	"alfa-hack-backend/internal/database"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Загрузка переменных окружения из .env файла (если существует)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Инициализация базы данных
	// Используем переменную окружения или путь по умолчанию
	dbDir := os.Getenv("DB_DIR")
	if dbDir == "" {
		// Для локального запуска используем относительный путь
		dbDir = "../database"
	}
	if err := os.MkdirAll(dbDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}
	dbPath := filepath.Join(dbDir, "alfa_hack.db")
	db, err := database.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Создание таблиц
	if err := database.CreateTables(db); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Инициализация роутера
	router := gin.Default()

	// Настройка CORS
	config := cors.DefaultConfig()
	// Разрешаем запросы с frontend (локально или из Docker)
	allowedOrigins := os.Getenv("CORS_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:3000"
	}
	// Используем функцию для проверки origin (поддержка нескольких источников)
	config.AllowOriginFunc = func(origin string) bool {
		allowed := []string{allowedOrigins, "http://localhost:3000", "http://frontend:3000"}
		for _, allowedOrigin := range allowed {
			if origin == allowedOrigin {
				return true
			}
		}
		return false
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	config.AllowCredentials = true
	router.Use(cors.New(config))

	// Инициализация API handlers
	apiHandler := api.NewHandler(db)

	// API routes
	apiRoutes := router.Group("/api")
	{
		// Аутентификация
		apiRoutes.POST("/register", apiHandler.Register)
		apiRoutes.POST("/login", apiHandler.Login)

		// Защищенные routes
		protected := apiRoutes.Group("/")
		protected.Use(apiHandler.AuthMiddleware())
		{
			// Пользователь
			protected.GET("/user", apiHandler.GetUser)

			// Файлы
			protected.POST("/files/upload", apiHandler.UploadFile)
			protected.GET("/files", apiHandler.GetFiles)
			protected.DELETE("/files/:id", apiHandler.DeleteFile)

			// Чаты
			protected.POST("/chats", apiHandler.CreateChat)
			protected.GET("/chats", apiHandler.GetChats)
			protected.DELETE("/chats/:id", apiHandler.DeleteChat)

			// Сообщения
			protected.POST("/chat", apiHandler.SendMessage)
			protected.GET("/chat/:chatId/history", apiHandler.GetChatHistory)
		}
	}

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
