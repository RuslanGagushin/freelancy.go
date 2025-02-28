package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"freelancy.go/internal/models"
	"freelancy.go/storage"
	"freelancy.go/ui"
)

type model struct {
	storage     *storage.Storage
	activeView  string // "projects", "tasks", "new_project", "new_task", "income"
	projectList ProjectList
	taskTable   TaskTable
	projectForm ui.ProjectForm
	taskForm    ui.TaskForm
	incomeChart ui.IncomeChart
}

type ProjectList struct {
	projects  []models.Project
	selected  int
	style     lipgloss.Style
}

type TaskTable struct {
	tasks   []models.Task
	cursor  int
	focused string // "waiting", "in_progress", "done"
}

// Task status constants
const (
	TaskStatusWaiting    = models.TaskStatusWaiting
	TaskStatusInProgress = models.TaskStatusInProgress
	TaskStatusDone       = models.TaskStatusDone
)

func initialModel() model {
	storage, err := storage.NewStorage()
	if err != nil {
		fmt.Printf("Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	return model{
		storage:    storage,
		activeView: "projects",
		projectList: ProjectList{
			projects: storage.GetProjects(),
			selected: 0,
			style: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("63")).
				Padding(1),
		},
		taskTable: TaskTable{
			tasks:   make([]models.Task, 0),
			cursor:  0,
			focused: "waiting",
		},
		projectForm: ui.NewProjectForm(),
		taskForm:   ui.NewTaskForm(0),
		incomeChart: ui.NewIncomeChart(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle common commands
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			switch m.activeView {
			case "projects":
				m.activeView = "tasks"
				m.updateTaskTable()
			case "tasks":
				m.activeView = "income"
				m.projectList.projects = m.storage.GetProjects()
				var uiProjects []ui.Project
				for _, p := range m.projectList.projects {
					var uiTasks []ui.Task
					for _, t := range p.Tasks {
						uiTasks = append(uiTasks, ui.Task{
							ID:          t.ID,
							ProjectID:   t.ProjectID,
							Title:       t.Title,
							Description: t.Description,
							Status:      t.Status,
						})
					}
					uiProjects = append(uiProjects, ui.Project{
						ID:       p.ID,
						Name:     p.Name,
						Client:   p.Client,
						Cost:     p.Cost,
						Deadline: p.Deadline,
						Status:   p.Status,
						Tasks:    uiTasks,
					})
				}
				m.incomeChart.UpdateData(uiProjects)
			case "income":
				m.activeView = "projects"
			}
			return m, nil
		case "n":
			if m.activeView == "projects" {
				m.activeView = "new_project"
				m.projectForm = ui.NewProjectForm()
				return m, nil
			}
		case "t":
			if m.activeView == "projects" && len(m.projectList.projects) > 0 {
				m.activeView = "new_task"
				m.taskForm = ui.NewTaskForm(m.projectList.projects[m.projectList.selected].ID)
				return m, nil
			}
		case "s":
			if m.activeView == "tasks" {
				// Get current task
				var currentTask *models.Task
				var tasks []models.Task

				switch m.taskTable.focused {
				case "waiting":
					tasks = filterTasks(m.taskTable.tasks, models.TaskStatusWaiting)
				case "in_progress":
					tasks = filterTasks(m.taskTable.tasks, models.TaskStatusInProgress)
				case "done":
					tasks = filterTasks(m.taskTable.tasks, models.TaskStatusDone)
				}

				if m.taskTable.cursor < len(tasks) {
					currentTask = &tasks[m.taskTable.cursor]
				}

				if currentTask != nil {
					// Determine next status
					var newStatus string
					switch currentTask.Status {
					case models.TaskStatusWaiting:
						newStatus = models.TaskStatusInProgress
					case models.TaskStatusInProgress:
						newStatus = models.TaskStatusDone
					case models.TaskStatusDone:
						newStatus = models.TaskStatusWaiting
					}

					completedDate := ""
					if newStatus == models.TaskStatusDone {
						completedDate = time.Now().Format("2006-01-02")
					}

					if err := m.storage.UpdateTaskStatus(currentTask.ProjectID, currentTask.ID, newStatus, completedDate); err != nil {
						fmt.Printf("Error updating task status: %v\n", err)
					}
					m.updateTaskTable()
				}
				return m, nil
			} else if m.activeView == "projects" && len(m.projectList.projects) > 0 {
				project := m.projectList.projects[m.projectList.selected]
				newStatus := "Completed"
				if project.Status == "Completed" {
					newStatus = "Active"
				}
				if err := m.storage.UpdateProjectStatus(project.ID, newStatus); err != nil {
					fmt.Printf("Error updating project status: %v\n", err)
				}
				m.projectList.projects = m.storage.GetProjects()
				return m, nil
			}
		case "d":
			if m.activeView == "projects" && len(m.projectList.projects) > 0 {
				project := m.projectList.projects[m.projectList.selected]
				if err := m.storage.DeleteProject(project.ID); err != nil {
					fmt.Printf("Error deleting project: %v\n", err)
				}
				m.projectList.projects = m.storage.GetProjects()
				if m.projectList.selected >= len(m.projectList.projects) {
					m.projectList.selected = len(m.projectList.projects) - 1
				}
				if m.projectList.selected < 0 {
					m.projectList.selected = 0
				}
				return m, nil
			} else if m.activeView == "tasks" {
				// Get current task
				var currentTask *models.Task
				var tasks []models.Task

				switch m.taskTable.focused {
				case "waiting":
					tasks = filterTasks(m.taskTable.tasks, models.TaskStatusWaiting)
				case "in_progress":
					tasks = filterTasks(m.taskTable.tasks, models.TaskStatusInProgress)
				case "done":
					tasks = filterTasks(m.taskTable.tasks, models.TaskStatusDone)
				}

				if m.taskTable.cursor < len(tasks) {
					currentTask = &tasks[m.taskTable.cursor]
				}

				if currentTask != nil {
					if err := m.storage.DeleteTask(currentTask.ProjectID, currentTask.ID); err != nil {
						fmt.Printf("Error deleting task: %v\n", err)
					}
					m.updateTaskTable()
				}
				return m, nil
			}
		case "esc":
			if m.activeView == "new_project" || m.activeView == "new_task" {
				m.activeView = "projects"
				return m, nil
			}
		case "up", "down", "left", "right":
			if m.activeView == "projects" {
				if keyMsg.String() == "up" {
					m.projectList.selected--
				} else if keyMsg.String() == "down" {
					m.projectList.selected++
				}

				if m.projectList.selected >= len(m.projectList.projects) {
					m.projectList.selected = 0
				} else if m.projectList.selected < 0 {
					m.projectList.selected = len(m.projectList.projects) - 1
				}
				return m, nil
			} else if m.activeView == "tasks" {
				switch keyMsg.String() {
				case "left":
					switch m.taskTable.focused {
					case "in_progress":
						m.taskTable.focused = "waiting"
					case "done":
						m.taskTable.focused = "in_progress"
					}
					m.taskTable.cursor = 0
				case "right":
					switch m.taskTable.focused {
					case "waiting":
						m.taskTable.focused = "in_progress"
					case "in_progress":
						m.taskTable.focused = "done"
					}
					m.taskTable.cursor = 0
				case "up":
					m.taskTable.cursor--
					var tasks []models.Task
					switch m.taskTable.focused {
					case "waiting":
						tasks = filterTasks(m.taskTable.tasks, models.TaskStatusWaiting)
					case "in_progress":
						tasks = filterTasks(m.taskTable.tasks, models.TaskStatusInProgress)
					case "done":
						tasks = filterTasks(m.taskTable.tasks, models.TaskStatusDone)
					}
					if m.taskTable.cursor < 0 {
						m.taskTable.cursor = len(tasks) - 1
					}
				case "down":
					var tasks []models.Task
					switch m.taskTable.focused {
					case "waiting":
						tasks = filterTasks(m.taskTable.tasks, models.TaskStatusWaiting)
					case "in_progress":
						tasks = filterTasks(m.taskTable.tasks, models.TaskStatusInProgress)
					case "done":
						tasks = filterTasks(m.taskTable.tasks, models.TaskStatusDone)
					}
					m.taskTable.cursor++
					if m.taskTable.cursor >= len(tasks) {
						m.taskTable.cursor = 0
					}
				}
				return m, nil
			}
		}
	}

	// Handle view-specific updates
	switch m.activeView {
	case "income":
		m.incomeChart, cmd = m.incomeChart.Update(msg)
	case "new_project":
		var formModel tea.Model
		formModel, cmd = m.projectForm.Update(msg)
		m.projectForm = formModel.(ui.ProjectForm)

		if m.projectForm.Done() {
			name, client, costStr, deadlineStr := m.projectForm.GetValues()
			cost, _ := strconv.ParseFloat(costStr, 64)
			
			newProject := models.Project{
				Name:     name,
				Client:   client,
				Cost:     cost,
				Deadline: deadlineStr,
				Tasks:    make([]models.Task, 0),
			}
			
			if err := m.storage.AddProject(newProject); err != nil {
				fmt.Printf("Error saving project: %v\n", err)
			}
			
			m.projectList.projects = m.storage.GetProjects()
			m.activeView = "projects"
		}
	case "new_task":
		var formModel tea.Model
		formModel, cmd = m.taskForm.Update(msg)
		m.taskForm = formModel.(ui.TaskForm)

		if m.taskForm.Done() {
			title, description, deadline := m.taskForm.GetValues()
			projectID := m.taskForm.GetProjectID()
			
			newTask := models.Task{
				Title:       title,
				Description: description,
				Deadline:    deadline,
				Status:      models.TaskStatusWaiting,
			}
			
			if err := m.storage.AddTask(projectID, newTask); err != nil {
				fmt.Printf("Error saving task: %v\n", err)
			}
			
			m.projectList.projects = m.storage.GetProjects()
			m.updateTaskTable()
			m.activeView = "projects"
		}
	}

	return m, cmd
}

