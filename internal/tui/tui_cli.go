// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"github.com/networkables/mason/internal/network"
	"github.com/networkables/mason/internal/server"
)

type tuicli struct {
	render       *lipgloss.Renderer
	height       int
	width        int
	input        textinput.Model
	cmds         []string
	mason        server.MasonReaderWriter
	adminEnabled bool
}

const (
	defaultPrompt = "> "
	adminPrompt   = "admin> "
)

var (
	defaultPromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	adminPromptStyle   = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#770000"))
)

func newTuiCLI(m server.MasonReaderWriter, r *lipgloss.Renderer) *tuicli {
	ti := textinput.New()
	ti.Prompt = defaultPrompt
	ti.PromptStyle = defaultPromptStyle
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 20
	ti.ShowSuggestions = true

	c := make([]string, 0, 1000)
	c = append(c, "Mason v.0.1")
	c = append(c, "Networkables, LLLP.")
	c = append(c, " ")

	return &tuicli{
		render: r,
		input:  ti,
		cmds:   c,
		width:  25,
		height: 25,
		mason:  m,
	}
}

func (c tuicli) Init() tea.Cmd {
	return textinput.Blink
}

func (c tuicli) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd1, cmd2 tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.height = msg.Height - 1
		c.width = msg.Width
		c.input.Width = c.width
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			cmdstr := c.input.Value()
			c.cmds, cmd1 = c.processCmd(cmdstr)
			c.input.SetValue("")
		}
	case SwitchToAdminModeMsg:
		c = c.setAdminPrompt()
	case SwitchToDefaultModeMsg:
		c = c.setDefaultPrompt()
	}
	c.input, cmd2 = c.input.Update(msg)
	return c, tea.Batch(cmd1, cmd2)
}

type (
	SwitchToAdminModeMsg   struct{}
	SwitchToDefaultModeMsg struct{}
)

func SwitchToAdminMode() tea.Msg {
	return SwitchToAdminModeMsg{}
}

func SwitchToDefaultMode() tea.Msg {
	return SwitchToDefaultModeMsg{}
}

/*
- Need to figure out how the l1 style was adding somewhat random spaces in from of text
- Ideally would like to be able to color the command issued a different color from the command output
  - This would mean the cmds would be a slice of X struct { str string, kind ENUMIOTA }
*/
func (c tuicli) View() string {
	s1 := c.render.NewStyle().Height(c.height).Width(c.width)
	// l1 := c.render.NewStyle().Foreground(lipgloss.Color("#00AACC")).Background(lipgloss.Color("#444444"))
	ls := ""
	for _, l := range c.cmds {
		ls += l + "\n"
		// ls += l1.Render(l + "\n")
	}
	return s1.Render(lipgloss.JoinVertical(
		lipgloss.Top,
		ls,
		c.input.View(),
	))
}

const (
	cmdQuit         = "quit"
	cmdExit         = "exit"
	cmdAdmin        = "admin"
	cmdShow         = "show"
	cmdShowNetworks = "networks"
	cmdShowDevices  = "devices"
	cmdAdd          = "add"
	cmdAddNetwork   = "network"
)

func (c tuicli) setDefaultPrompt() tuicli {
	c.input.Prompt = defaultPrompt
	c.input.PromptStyle = defaultPromptStyle
	c.adminEnabled = true
	return c
}

func (c tuicli) setAdminPrompt() tuicli {
	c.input.Prompt = adminPrompt
	c.input.PromptStyle = adminPromptStyle
	c.adminEnabled = true
	return c
}

func (c tuicli) processCmd(s string) ([]string, tea.Cmd) {
	log.Debug("cmd: " + s)
	var cmd tea.Cmd
	cmdstr := s
	args := strings.Split(s, " ")
	if len(args) > 0 {
		cmdstr = args[0]
	}
	switch cmdstr {
	case cmdQuit:
		cmd = tea.Quit
	case cmdExit:
		cmd = tea.Quit
		if c.adminEnabled {
			cmd = SwitchToDefaultMode
		}
	case cmdAdmin:
		cmd = SwitchToAdminMode
	case cmdShow:
		if len(args) > 1 {
			showCmd := args[1]
			switch showCmd {
			case cmdShowNetworks:
				nets := c.mason.ListNetworks()
				lines := make([]string, 0, len(nets))
				for _, net := range nets {
					lines = append(lines, net.String())
				}
				s += "\n" + strings.Join(lines, "\n")
			case cmdShowDevices:
				devs := c.mason.ListDevices()
				lines := make([]string, 0, len(devs))
				for _, dev := range devs {
					lines = append(lines, dev.String())
				}
				s += "\n" + strings.Join(lines, "\n")
			default:
				s += "\n" + "unknown show command: " + showCmd
			}
		} else {
			s += "\n" + "show networks|devices"
		}
	case cmdAdd:
		if len(args) > 1 {
			addCmd := args[1]
			switch addCmd {
			case cmdAddNetwork:
				if len(args) > 2 {
					prefix := args[2]
					newnet, err := network.New(prefix, prefix)
					if err != nil {
						s += "\n" + "error creating network: " + err.Error()
					} else {
						err := c.mason.AddNetwork(newnet)
						if err != nil {
							s += "\n" + "error creating network: " + err.Error()
						} else {
							s += "\n" + "network added"
						}
					}
				}
			}
		} else {
			s += "\n" + "add network"
		}
	case "":

	default:
		s += "\n" + "Invalid command"
	}
	s += "\n"
	return append(c.cmds, s), cmd
}
