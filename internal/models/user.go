package models

type User struct {
	UserID   string `json:"user_id" db:"id"`
	Username string `json:"username" db:"name"`
	TeamName string `json:"team_name" db:"team_name"`
	IsActive bool   `json:"is_active" db:"is_active"`
}

type SetActiveRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	IsActive bool   `json:"is_active"`
}
