package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// MenuItem represents an option in the menu.
type MenuItem struct {
	Title string
	Key   string
}

// MenuModel is a bubbletea model for arrow-key menu navigation.
type MenuModel struct {
	Title    string
	Items    []MenuItem
	cursor   int
	chosen   bool
	quitting bool
}

// NewMenuModel creates a new menu model.
func NewMenuModel(title string, items []MenuItem) MenuModel {
	return MenuModel{
		Title: title,
		Items: items,
	}
}

// Selected returns the index of the selected item, or -1 if quit.
func (m MenuModel) Selected() int {
	if m.quitting {
		return -1
	}
	return m.cursor
}

// Chosen returns true if the user made a selection.
func (m MenuModel) Chosen() bool {
	return m.chosen
}

func (m MenuModel) Init() tea.Cmd {
	return nil
}

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.Items)-1 {
				m.cursor++
			}
		case "enter":
			m.chosen = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m MenuModel) View() string {
	s := "\n"
	s += Bold.Render(m.Title) + "\n\n"

	for i, item := range m.Items {
		cursor := MenuNoCursor.String()
		style := MenuUnselected
		if m.cursor == i {
			cursor = MenuCursor.String()
			style = MenuSelected
		}
		s += fmt.Sprintf("%s%s\n", cursor, style.Render(item.Title))
	}

	s += "\n" + Subtle.Render("↑/↓ to navigate • enter to select • q to quit") + "\n"
	return s
}
