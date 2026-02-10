package handlers

import (
	"database/sql"
	"testing"
	"time"
)

func TestNullIntToPtr(t *testing.T) {
	val := sql.NullInt64{Int64: 42, Valid: true}
	if got := nullIntToPtr(val); got == nil || *got != 42 {
		t.Fatalf("nullIntToPtr(%v) = %v, want pointer to 42", val, got)
	}

	if got := nullIntToPtr(sql.NullInt64{}); got != nil {
		t.Fatalf("nullIntToPtr(invalid) = %v, want nil", got)
	}
}

func TestNullStringToPtr(t *testing.T) {
	val := sql.NullString{String: "hello", Valid: true}
	if got := nullStringToPtr(val); got == nil || *got != "hello" {
		t.Fatalf("nullStringToPtr(%v) = %v, want pointer to \"hello\"", val, got)
	}

	if got := nullStringToPtr(sql.NullString{}); got != nil {
		t.Fatalf("nullStringToPtr(invalid) = %v, want nil", got)
	}
}

func TestNullInt64ToPtr(t *testing.T) {
	val := sql.NullInt64{Int64: 99, Valid: true}
	if got := nullInt64ToPtr(val); got == nil || *got != 99 {
		t.Fatalf("nullInt64ToPtr(%v) = %v, want pointer to 99", val, got)
	}

	if got := nullInt64ToPtr(sql.NullInt64{}); got != nil {
		t.Fatalf("nullInt64ToPtr(invalid) = %v, want nil", got)
	}
}

func TestNullTimeToPtr(t *testing.T) {
	now := time.Now()
	val := sql.NullTime{Time: now, Valid: true}
	if got := nullTimeToPtr(val); got == nil || !got.Equal(now) {
		t.Fatalf("nullTimeToPtr(%v) = %v, want pointer to %v", val, got, now)
	}

	if got := nullTimeToPtr(sql.NullTime{}); got != nil {
		t.Fatalf("nullTimeToPtr(invalid) = %v, want nil", got)
	}
}
