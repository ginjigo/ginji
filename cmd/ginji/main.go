package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor   = lipgloss.Color("#7D56F4")
	secondaryColor = lipgloss.Color("#FAFAFA")
	subtleColor    = lipgloss.Color("241")
	successColor   = lipgloss.Color("#04B575")
	errorColor     = lipgloss.Color("#FF0000")

	// Styles
	focusedStyle = lipgloss.NewStyle().Foreground(primaryColor)
	blurredStyle = lipgloss.NewStyle().Foreground(subtleColor)
	cursorStyle  = focusedStyle
	helpStyle    = blurredStyle.Italic(true)

	titleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Background(primaryColor).
			Padding(0, 1).
			Bold(true).
			MarginBottom(1)

	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(primaryColor)
	checkMark         = lipgloss.NewStyle().Foreground(successColor).SetString("✓")
	crossMark         = lipgloss.NewStyle().Foreground(errorColor).SetString("✗")
)

type tickMsg time.Time

type model struct {
	focusIndex int
	inputs     []textinput.Model

	// Components
	spinner  spinner.Model
	progress progress.Model

	// Selections
	dbIndex     int
	ormIndex    int
	mwIndices   map[int]bool
	deployIndex int
	testsIndex  int // 0: Yes, 1: No

	// State
	step         int // 0: Name, 1: DB, 2: ORM, 3: Middleware, 4: Deployment, 5: Tests, 6: Generating, 7: Done
	err          error
	progressPct  float64
	generateDone bool
}

