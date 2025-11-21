package main

import (
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	workMode mode = iota
	breakMode
)

type mode int

type task struct {
	ID   int
	name string
	done bool
}

type tickMsg struct{}

type model struct {
	mode          mode
	workDuration  int
	breakDuration int
	secondsLeft   int
	running       bool
	quitting      bool

	tasks []task

	nextID int
	adding bool
	input  textinput.Model
}

func initialModel() model {
	const workTime int = 15
	const breakTime int = 5

	ti := textinput.New()
	ti.Placeholder = "input something"
	ti.CharLimit = 200
	ti.Width = 40

	tasks := []task{
		{1, "do anything", true},
		{2, "do stage 5", false},
	}
	return model{
		mode:          workMode,
		workDuration:  workTime,
		breakDuration: breakTime,
		secondsLeft:   workTime,
		running:       true,
		quitting:      false,
		tasks:         tasks,
		nextID:        3,
		adding:        false,
		input:         ti,
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m model) currentDuration() int {
	if m.mode == workMode {
		return m.workDuration
	}
	return m.breakDuration
}

func (m *model) switchMode() {
	if m == nil {
		log.Fatal("model is nil")
	}
	if m.mode == workMode {
		m.mode = breakMode
		m.secondsLeft = m.breakDuration
	} else {
		m.mode = workMode
		m.secondsLeft = m.workDuration
	}
}

func (m model) Init() tea.Cmd { return tickCmd() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tickMsg:
		if m.running && m.secondsLeft > 0 {
			m.secondsLeft--
			if m.secondsLeft == 0 {
				m.switchMode()
			}
		}
		return m, tickCmd()
	}

	if m.adding {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "enter":
				val := m.input.Value()
				if val != "" {
					m.tasks = append(m.tasks, task{ID: m.nextID, name: val})
					m.nextID++
				}
				m.input.Reset()
				m.adding = false
				return m, nil
			case "esc":
				m.input.Reset()
				m.adding = false
				return m, nil
			case "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			}
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	// 3. Обычный режим
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "p":
			m.running = !m.running
			return m, nil
		case "r":
			m.secondsLeft = m.currentDuration()
			m.running = true
			return m, nil
		case "m":
			m.switchMode()
			m.running = true
			return m, nil
		case "a":
			m.adding = true
			m.input.Focus()
			return m, nil
		}
	}

	return m, nil
}

func (m model) View() string {
	var tasks string
	if len(m.tasks) == 0 {
		tasks = "there is no tasks"
	} else {
		for _, task := range m.tasks {
			doneString := ""
			if task.done {
				doneString = "[X]"
			} else {
				doneString = "[ ]"
			}
			tasks += fmt.Sprintf("ID: %d | title: %s | done: %s", task.ID, task.name, doneString) + "\n     "
		}
	}

	var addBlock string
	if m.adding {
		addBlock = "\nadd a task:\n" + m.input.View() + "\n(Enter to add, Esc to cancel)\n"
	} else {
		addBlock = "\nPress a to add a new task\n"
	}

	mm := m.secondsLeft / 60
	ss := m.secondsLeft % 60
	status := "running"
	modeStr := "work"
	if m.mode == breakMode {
		modeStr = "break"
	}
	if !m.running {
		status = "not running"
	}
	if m.quitting {
		return "Bye!"
	} else {
		return fmt.Sprintf("\n		  Pomodoro\n\n"+
			"      Mode: %s  Time left: %d%d (%s)\n\n"+
			"     %s\n\n"+
			"	  %s\n\n"+
			"Press 'p' to pause,"+
			" 'a' to new task, 'r' to reset, 'm' to change mode\n\n"+
			"	  Press 'q'/'ctrl+c' to quit",
			modeStr, mm, ss, status, tasks, addBlock)
	}
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
