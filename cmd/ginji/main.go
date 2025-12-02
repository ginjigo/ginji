package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle.Copy()
	noStyle      = lipgloss.NewStyle()
	helpStyle    = blurredStyle.Copy()
	cursorMode   = cursor.CursorBlink

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Bold(true)
)

type model struct {
	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode

	// Selections
	dbIndex     int
	ormIndex    int
	mwIndices   map[int]bool
	deployIndex int
	testsIndex  int // 0: Yes, 1: No

	// State
	step int // 0: Name, 1: DB, 2: ORM, 3: Middleware, 4: Deployment, 5: Tests, 6: Done
}

func initialModel(initialName string) model {
	m := model{
		inputs:    make([]textinput.Model, 1),
		mwIndices: make(map[int]bool),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Project Name"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
			if initialName != "" {
				t.SetValue(initialName)
			}
		}

		m.inputs[i] = t
	}

	if initialName != "" {
		m.step = 1
	}

	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			if m.step == 0 {
				if m.inputs[0].Value() != "" {
					m.step++
				}
			} else if m.step == 1 {
				m.step++
			} else if m.step == 2 {
				m.step++
			} else if m.step == 3 {
				m.step++
			} else if m.step == 4 {
				m.step++
			} else if m.step == 5 {
				// Generate Project
				opts := ProjectOptions{
					Name:        m.inputs[0].Value(),
					Database:    dbOptions[m.dbIndex],
					ORM:         ormOptions[m.ormIndex],
					Middlewares: getSelectedMiddlewares(m.mwIndices),
					Deployment:  deployOptions[m.deployIndex],
					Tests:       testOptions[m.testsIndex] == "Yes",
				}
				if err := generateProject(opts); err != nil {
					fmt.Printf("Error generating project: %v\n", err)
					return m, tea.Quit
				}
				m.step++
				return m, nil
			} else if m.step == 6 {
				return m, tea.Quit
			}

		case "up", "k":
			if m.step == 1 {
				if m.dbIndex > 0 {
					m.dbIndex--
				}
			} else if m.step == 2 {
				if m.ormIndex > 0 {
					m.ormIndex--
				}
			} else if m.step == 3 {
				if m.focusIndex > 0 {
					m.focusIndex--
				}
			} else if m.step == 4 {
				if m.deployIndex > 0 {
					m.deployIndex--
				}
			} else if m.step == 5 {
				if m.testsIndex > 0 {
					m.testsIndex--
				}
			}

		case "down", "j":
			if m.step == 1 {
				if m.dbIndex < len(dbOptions)-1 {
					m.dbIndex++
				}
			} else if m.step == 2 {
				if m.ormIndex < len(ormOptions)-1 {
					m.ormIndex++
				}
			} else if m.step == 3 {
				if m.focusIndex < len(mwOptions)-1 {
					m.focusIndex++
				}
			} else if m.step == 4 {
				if m.deployIndex < len(deployOptions)-1 {
					m.deployIndex++
				}
			} else if m.step == 5 {
				if m.testsIndex < len(testOptions)-1 {
					m.testsIndex++
				}
			}

		case " ":
			if m.step == 3 {
				if m.mwIndices[m.focusIndex] {
					delete(m.mwIndices, m.focusIndex)
				} else {
					m.mwIndices[m.focusIndex] = true
				}
			}
		}
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func getSelectedMiddlewares(indices map[int]bool) []string {
	var mws []string
	for i, selected := range indices {
		if selected {
			mws = append(mws, mwOptions[i])
		}
	}
	return mws
}

func (m model) View() string {
	var s string

	if m.step == 0 {
		s = titleStyle.Render("Create New Ginji Project") + "\n\n"
		s += m.inputs[0].View() + "\n\n"
		s += helpStyle.Render("Enter project name")
	} else if m.step == 1 {
		s = titleStyle.Render("Select Database") + "\n\n"
		for i, choice := range dbOptions {
			cursor := " "
			if m.dbIndex == i {
				cursor = ">"
				choice = focusedStyle.Render(choice)
			}
			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}
	} else if m.step == 2 {
		s = titleStyle.Render("Select ORM") + "\n\n"
		for i, choice := range ormOptions {
			cursor := " "
			if m.ormIndex == i {
				cursor = ">"
				choice = focusedStyle.Render(choice)
			}
			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}
	} else if m.step == 3 {
		s = titleStyle.Render("Select Middleware (Space to select)") + "\n\n"
		for i, choice := range mwOptions {
			cursor := " "
			if m.focusIndex == i {
				cursor = ">"
			}
			checked := "[ ]"
			if m.mwIndices[i] {
				checked = "[x]"
			}
			if m.focusIndex == i {
				choice = focusedStyle.Render(choice)
			}
			s += fmt.Sprintf("%s %s %s\n", cursor, checked, choice)
		}
	} else if m.step == 4 {
		s = titleStyle.Render("Select Deployment") + "\n\n"
		for i, choice := range deployOptions {
			cursor := " "
			if m.deployIndex == i {
				cursor = ">"
				choice = focusedStyle.Render(choice)
			}
			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}
	} else if m.step == 5 {
		s = titleStyle.Render("Generate Tests?") + "\n\n"
		for i, choice := range testOptions {
			cursor := " "
			if m.testsIndex == i {
				cursor = ">"
				choice = focusedStyle.Render(choice)
			}
			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}
	} else {
		s = titleStyle.Render("Project Created Successfully!") + "\n\n"
		s += fmt.Sprintf("cd %s\n", m.inputs[0].Value())
		s += "go mod tidy\n"
		s += "go run cmd/server/main.go\n"
	}

	return s
}

var dbOptions = []string{"None", "SQLite", "PostgreSQL", "MySQL"}
var ormOptions = []string{"None", "GORM"}
var mwOptions = []string{"Logger", "Recovery", "CORS"}
var deployOptions = []string{"None", "Docker"}
var testOptions = []string{"Yes", "No"}

func main() {
	initialName := ""
	if len(os.Args) > 2 && os.Args[1] == "new" {
		initialName = os.Args[2]
	}

	p := tea.NewProgram(initialModel(initialName))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
