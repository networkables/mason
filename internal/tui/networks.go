// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type TuiNetworksModel struct {
	tab table.Model
}

func newTuiNetworksModel() TuiNetworksModel {
	columns := []table.Column{
		{Title: "Name", Width: 26},
		{Title: "Block", Width: 20},
		{Title: "External IP", Width: 20},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		// table.WithHeight(7),
	)
	t.FromValues("A\tB\tC\nD\tE\tF\n", "\t")

	s := table.DefaultStyles()
	s.Header = s.Header.BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return TuiNetworksModel{tab: t}
}

func (nm TuiNetworksModel) Init() tea.Cmd {
	log.Debug("init")
	return nil
}

func (nm TuiNetworksModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if nm.tab.Focused() {
				nm.tab.Blur()
			} else {
				nm.tab.Focus()
			}
		case tea.KeyEnter:
			log.Debug(nm.tab.SelectedRow()[1])
			return nm, tea.Batch(
				tea.Printf("Lets go to %s", nm.tab.SelectedRow()[1]),
			)
		case tea.KeyCtrlC:
			return nm, tea.Quit
		}
	}
	nm.tab, cmd = nm.tab.Update(msg)
	return nm, cmd
}

func (nm TuiNetworksModel) View() string {
	// s := "There are %d customers\n"
	// return fmt.Sprintf(s, len(cm.Customers))
	return baseStyle.Render(nm.tab.View()) + "\n"
}
