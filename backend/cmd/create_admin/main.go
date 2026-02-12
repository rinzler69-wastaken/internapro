package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Try to load .env from current directory or parent directories
	// Since we run this via 'go run cmd/create_admin/main.go' from 'backend/',
	// or 'go run backend/cmd/create_admin/main.go' from root, we need to find .env.
	// We'll trust the running context or standard locations.
	if err := godotenv.Load(); err != nil {
		// Try loading from parent directory (backend root) if not found
		if err := godotenv.Load("../.env"); err != nil {
			if err := godotenv.Load("../../.env"); err != nil {
				log.Println("⚠️  Warning: No .env file found. Relying on system environment variables.")
			}
		}
	}

	db := openDB()
	defer db.Close()

	fmt.Println("✅ Connected to database")
	fmt.Println("")

	seedAdmin(db)
}

func openDB() *sql.DB {
	dbUser := env("DB_USER", "root")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := env("DB_HOST", "localhost")
	dbPort := env("DB_PORT", "3306")
	dbName := env("DB_NAME", "interna_db")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	return db
}

func seedAdmin(db *sql.DB) {
	email := "office@klikdsi.com"
	password := "internapro2025"
	name := "DSI-Admin"
	role := "admin"

	fmt.Printf("Creating/Updating Admin Account...\n")
	fmt.Printf("Email: %s\n", email)
	fmt.Printf("Name: %s\n", name)
	fmt.Printf("Role: %s\n", role)

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// upsert logic
	query := `
		INSERT INTO users (name, email, password_hash, role, is_2fa_enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, 0, NOW(), NOW())
		ON DUPLICATE KEY UPDATE 
			name = VALUES(name), 
			password_hash = VALUES(password_hash), 
			role = VALUES(role),
			updated_at = NOW();
	`

	res, err := db.Exec(query, name, email, string(hash), role)
	if err != nil {
		log.Fatal("Failed to upsert admin user:", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		// This technically shouldn't happen with the updated_at logic, but just in case
		fmt.Println("ℹ️  Admin user already existed with identical data.")
	} else {
		fmt.Println("✅ Admin user successfully created/updated.")
	}

	fmt.Println("-------------------------------------------")
	fmt.Println("LOGIN CREDENTIALS:")
	fmt.Printf("Email:    %s\n", email)
	fmt.Printf("Password: %s\n", password)
	fmt.Println("-------------------------------------------")
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
