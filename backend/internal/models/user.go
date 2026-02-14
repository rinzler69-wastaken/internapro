package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID                 int64          `json:"id"`
	Name               sql.NullString `json:"name"`
	Email              string         `json:"email"`
	EmailVerifiedAt    sql.NullTime   `json:"email_verified_at,omitempty"`
	PasswordHash       sql.NullString `json:"-"`    // Never expose password hash
	Role               string         `json:"role"` // admin, supervisor, pembimbing, intern
	Avatar             sql.NullString `json:"avatar,omitempty"`
	GoogleID           sql.NullString `json:"-"`
	Provider           sql.NullString `json:"-"`
	TOTPSecret         sql.NullString `json:"-"` // Never expose TOTP secret
	Is2FAEnabled       bool           `json:"is_2fa_enabled"`
	Google2FASecret    sql.NullString `json:"-"`
	Google2FAEnabled   bool           `json:"google2fa_enabled"`
	TwoFactorSecret    sql.NullString `json:"-"`
	TwoFactorRecovery  sql.NullString `json:"-"`
	TwoFactorConfirmed sql.NullTime   `json:"-"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

type Supervisor struct {
	ID           int64          `json:"id"`
	UserID       int64          `json:"user_id"`
	FullName     string         `json:"full_name"`
	Email        string         `json:"email"`
	NIP          *string        `json:"nip,omitempty"`
	Phone        *string        `json:"phone,omitempty"`
	Position     *string        `json:"position,omitempty"`
	Address      *string        `json:"address,omitempty"`
	Institution  *string        `json:"institution,omitempty"`
	Status       string         `json:"status"`
	InternsCount int64          `json:"interns_count,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	Avatar       sql.NullString `json:"avatar"`
}

type Institution struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