func (m *model) updateTaskTable() {
	m.taskTable.tasks = m.storage.GetTasks()
}

func (m model) View() string {
	switch m.activeView {
	case "projects":
		return m.renderProjects()
	case "tasks":
		return m.renderTasks()
	case "new_project":
		return m.projectForm.View()
	case "new_task":
		return m.taskForm.View()
	case "income":
		return m.incomeChart.View()
	default:
		return "Unknown view"
	}
}

func (m model) renderProjects() string {
	var s string
	s += "Projects (TAB: switch view, N: new project, T: new task, S: toggle status, D: delete project, ↑/↓: select, Q: quit)\n\n"

	// Define styles for project card
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1).
		Width(30)

	// Number of projects per row
	projectsPerRow := 3

	// Create project rows
	for i := 0; i < len(m.projectList.projects); i += projectsPerRow {
		var rowCards []string
		
		// Add up to 3 projects to a row
		for j := 0; j < projectsPerRow && i+j < len(m.projectList.projects); j++ {
			p := m.projectList.projects[i+j]
			style := cardStyle.Copy()
			
			// Highlight selected project
			if i+j == m.projectList.selected {
				style = style.BorderForeground(lipgloss.Color("205"))
			}

			// Style for status
			statusStyle := lipgloss.NewStyle()
			if p.Status == "Completed" {
				statusStyle = statusStyle.Foreground(lipgloss.Color("42"))
			} else {
				statusStyle = statusStyle.Foreground(lipgloss.Color("208"))
			}

			// Form project card content
			card := fmt.Sprintf(
				"Project: %s\nClient: %s\nCost: $%.2f\nDeadline: %s\nStatus: %s\nTasks: %d",
				p.Name, p.Client, p.Cost, p.Deadline,
				statusStyle.Render(p.Status),
				len(p.Tasks),
			)
			
			rowCards = append(rowCards, style.Render(card))
		}
		
		// Join cards in row with spaces between them
		s += lipgloss.JoinHorizontal(lipgloss.Top, rowCards...) + "\n\n"
	}
	
	return s
}

