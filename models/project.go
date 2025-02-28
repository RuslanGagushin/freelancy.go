package models

import "time"

type Project struct {
	ID          int
	Name        string
	Client      string
	Cost        float64
	Deadline    time.Time
	Tasks       []Task
	CreatedAt   time.Time
}

type Task struct {
	ID          int
	ProjectID   int
	Title       string
	Description string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
} 