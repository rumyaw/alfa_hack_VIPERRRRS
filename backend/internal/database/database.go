package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath+"?_foreign_keys=1")
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Database connection established")
	return db, nil
}

func CreateTables(db *sql.DB) error {
	queries := []string{
		// Таблица пользователей
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			business_name TEXT NOT NULL,
			specialization TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		

		// Таблица файлов
		`CREATE TABLE IF NOT EXISTS files (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			filename TEXT NOT NULL,
			file_path TEXT NOT NULL,
			file_type TEXT,
			file_size INTEGER,
			uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		// Таблица чатов
		`CREATE TABLE IF NOT EXISTS chats (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			title TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		// Таблица сообщений чата
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			chat_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			message TEXT NOT NULL,
			response TEXT,
			category TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (chat_id) REFERENCES chats(id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		// Индекс для быстрого поиска
		`CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_chats_user_id ON chats(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages(chat_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_user_id ON messages(user_id)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	// Миграция: добавление колонки business_name если её нет
	err := addColumnIfNotExists(db, "users", "business_name", "TEXT DEFAULT ''")
	if err != nil {
		log.Printf("Warning: Failed to add business_name column (might already exist): %v", err)
	}

	log.Println("Database tables created successfully")
	return nil
}

// addColumnIfNotExists добавляет колонку в таблицу, если её нет
func addColumnIfNotExists(db *sql.DB, tableName, columnName, columnDef string) error {
	// Проверяем, существует ли колонка
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name='%s'", tableName, columnName)
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return err
	}

	// Если колонка не существует, добавляем её
	if count == 0 {
		alterQuery := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnDef)
		_, err = db.Exec(alterQuery)
		if err != nil {
			return err
		}
		log.Printf("Added column %s to table %s", columnName, tableName)
	}

	return nil
}
