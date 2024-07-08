// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/networkables/mason/internal/model"
	"github.com/networkables/mason/internal/server"
	"github.com/networkables/mason/internal/sqlitestore"
)

var (
	cmdTool = &cobra.Command{
		Use:   "tool",
		Short: "network tools",
		PersistentPreRunE: func(*cobra.Command, []string) error {
			cfg := server.GetConfig()
			if !server.HasCapabilities(cfg) {
				return errors.New(
					"capabilities required based on config, but not granted to the executable",
				)
			}
			return nil
		},
	}

	cmdToolPing = &cobra.Command{
		Use:   "ping [target]",
		Short: "icmp ping the target",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmdPing(args)
		},
	}

	cmdToolArpPing = &cobra.Command{
		Use:   "arpping [target]",
		Short: "send arp request packet for target",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmdArpPing(args)
		},
	}

	cmdToolPortScan = &cobra.Command{
		Use:   "portscan [target]",
		Short: "scan for open ports of the target",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmdToolPortScan(args)
		},
	}

	cmdToolExternalIP = &cobra.Command{
		Use:   "externalip",
		Short: "discover the external ip of the instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmdToolExternalIP(args)
		},
	}

	cmdToolTraceroute = &cobra.Command{
		Use:   "traceroute [target]",
		Short: "discover hops between mason and the target",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmdToolTraceroute(args)
		},
	}

	cmdToolTLS = &cobra.Command{
		Use:   "tls [target]",
		Short: "show tls information",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmdToolTLS(args)
		},
	}

	cmdToolSNMP = &cobra.Command{
		Use:   "snmp [target]",
		Short: "show snmp for a target",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmdToolSNMP(args)
		},
	}

	cmdToolCheckDNS = &cobra.Command{
		Use:   "dns [target]",
		Short: "show all type A DNS records for target",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmdToolCheckDNS(args)
		},
	}
)

func init() {
	cmdTool.AddCommand(
		cmdToolPing,
		cmdToolArpPing,
		cmdToolPortScan,
		cmdToolExternalIP,
		cmdToolTraceroute,
		cmdToolTLS,
		cmdToolSNMP,
		cmdToolCheckDNS,
	)
}

func runCmdArpPing(args []string) error {
	target := args[0]
	ctx := context.Background()

	cfg := server.GetConfig()
	m := server.New(server.WithConfig(cfg))

	mac, err := m.ArpPing(ctx, target, cfg.Discovery.Arp.Timeout)
	if err != nil {
		return err
	}
	log.Info("arpping", "target", target, "mac", mac)
	return nil
}

func runCmdPing(args []string) error {
	target := args[0]

	cfg := server.GetConfig()
	m := server.New(server.WithConfig(cfg))

	stats, err := m.IcmpPing(
		context.Background(),
		target,
		cfg.Discovery.Icmp.PingCount,
		cfg.Discovery.Icmp.Timeout,
		cfg.Discovery.Icmp.Privileged,
	)
	if err != nil {
		return err
	}
	log.Info(
		"ping",
		"target",
		target,
		"count",
		stats.TotalPackets,
		"packetloss",
		stats.PacketLoss,
		"min",
		stats.Minimum,
		"mean",
		stats.Mean,
		"max",
		stats.Maximum,
		"stddev",
		stats.StdDev,
	)
	return nil
}

func runCmdToolPortScan(args []string) error {
	cfg := server.GetConfig()
	target := args[0]

	m := server.New(server.WithConfig(cfg))
	ports, err := m.Portscan(context.Background(), target, cfg.Enrichment.PortScan)
	if err != nil {
		return err
	}
	log.Info("portscan", "target", target, "openports", ports)

	return nil
}

func runCmdToolExternalIP(args []string) error {
	cfg := server.GetConfig()
	m := server.New(server.WithConfig(cfg))

	addr, err := m.GetExternalAddr(context.Background())
	if err != nil {
		return err
	}
	log.Info("external address", "ip", addr)
	return nil
}

