package ui

// Project represents a project in the UI layer
type Project struct {
	ID       int
	Name     string
	Client   string
	Cost     float64
	Deadline string
	Status   string
	Tasks    []Task
}

// Task represents a task in the UI layer
type Task struct {
	ID          int
	ProjectID   int
	Title       string
	Description string
	Status      string
} 