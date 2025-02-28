package models

import "time"

// Task represents a single task in a project
type Task struct {
	ID            int       `json:"id"`
	ProjectID     int       `json:"project_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Deadline      string    `json:"deadline"`
	Status        string    `json:"status"` // "Waiting", "In Progress", "Done"
	CompletedDate string    `json:"completed_date,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Project represents a freelance project
type Project struct {
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	Client   string    `json:"client"`
	Cost     float64   `json:"cost"`
	Deadline string    `json:"deadline"`
	Status   string    `json:"status"`
	Tasks    []Task    `json:"tasks"`
}

// Task status constants
const (
	TaskStatusWaiting    = "Waiting"
	TaskStatusInProgress = "In Progress"
	TaskStatusDone       = "Done"
) 