func runCmdToolTraceroute(args []string) error {
	target := args[0]

	cfg := server.GetConfig()
	svropts := []server.Option{
		server.WithConfig(cfg),
	}
	headers := []string{"Hop", "Address", "Loss", "Min", "Max"}
	if cfg.Asn.Enabled {
		headers = append(headers, "Asn", "Org")
		sqls, err := sqlitestore.New(cfg.Store.Sqlite)
		if err != nil {
			return err
		}
		svropts = append(svropts,
			server.WithStore(sqls),
			server.WithNetflowStorer(sqls),
		)
	}
	m := server.New(svropts...)

	hops, err := m.Traceroute(context.Background(), target)
	if err != nil {
		return err
	}
	// log.Info("traceroute", "target", target)

	re := lipgloss.NewRenderer(os.Stdout)

	var (
		purple    = lipgloss.Color("99")
		gray      = lipgloss.Color("245")
		lightGray = lipgloss.Color("241")
		// HeaderStyle is the lipgloss style used for the table headers.
		HeaderStyle = re.NewStyle().Foreground(purple).Bold(true).Align(lipgloss.Center)
		// CellStyle is the base lipgloss style used for the table rows.
		// CellStyle = re.NewStyle().Padding(0, 1).Width(14)
		CellStyle = re.NewStyle().Padding(0, 1)
		// OddRowStyle is the lipgloss style used for odd-numbered table rows.
		OddRowStyle = CellStyle.Foreground(gray)
		// EvenRowStyle is the lipgloss style used for even-numbered table rows.
		EvenRowStyle = CellStyle.Foreground(lightGray)
		// BorderStyle is the lipgloss style used for the table border.
		BorderStyle = lipgloss.NewStyle().Foreground(purple)
		colstyles   = []lipgloss.Style{
			// Hop
			lipgloss.NewStyle().Width(5).Align(lipgloss.Center),
			// Address
			lipgloss.NewStyle().Width(19).Align(lipgloss.Right),
			// Loss
			lipgloss.NewStyle().Width(7).Align(lipgloss.Right),
			// Min
			lipgloss.NewStyle().Width(9).Align(lipgloss.Right),
			// Max
			lipgloss.NewStyle().Width(9).Align(lipgloss.Right),
			// Asn
			lipgloss.NewStyle().Width(7).Align(lipgloss.Center),
			// Org
			lipgloss.NewStyle().Width(50).Align(lipgloss.Left),
		}
	)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(BorderStyle).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == 0:
				return HeaderStyle
			case row%2 == 0:
				return EvenRowStyle.Inherit(colstyles[col])
			default:
				return OddRowStyle.Inherit(colstyles[col])
			}
		}).
		Headers(headers...)

	for i, hop := range hops {
		row := []string{
			strconv.Itoa(i),
			hop.Peer.String(),
			fmt.Sprintf("%.2f", hop.PacketLoss),
			hop.Maximum.Round(50 * time.Microsecond).String(),
			hop.Maximum.Round(50 * time.Microsecond).String(),
		}
		if cfg.Asn.Enabled {
			asn := m.LookupIP(model.AddrToModelAddr(hop.Peer))
			asninfo, err := m.GetAsn(context.Background(), asn)
			if err != nil {
				log.Error("getasn", "error", err)
			}
			row = append(row, asninfo.Asn, asninfo.Name)
		}
		t.Row(row...)
	}
	fmt.Println(t)
	// for i, hop := range hops {
	// 	keyvals := []interface{}{
	// 		"pos",
	// 		i,
	// 		"peer",
	// 		hop.Peer,
	// 		"count",
	// 		hop.TotalPackets,
	// 		"packetloss",
	// 		hop.PacketLoss,
	// 		"min",
	// 		hop.Minimum,
	// 		"mean",
	// 		hop.Mean,
	// 		"max",
	// 		hop.Maximum,
	// 		"stddev",
	// 		hop.StdDev,
	// 	}
	// 	if cfg.Asn.Enabled {
	// 		asn := m.LookupIP(model.AddrToModelAddr(hop.Peer))
	// 		asninfo, err := m.GetAsn(context.Background(), asn)
	// 		if err != nil {
	// 			log.Error("getasn", "error", err)
	// 		}
	// 		hop.Asn = asn
	// 		hop.OrgName = asninfo.Name
	// 		keyvals = append(keyvals, "asn", hop.Asn, "org", hop.OrgName)
	// 	}
	// 	log.Info("hop", keyvals)
	return nil
}

func runCmdToolTLS(args []string) error {
	target := args[0]

	cfg := server.GetConfig()
	m := server.New(server.WithConfig(cfg))

	info, err := m.FetchTLSInfo(context.Background(), target)
	if err != nil {
		return err
	}
	log.Info("tls", "target", target, "tls", info)
	return nil
}

func runCmdToolSNMP(args []string) error {
	target := args[0]

	cfg := server.GetConfig()
	m := server.New(server.WithConfig(cfg))

	info, err := m.FetchSNMPInfo(context.Background(), target)
	if err != nil {
		return err
	}
	log.Info(
		"snmp systeminfo",
		"target",
		target,
		"name",
		info.SystemInfo.Name,
		"contact",
		info.SystemInfo.Contact,
		"location",
		info.SystemInfo.Location,
		"description",
		info.SystemInfo.Description,
	)
	for _, iface := range info.Interfaces {
		log.Info("snmp interface", "addr", iface.Addr())
	}
	for _, arpentry := range info.ArpTable {
		log.Info("snmp arpentry", "addr", arpentry.Addr, "mac", arpentry.MAC)
	}

	return nil
}

func runCmdToolCheckDNS(args []string) error {
	target := args[0]

	cfg := server.GetConfig()
	m := server.New(server.WithConfig(cfg))

	ret, err := m.CheckDNS(context.Background(), target)
	if err != nil {
		return err
	}

	for company, servers := range ret {
		for server, recs := range servers {
			log.Info("dns", "target", target, "company", company, "server", server, "records", recs)
		}
	}

	return nil
}
