package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID `json:"id"`
	PhoneNumber  string    `json:"phone_number"`
	TelegramID   string    `json:"telegram_id"`
	Username     string    `json:"username"`
	RegisteredAt time.Time `json:"registered_at"`
	Role         string    `json:"role"` // user, admin, judge
}

// Problem represents a coding problem
type Problem struct {
	ID         uuid.UUID `json:"id"`
	Day        int       `json:"day"`
	Type       string    `json:"type"` // dsa, linux, build
	Slug       string    `json:"slug"`
	Title      string    `json:"title"`
	FilePath   string    `json:"file_path"`
	Score      int       `json:"score"`
	UnlockTime time.Time `json:"unlock_time"`
}

// Submission represents a user's submission for a problem
type Submission struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	ProblemID   uuid.UUID `json:"problem_id"`
	Language    string    `json:"language"`
	Status      string    `json:"status"` // pending, passed, failed, error
	Output      string    `json:"output"`
	Score       int       `json:"score"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// LeaderboardEntry represents a row in the leaderboard
type LeaderboardEntry struct {
	UserID     uuid.UUID `json:"user_id"`
	Username   string    `json:"username"`
	TotalScore int       `json:"total_score"`
	Rank       int       `json:"rank"`
}

// Testcase represents a single test case for a problem
type Testcase struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
	IsHidden       bool   `json:"is_hidden"`
}