func initialModel(initialName string) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(primaryColor)

	p := progress.New(progress.WithDefaultGradient())

	m := model{
		inputs:    make([]textinput.Model, 1),
		mwIndices: make(map[int]bool),
		spinner:   s,
		progress:  p,
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
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.step != 6 { // Disable input during generation
			switch msg.String() {
			case "ctrl+c", "esc":
				return m, tea.Quit

			case "enter":
				switch m.step {
				case 0:
					if m.inputs[0].Value() != "" {
						m.step++
					}
				case 1, 2, 3, 4:
					m.step++
				case 5:
					// Start Generation
					m.step++
					return m, tea.Batch(
						tickCmd(),
						func() tea.Msg {
							projectName := strings.TrimSpace(m.inputs[0].Value())
							opts := ProjectOptions{
								Name:        projectName,
								Database:    dbOptions[m.dbIndex],
								ORM:         ormOptions[m.ormIndex],
								Middlewares: getSelectedMiddlewares(m.mwIndices),
								Deployment:  deployOptions[m.deployIndex],
								Tests:       testOptions[m.testsIndex] == "Yes",
							}
							if err := generateProject(opts); err != nil {
								return err
							}
							return true // Success
						},
					)
				case 7:
					return m, tea.Quit
				}

			case "up", "k":
				switch m.step {
				case 1:
					if m.dbIndex > 0 {
						m.dbIndex--
					}
				case 2:
					if m.ormIndex > 0 {
						m.ormIndex--
					}
				case 3:
					if m.focusIndex > 0 {
						m.focusIndex--
					}
				case 4:
					if m.deployIndex > 0 {
						m.deployIndex--
					}
				case 5:
					if m.testsIndex > 0 {
						m.testsIndex--
					}
				}

			case "down", "j":
				switch m.step {
				case 1:
					if m.dbIndex < len(dbOptions)-1 {
						m.dbIndex++
					}
				case 2:
					if m.ormIndex < len(ormOptions)-1 {
						m.ormIndex++
					}
				case 3:
					if m.focusIndex < len(mwOptions)-1 {
						m.focusIndex++
					}
				case 4:
					if m.deployIndex < len(deployOptions)-1 {
						m.deployIndex++
					}
				case 5:
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

	case tickMsg:
		if m.step == 6 && !m.generateDone {
			m.progressPct += 0.1
			if m.progressPct > 1.0 {
				m.progressPct = 1.0
			}
			return m, tea.Batch(m.progress.SetPercent(m.progressPct), tickCmd())
		}

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case error:
		m.err = msg
		m.step = 7
		return m, nil

	case bool: // Generation Success
		m.generateDone = true
		m.progressPct = 1.0
		m.step = 7
		return m, m.progress.SetPercent(1.0)
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
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

	if m.err != nil {
		return fmt.Sprintf("\n%s Error: %v\n\nPress Enter to quit", crossMark, m.err)
	}

	switch m.step {
	case 0:
		s = titleStyle.Render("Create New Ginji Project") + "\n\n"
		s += m.inputs[0].View() + "\n\n"
		s += helpStyle.Render("Enter project name")
	case 1:
		s = titleStyle.Render("Select Database") + "\n\n"
		for i, choice := range dbOptions {
			cursor := " "
			style := itemStyle
			if m.dbIndex == i {
				cursor = ">"
				style = selectedItemStyle
			}
			s += fmt.Sprintf("%s %s\n", cursor, style.Render(choice))
		}
	case 2:
		s = titleStyle.Render("Select ORM") + "\n\n"
		for i, choice := range ormOptions {
			cursor := " "
			style := itemStyle
			if m.ormIndex == i {
				cursor = ">"
				style = selectedItemStyle
			}
			s += fmt.Sprintf("%s %s\n", cursor, style.Render(choice))
		}
	case 3:
		s = titleStyle.Render("Select Middleware (Space to select)") + "\n\n"
		for i, choice := range mwOptions {
			cursor := " "
			style := itemStyle
			if m.focusIndex == i {
				cursor = ">"
				style = selectedItemStyle
			}
			checked := "[ ]"
			if m.mwIndices[i] {
				checked = "[x]"
			}
			s += fmt.Sprintf("%s %s %s\n", cursor, checked, style.Render(choice))
		}
	case 4:
		s = titleStyle.Render("Select Deployment") + "\n\n"
		for i, choice := range deployOptions {
			cursor := " "
			style := itemStyle
			if m.deployIndex == i {
				cursor = ">"
				style = selectedItemStyle
			}
			s += fmt.Sprintf("%s %s\n", cursor, style.Render(choice))
		}
	case 5:
		s = titleStyle.Render("Generate Tests?") + "\n\n"
		for i, choice := range testOptions {
			cursor := " "
			style := itemStyle
			if m.testsIndex == i {
				cursor = ">"
				style = selectedItemStyle
			}
			s += fmt.Sprintf("%s %s\n", cursor, style.Render(choice))
		}
	case 6:
		s = titleStyle.Render("Generating Project...") + "\n\n"
		s += fmt.Sprintf("%s Downloading templates and configuring project...\n\n", m.spinner.View())
		s += m.progress.View() + "\n"
	case 7:
		s = titleStyle.Render("Project Created Successfully!") + "\n\n"
		s += fmt.Sprintf("%s Project ready at %s\n\n", checkMark, strings.TrimSpace(m.inputs[0].Value()))
		s += lipgloss.NewStyle().Foreground(subtleColor).Render("Next steps:") + "\n"
		s += fmt.Sprintf("  cd %s\n", strings.TrimSpace(m.inputs[0].Value()))
		s += "  go mod tidy\n"
		s += "  go run cmd/server/main.go\n\n"
		s += helpStyle.Render("Press Enter to quit")
	}

	return s
}

var dbOptions = []string{"None", "SQLite", "PostgreSQL", "MySQL"}
var ormOptions = []string{"None", "GORM", "sqlc", "ent"}
var mwOptions = []string{"Logger", "Recovery", "CORS"}
var deployOptions = []string{"None", "Docker"}
var testOptions = []string{"Yes", "No"}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "new":
		initialName := ""
		if len(os.Args) > 2 {
			initialName = os.Args[2]
		}
		p := tea.NewProgram(initialModel(initialName))
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}

	case "generate", "g":
		if len(os.Args) < 3 {
			fmt.Println("Usage: ginji generate [handler|middleware|crud] <name>")
			os.Exit(1)
		}

		subCommand := os.Args[2]

		switch subCommand {
		case "handler", "h":
			if len(os.Args) < 4 {
				fmt.Println("Usage: ginji generate handler <name> [--json]")
				os.Exit(1)
			}
			name := os.Args[3]
			useJSON := len(os.Args) > 4 && os.Args[4] == "--json"

			if err := generateHandler(name, useJSON); err != nil {
				fmt.Printf("Error generating handler: %v\n", err)
				os.Exit(1)
			}

		case "middleware", "mw":
			if len(os.Args) < 4 {
				fmt.Println("Usage: ginji generate middleware <name>")
				os.Exit(1)
			}
			name := os.Args[3]

			if err := generateMiddleware(name); err != nil {
				fmt.Printf("Error generating middleware: %v\n", err)
				os.Exit(1)
			}

		case "crud":
			if len(os.Args) < 4 {
				fmt.Println("Usage: ginji generate crud <resource>")
				os.Exit(1)
			}
			resource := os.Args[3]

			if err := generateCRUD(resource); err != nil {
				fmt.Printf("Error generating CRUD: %v\n", err)
				os.Exit(1)
			}

		default:
			fmt.Printf("Unknown generate command: %s\n", subCommand)
			fmt.Println("Available: handler, middleware, crud")
			os.Exit(1)
		}

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Ginji CLI - Ultra-fast Go API framework

Usage:
  ginji new <project-name>              Create a new Ginji project
  ginji generate <type> <name>          Generate code
  
Generate Commands:
  ginji generate handler <name>         Generate a handler
  ginji generate handler <name> --json  Generate a handler with JSON binding
  ginji generate middleware <name>      Generate middleware
  ginji generate crud <resource>        Generate full CRUD handlers
  
Aliases:
  g, generate                           Code generation
  h, handler                            Handler generation
  mw, middleware                        Middleware generation
  
Examples:
  ginji new my-api
  ginji g handler GetUser --json
  ginji g middleware Auth
  ginji g crud User

For more information, visit: https://github.com/ginjigo/ginji`)
}
