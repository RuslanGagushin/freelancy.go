package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"freelancy.go/internal/models"
)

// Storage handles data persistence for projects and tasks
type Storage struct {
	dataFile string
	Projects []models.Project `json:"projects"`
}

// NewStorage creates a new storage instance and initializes the data file
func NewStorage() (*Storage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dataDir := filepath.Join(homeDir, ".freelancy")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	dataFile := filepath.Join(dataDir, "data.json")
	storage := &Storage{
		dataFile: dataFile,
	}

	if err := storage.Load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		storage.Projects = make([]models.Project, 0)
		if err := storage.Save(); err != nil {
			return nil, err
		}
	}

	return storage, nil
}

// Load reads data from the storage file
func (s *Storage) Load() error {
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, s)
}

// Save writes data to the storage file
func (s *Storage) Save() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	file, err := os.OpenFile(s.dataFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

// AddProject adds a new project to storage
func (s *Storage) AddProject(project models.Project) error {
	if len(s.Projects) == 0 {
		project.ID = 1
	} else {
		project.ID = s.Projects[len(s.Projects)-1].ID + 1
	}
	if project.Status == "" {
		project.Status = "Active"
	}
	s.Projects = append(s.Projects, project)
	return s.Save()
}

// AddTask adds a new task to the specified project
func (s *Storage) AddTask(projectID int, task models.Task) error {
	for i, p := range s.Projects {
		if p.ID == projectID {
			if len(p.Tasks) == 0 {
				task.ID = 1
			} else {
				task.ID = p.Tasks[len(p.Tasks)-1].ID + 1
			}
			task.ProjectID = projectID
			task.CreatedAt = time.Now()
			task.UpdatedAt = time.Now()
			s.Projects[i].Tasks = append(s.Projects[i].Tasks, task)
			return s.Save()
		}
	}
	return nil
}

// GetProjects returns all projects sorted by status (active first)
func (s *Storage) GetProjects() []models.Project {
	var activeProjects []models.Project
	var completedProjects []models.Project

	for _, project := range s.Projects {
		if project.Status == "Completed" {
			completedProjects = append(completedProjects, project)
		} else {
			activeProjects = append(activeProjects, project)
		}
	}

	sortedProjects := make([]models.Project, 0, len(s.Projects))
	sortedProjects = append(sortedProjects, activeProjects...)
	sortedProjects = append(sortedProjects, completedProjects...)

	return sortedProjects
}

// GetTasks returns all tasks from all projects
func (s *Storage) GetTasks() []models.Task {
	var allTasks []models.Task
	for _, p := range s.Projects {
		allTasks = append(allTasks, p.Tasks...)
	}
	return allTasks
}

// UpdateProjectStatus changes the status of a project
func (s *Storage) UpdateProjectStatus(projectID int, status string) error {
	for i, p := range s.Projects {
		if p.ID == projectID {
			s.Projects[i].Status = status
			return s.Save()
		}
	}
	return fmt.Errorf("project not found")
}

// DeleteProject removes a project and all its tasks
func (s *Storage) DeleteProject(projectID int) error {
	for i, p := range s.Projects {
		if p.ID == projectID {
			s.Projects = append(s.Projects[:i], s.Projects[i+1:]...)
			return s.Save()
		}
	}
	return fmt.Errorf("project not found")
}

// UpdateTaskStatus changes the status of a task and updates completion date
func (s *Storage) UpdateTaskStatus(projectID, taskID int, status string, completedDate string) error {
	for i, p := range s.Projects {
		if p.ID == projectID {
			for j, t := range p.Tasks {
				if t.ID == taskID {
					s.Projects[i].Tasks[j].Status = status
					s.Projects[i].Tasks[j].UpdatedAt = time.Now()
					if status == models.TaskStatusDone {
						s.Projects[i].Tasks[j].CompletedDate = completedDate
					} else {
						s.Projects[i].Tasks[j].CompletedDate = ""
					}
					return s.Save()
				}
			}
		}
	}
	return fmt.Errorf("task not found")
}

// DeleteTask removes a task from its project
func (s *Storage) DeleteTask(projectID, taskID int) error {
	for i, p := range s.Projects {
		if p.ID == projectID {
			for j, t := range p.Tasks {
				if t.ID == taskID {
					s.Projects[i].Tasks = append(s.Projects[i].Tasks[:j], s.Projects[i].Tasks[j+1:]...)
					return s.Save()
				}
			}
		}
	}
	return fmt.Errorf("task not found")
} 