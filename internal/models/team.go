package models

type Team struct {
	TeamName string       `json:"team_name" db:"name" binding:"required,min=1,max=100"`
	Members  []TeamMember `json:"members" binding:"required,dive"`
}

type TeamMember struct {
	UserID   string `json:"user_id" db:"id" binding:"required,min=1,max=50"`
	Username string `json:"username" db:"name" binding:"required,min=1,max=100"`
	IsActive bool   `json:"is_active" db:"is_active" binding:"required"`
}
