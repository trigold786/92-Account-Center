package model

import "time"

type Role struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type RolePermission struct {
	ID         int64     `json:"id"`
	RoleID     int64     `json:"role_id"`
	Permission string    `json:"permission"`
	CreatedAt  time.Time `json:"created_at"`
}

type UserRole struct {
	ID        int64     `json:"id"`
	UserID    string    `json:"user_id"`
	RoleID    int64     `json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}
