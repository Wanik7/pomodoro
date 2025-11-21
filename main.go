package main

import (
	"fmt"
	"log"
	"time"

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
}

func initialModel() model {
	const workTime int = 15
	const breakTime int = 5

	tasks := []task{
		task{1, "do anything", true},
		task{2, "do stage 5", false},
	}
	return model{
		mode:          workMode,
		workDuration:  workTime,
		breakDuration: breakTime,
		secondsLeft:   workTime,
		running:       true,
		quitting:      false,

		tasks: tasks,
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
	switch msg := msg.(type) {
	case tickMsg:
		if m.running && m.secondsLeft > 0 {
			m.secondsLeft--
			if m.secondsLeft == 0 {
				m.switchMode()
			}
		}
		return m, tickCmd()

	case tea.KeyMsg:

		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "p":
			if m.secondsLeft > 0 {
				m.running = !m.running
			}
			return m, nil

		case "r":
			m.secondsLeft = m.currentDuration()
			m.running = true
			return m, nil

		case "m":
			m.switchMode()
			m.running = true
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	var tasks string
	for _, task := range m.tasks {
		doneString := ""
		if task.done {
			doneString = "[X]"
		} else {
			doneString = "[ ]"
		}
		tasks += fmt.Sprintf("ID: %d | title: %s | done: %s", task.ID, task.name, doneString) + "\n     "
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
			"Press 'p' to pause,"+
			" 'r' to reset, 'm' to change mode\n\n"+
			"	  Press 'q'/'ctrl+c' to quit",
			modeStr, mm, ss, status, tasks)
	}
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
