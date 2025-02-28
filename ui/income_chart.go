package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MonthIncome struct {
	Month    string
	Income   float64
	Projects []string
}

type IncomeChart struct {
	monthlyIncomes []MonthIncome
	maxIncome     float64
	graphHeight   int
	style         lipgloss.Style
	selected      int
}

func NewIncomeChart() IncomeChart {
	return IncomeChart{
		monthlyIncomes: make([]MonthIncome, 12),
		graphHeight:    15,
		style: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1),
		selected: -1,
	}
}

func (ic *IncomeChart) UpdateData(projects []Project) {
	// Получаем текущую дату
	now := time.Now()

	// Инициализируем массив месяцев
	for i := 0; i < 12; i++ {
		date := now.AddDate(0, -(11-i), 0)
		ic.monthlyIncomes[i] = MonthIncome{
			Month:    date.Format("Jan 2006"),
			Income:   0,
			Projects: make([]string, 0),
		}
	}

	for _, project := range projects {
		if project.Status != "Completed" {
			continue
		}

		deadline, err := time.Parse("2006-01-02", project.Deadline)
		if err != nil {
			continue
		}

		deadlineStr := deadline.Format("Jan 2006")
		for i, monthIncome := range ic.monthlyIncomes {
			if monthIncome.Month == deadlineStr {
				ic.monthlyIncomes[i].Income += project.Cost
				ic.monthlyIncomes[i].Projects = append(
					ic.monthlyIncomes[i].Projects,
					fmt.Sprintf("%s ($%.2f)", project.Name, project.Cost),
				)
				break
			}
		}
	}

	// Находим максимальный доход
	ic.maxIncome = 0
	for _, mi := range ic.monthlyIncomes {
		if mi.Income > ic.maxIncome {
			ic.maxIncome = mi.Income
		}
	}
}

func (ic IncomeChart) Update(msg tea.Msg) (IncomeChart, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			if ic.selected > 0 {
				ic.selected--
			} else if ic.selected == -1 {
				ic.selected = 11
			}
		case "right":
			if ic.selected < 11 {
				ic.selected++
			}
		case "esc":
			ic.selected = -1
		}
	}
	return ic, nil
}

func (ic IncomeChart) View() string {
	var s strings.Builder
	s.WriteString("Income Chart (← →: select month, ESC: deselect, Q: quit)\n\n")

	// Создаем график
	heightMultiplier := float64(ic.graphHeight) / ic.maxIncome
	if ic.maxIncome == 0 {
		heightMultiplier = 0
	}

	graph := make([][]string, ic.graphHeight)
	for i := range graph {
		graph[i] = make([]string, 12)
		for j := range graph[i] {
			graph[i][j] = " "
		}
	}

	// Заполняем столбцы
	barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Background(lipgloss.Color("236"))

	for month := 0; month < 12; month++ {
		income := ic.monthlyIncomes[month].Income
		height := int(income * heightMultiplier)
		style := barStyle
		if month == ic.selected {
			style = selectedStyle
		}
		for i := 0; i < height; i++ {
			graph[ic.graphHeight-1-i][month] = style.Render("█")
		}
	}

	// Отрисовываем график
	for i := 0; i < ic.graphHeight; i++ {
		value := int(float64(ic.graphHeight-i) * ic.maxIncome / float64(ic.graphHeight))
		s.WriteString(fmt.Sprintf("%6d │", value))
		
		for j := 0; j < 12; j++ {
			s.WriteString(graph[i][j] + " ")
		}
		s.WriteString("\n")
	}

	// Добавляем ось X
	s.WriteString("       └" + strings.Repeat("─", 24) + "\n")

	// Добавляем подписи месяцев
	s.WriteString("        ")
	monthStyle := lipgloss.NewStyle()
	selectedMonthStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))

	for i, mi := range ic.monthlyIncomes {
		style := monthStyle
		if i == ic.selected {
			style = selectedMonthStyle
		}
		s.WriteString(style.Render(fmt.Sprintf("%-2s", mi.Month[:2])))
	}
	s.WriteString("\n\n")

	// Показываем детальную информацию о выбранном месяце
	if ic.selected >= 0 {
		mi := ic.monthlyIncomes[ic.selected]
		s.WriteString(fmt.Sprintf("%s: $%.2f\n", mi.Month, mi.Income))
		if len(mi.Projects) > 0 {
			s.WriteString("Projects:\n")
			for _, proj := range mi.Projects {
				s.WriteString(fmt.Sprintf("  • %s\n", proj))
			}
		}
	} else {
		// Показываем общую статистику
		totalIncome := 0.0
		for _, mi := range ic.monthlyIncomes {
			totalIncome += mi.Income
		}
		s.WriteString(fmt.Sprintf("Total Income: $%.2f\n", totalIncome))
		s.WriteString(fmt.Sprintf("Average Monthly Income: $%.2f\n", totalIncome/12))
	}

	return ic.style.Render(s.String())
} 