// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/bubbletea"
)

type Tui struct {
	m          MasonReaderWriter
	mainwindow tea.Model
	statusbar  tea.Model
	render     *lipgloss.Renderer
}

func New(m MasonReaderWriter) *Tui {
	return &Tui{
		m: m,
	}
}

func (t *Tui) TeaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, active := s.Pty()
	if !active {
		wish.Fatalln(s, "no active terminal, skipping")
		return nil, nil
	}
	log.Info(pty.Term, pty.Window)
	t.render = bubbletea.MakeRenderer(s)
	t.mainwindow = newTuiCLI(t.m, t.render)
	t.statusbar = newTuiStatusbar(t.render)
	return t, []tea.ProgramOption{tea.WithAltScreen()}
}

func (t Tui) Init() tea.Cmd {
	return nil
}

func (t *Tui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return t, tea.Quit
		}
	}
	var maincmd, statuscmd tea.Cmd
	t.mainwindow, maincmd = t.mainwindow.Update(msg)
	stat, statuscmd := t.statusbar.Update(msg)
	t.statusbar = stat.(tuistatusbar)
	return t, tea.Batch(maincmd, statuscmd)
}

func (t Tui) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		t.mainwindow.View(),
		t.statusbar.View(),
	)
}