func (m model) renderTasks() string {
	s := "Tasks View (TAB: switch views, S: change status, D: delete task, ←/→: switch columns, ↑/↓: select, Q to quit)\n\n"

	// Define styles for columns and cards
	columnStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0).
		Width(30)

	focusedColumnStyle := columnStyle.Copy().
		BorderForeground(lipgloss.Color("205"))

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0).
		Width(28)

	selectedCardStyle := cardStyle.Copy().
		BorderForeground(lipgloss.Color("205"))

	// Group tasks by status
	waitingTasks := make([]models.Task, 0)
	inProgressTasks := make([]models.Task, 0)
	doneTasks := make([]models.Task, 0)

	for _, task := range m.taskTable.tasks {
		switch task.Status {
		case models.TaskStatusWaiting:
			waitingTasks = append(waitingTasks, task)
		case models.TaskStatusInProgress:
			inProgressTasks = append(inProgressTasks, task)
		case models.TaskStatusDone:
			doneTasks = append(doneTasks, task)
		}
	}

	// Function to render task card
	renderTaskCard := func(task models.Task, isSelected bool) string {
		style := cardStyle
		if isSelected {
			style = selectedCardStyle
		}

		var projectName string
		for _, p := range m.projectList.projects {
			if p.ID == task.ProjectID {
				projectName = p.Name
				break
			}
		}

		return style.Render(fmt.Sprintf(
			"%s\n%s | %s",
			task.Title,
			projectName,
			task.Deadline,
		))
	}

	// Render columns
	waitingColumn := "Waiting\n\n"
	inProgressColumn := "In Progress\n\n"
	doneColumn := "Done\n\n"

	// Add cards to columns
	for i, task := range waitingTasks {
		isSelected := m.taskTable.focused == "waiting" && m.taskTable.cursor == i
		waitingColumn += renderTaskCard(task, isSelected) + "\n"
	}

	for i, task := range inProgressTasks {
		isSelected := m.taskTable.focused == "in_progress" && m.taskTable.cursor == i
		inProgressColumn += renderTaskCard(task, isSelected) + "\n"
	}

	for i, task := range doneTasks {
		isSelected := m.taskTable.focused == "done" && m.taskTable.cursor == i
		doneColumn += renderTaskCard(task, isSelected) + "\n"
	}

	// Apply styles to columns
	waitingStyle := columnStyle
	inProgressStyle := columnStyle
	doneStyle := columnStyle

	switch m.taskTable.focused {
	case "waiting":
		waitingStyle = focusedColumnStyle
	case "in_progress":
		inProgressStyle = focusedColumnStyle
	case "done":
		doneStyle = focusedColumnStyle
	}

	waitingColumn = waitingStyle.Render(waitingColumn)
	inProgressColumn = inProgressStyle.Render(inProgressColumn)
	doneColumn = doneStyle.Render(doneColumn)

	// Join columns
	return s + lipgloss.JoinHorizontal(lipgloss.Top, waitingColumn, inProgressColumn, doneColumn)
}

// Helper function to filter tasks by status
func filterTasks(tasks []models.Task, status string) []models.Task {
	var filtered []models.Task
	for _, task := range tasks {
		if task.Status == status {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
} 