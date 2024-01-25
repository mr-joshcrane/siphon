package main

import (
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return string(i) }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	content  map[list.Item]string
	ready    bool
	viewport viewport.Model
	list     list.Model
	V        bool
	L        bool
}

func items(m map[list.Item]string) []list.Item {
	var items []list.Item
	for k, _ := range m {
		items = append(items, k)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].(item) < items[j].(item)
	})
	return items
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "down", "j":
			m.list.CursorDown()
		case "up", "k":
			m.list.CursorUp()
		case "\\":
			m.L = true
			m.list = list.New(
				items(m.content),
				itemDelegate{},
				20,
				14,
			)
		case "enter":
			m.V = true
			m.viewport = viewport.Model{}
			i := m.list.SelectedItem()
			text := m.content[i]
			m.viewport.SetContent(text)
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.V {
		return m.viewport.View()
	}

	return m.list.View()
}

func randomGuids() []string {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Define a random size for the list (1-10)
	size := rand.Intn(10) + 1

	// Create a slice to hold the random GUIDs
	guids := make([]string, size)

	// Populate the slice with random GUIDs
	for i := 0; i < size; i++ {
		guids[i] = uuid.New().String()
	}
	return guids
}

func randomMap() map[string]string {
	m := make(map[string]string)
	for _, guid := range randomGuids() {
		m[guid] = guid
	}
	return m
}

func main() {
	contents := make(map[list.Item]string)
	// /Users/josh.crane/.local/state/assistant/
	localfs := os.DirFS("/Users/josh.crane/.local/state/assistant/audit_logs/")
	entries, err := fs.ReadDir(localfs, ".")
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := fs.ReadFile(localfs, entry.Name())
		if err != nil {
			panic(err)
		}
		contents[item(entry.Name())] = string(data)
	}

	l := list.New(
		items(contents),
		itemDelegate{},
		20,
		14,
	)
	l.Title = titleStyle.Render("Big ol list")
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(true)
	p := tea.NewProgram(
		model{
			content: contents,
			list:    l,
		},
	)
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
