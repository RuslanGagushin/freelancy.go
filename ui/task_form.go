package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TaskForm struct {
	inputs     []textinput.Model
	focusIndex int
	done       bool
	projectID  int
}

func NewTaskForm(projectID int) TaskForm {
	inputs := make([]textinput.Model, 3)
	
	// Title input
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Task Title"
	inputs[0].Focus()
	
	// Description input
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Description"
	
	// Deadline input
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Deadline (YYYY-MM-DD)"
	
	return TaskForm{
		inputs:     inputs,
		focusIndex: 0,
		projectID:  projectID,
	}
}

func (m TaskForm) Init() tea.Cmd {
	return textinput.Blink
}

func (m TaskForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()
			
			if s == "enter" && m.focusIndex == len(m.inputs)-1 {
				m.done = true
				return m, nil
			}
			
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}
			
			if m.focusIndex >= len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}
			
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i < len(m.inputs); i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					continue
				}
				m.inputs[i].Blur()
			}
			return m, tea.Batch(cmds...)
		}
	}
	
	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *TaskForm) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(m.inputs))
	
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	
	return tea.Batch(cmds...)
}

func (m TaskForm) View() string {
	var b strings.Builder
	
	b.WriteString("Create New Task\n\n")
	
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}
	
	button := "[ Submit ]"
	if m.focusIndex == len(m.inputs) {
		button = "[ " + lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("Submit") + " ]"
	}
	b.WriteString("\n\n" + button + "\n")
	
	return b.String()
}

func (m TaskForm) Done() bool {
	return m.done
}

func (m TaskForm) GetValues() (string, string, string) {
	return m.inputs[0].Value(),
		m.inputs[1].Value(),
		m.inputs[2].Value()
}

func (m TaskForm) GetProjectID() int {
	return m.projectID
} 