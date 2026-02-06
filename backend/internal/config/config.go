package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Upload   UploadConfig
	Office   OfficeConfig
	App      AppConfig
	OAuth    OAuthConfig
	SMTP     SMTPConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

type UploadConfig struct {
	Dir         string
	MaxSize     int64
	AllowedExts []string
}

type OfficeConfig struct {
	Latitude             float64
	Longitude            float64
	Radius               float64
	AttendanceOpenTime   string // <--- NEW FIELD
	CheckInTime          string
	CheckOutTime         string
	LateToleranceMinutes int
}

type AppConfig struct {
	Name string
	Env  string
}

type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	FrontendURL        string
}

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	UseTLS   bool
}

var Loaded *Config

// Load loads configuration from environment variables
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	jwtExpiry, err := time.ParseDuration(getEnv("JWT_EXPIRY", "24h"))
	if err != nil {
		jwtExpiry = 24 * time.Hour
	}

	maxUploadSize, err := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE", "5242880"), 10, 64)
	if err != nil {
		maxUploadSize = 5242880
	}

	latitude, err := strconv.ParseFloat(getEnv("OFFICE_LATITUDE", "-7.035549620262833"), 64)
	if err != nil {
		latitude = -7.035549620262833
	}

	longitude, err := strconv.ParseFloat(getEnv("OFFICE_LONGITUDE", "110.47464898482643"), 64)
	if err != nil {
		longitude = 110.47464898482643
	}

	radius, err := strconv.ParseFloat(getEnv("OFFICE_RADIUS", "1000"), 64)
	if err != nil {
		radius = 1000
	}

	lateTolerance, err := strconv.Atoi(getEnv("LATE_TOLERANCE_MINUTES", "15"))
	if err != nil {
		lateTolerance = 15
	}

	config := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "interna_db"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "change-this-secret-key"),
			Expiry: jwtExpiry,
		},
		Upload: UploadConfig{
			Dir:         getEnv("UPLOAD_DIR", "./uploads"),
			MaxSize:     maxUploadSize,
			AllowedExts: []string{".jpg", ".jpeg", ".png", ".pdf"},
		},
		Office: OfficeConfig{
			Latitude:             latitude,
			Longitude:            longitude,
			Radius:               radius,
			AttendanceOpenTime:   getEnv("ATTENDANCE_OPEN_TIME", "07:00:00"), // <--- NEW VALUE
			CheckInTime:          getEnv("CHECK_IN_TIME", "08:30:00"),
			CheckOutTime:         getEnv("CHECK_OUT_TIME", "16:00:00"),
			LateToleranceMinutes: lateTolerance,
		},
		App: AppConfig{
			Name: getEnv("APP_NAME", "INTERNA"),
			Env:  getEnv("APP_ENV", "development"),
		},
		OAuth: OAuthConfig{
			GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", ""),
			FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:5173"),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     getEnv("SMTP_PORT", "587"),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", ""),
			UseTLS:   getEnv("SMTP_USE_TLS", "false") == "true",
		},
	}

	Loaded = config
	return config, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
