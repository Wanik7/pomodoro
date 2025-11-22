package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	workMode mode = iota
	breakMode

	persist_path = "persist.json"
)

type mode int

type task struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Done bool   `json:"done"`
}

type loadMsg struct {
	tasks  []task
	nextID int
	err    error
}

type persistMsg struct {
	err error
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

	err error
}

func loadCmd(path string) tea.Cmd {
	return func() tea.Msg {
		f, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				return loadMsg{tasks: nil, nextID: 1, err: nil}
			}
			return loadMsg{tasks: nil, nextID: 1, err: err}
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(f)

		var snapshot struct {
			Tasks  []task `json:"tasks"`
			NextID int    `json:"next_id"`
		}
		if err := json.NewDecoder(f).Decode(&snapshot); err != nil {
			return loadMsg{tasks: nil, nextID: 1, err: err}
		}
		next := snapshot.NextID
		if next <= 0 {
			maxID := 0
			for _, task := range snapshot.Tasks {
				if task.ID > maxID {
					maxID = task.ID
				}
			}
			next = maxID + 1
		}
		return loadMsg{tasks: snapshot.Tasks, nextID: next, err: nil}
	}
}

func saveCmd(path string, tasks []task, nextID int) tea.Cmd {
	snapshot := struct {
		Tasks  []task `json:"tasks"`
		NextID int    `json:"next_id"`
	}{
		Tasks:  tasks,
		NextID: nextID,
	}

	return func() tea.Msg {
		f, err := os.Create(path)
		if err != nil {
			return loadMsg{tasks: nil, nextID: 1, err: err}
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(f)

		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		if err := enc.Encode(snapshot); err != nil {
			return loadMsg{tasks: nil, nextID: 1, err: err}
		}
		return persistMsg{err: nil}
	}
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

func (m model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), loadCmd(persist_path))
}

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
					m.tasks = append(m.tasks, task{ID: m.nextID, Name: val})
					m.nextID++
					cmd := saveCmd(persist_path, m.tasks, m.nextID)
					m.input.Reset()
					m.adding = false
					return m, cmd
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

	switch msg := msg.(type) {
	case loadMsg:
		if msg.err != nil {
			m.err = fmt.Errorf("load failed: %v", msg.err)
			return m, nil
		}
		if msg.tasks != nil {
			m.tasks = msg.tasks
			m.nextID = msg.nextID
		}
		return m, nil

	case persistMsg:
		if msg.err != nil {
			m.err = fmt.Errorf("save failed: %v", msg.err)
		} else {
			m.err = nil
		}
		return m, nil

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
			if task.Done {
				doneString = "[X]"
			} else {
				doneString = "[ ]"
			}
			tasks += fmt.Sprintf("ID: %d | title: %s | done: %s", task.ID, task.Name, doneString) + "\n"
		}
	}

	var addBlock string
	if m.adding {
		addBlock = "\nadd a task:\n" + m.input.View() + "\n(Enter to add, Esc to cancel)\n"
	} else {
		addBlock = "\nPress a to add a new task\n"
	}

	var errBlock string
	if m.err != nil {
		errBlock = fmt.Sprintf("\nError: %v\n", m.err)
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
		return fmt.Sprintf("Pomodoro\n\n"+
			"Mode: %s"+
			"\nTime: %02d:%02d (%s)"+
			"\n\nTasks:\n"+
			"%s%sKeys: 'a' add | 'p' pause | 'r' reset | 'm' mode | 'q' quit\n"+
			"%s",
			modeStr, mm, ss, status, tasks, addBlock, errBlock)
	}
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
