// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tuistatusbar struct {
	render *lipgloss.Renderer
	width  int
	style  lipgloss.Style
}

func newTuiStatusbar(r *lipgloss.Renderer) tuistatusbar {
	return tuistatusbar{
		render: r,
	}
}

func (s tuistatusbar) Init() tea.Cmd {
	return nil
}

func (s tuistatusbar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
	}
	return s, nil
}

func (s tuistatusbar) View() string {
	t1 := "statusbar"
	t2 := time.Now().Format("15:04")

	s1 := s.render.NewStyle().Align(lipgloss.Left).Background(lipgloss.Color("#FF5F87"))
	s2 := s.render.NewStyle().
		Width(s.width - lipgloss.Width(t1)).
		Align(lipgloss.Right).
		Background(lipgloss.Color("10"))

	return s.style.Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			s1.Render(t1),
			s2.Render(t2),
		),
	)
}
