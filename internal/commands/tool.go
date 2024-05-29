// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"context"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/networkables/mason/internal/config"
	"github.com/networkables/mason/internal/server"
)

var (
	cmdTool = &cobra.Command{
		Use:   "tool",
		Short: "network tools",
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

	m := server.NewTool()
	cfg := config.GetConfig()

	mac, err := m.ArpPing(ctx, target, *cfg.Discovery.ArpPing.Timeout)
	if err != nil {
		return err
	}
	log.Info("arpping", "target", target, "mac", mac)
	return nil
}

func runCmdPing(args []string) error {
	target := args[0]

	cfg := config.GetConfig()
	m := server.NewTool()

	stats, err := m.IcmpPing(
		context.Background(),
		target,
		*cfg.Discovery.IcmpPing.PingCount,
		*cfg.Discovery.IcmpPing.Timeout,
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
	cfg := config.GetConfig()
	target := args[0]

	m := server.NewTool()
	ports, err := m.Portscan(context.Background(), target, cfg.Enrichment.PortScanner)
	if err != nil {
		return err
	}
	log.Info("portscan", "target", target, "openports", ports)

	return nil
}

func runCmdToolExternalIP(args []string) error {
	m := server.NewTool()
	addr, err := m.GetExternalAddr(context.Background())
	if err != nil {
		return err
	}
	log.Info("external address", "ip", addr)
	return nil
}

func runCmdToolTraceroute(args []string) error {
	target := args[0]

	m := server.NewTool()
	hops, err := m.Traceroute(context.Background(), target)
	if err != nil {
		return err
	}
	log.Info("traceroute", "target", target)
	for i, hop := range hops {
		log.Info("hop", "position", i, "hop", hop)
		log.Info(
			"hop",
			"peer",
			hop.Peer,
			"count",
			hop.TotalPackets,
			"packetloss",
			hop.PacketLoss,
			"min",
			hop.Minimum,
			"mean",
			hop.Mean,
			"max",
			hop.Maximum,
			"stddev",
			hop.StdDev,
		)
	}
	return nil
}

func runCmdToolTLS(args []string) error {
	target := args[0]

	m := server.NewTool()
	info, err := m.FetchTLSInfo(context.Background(), target)
	if err != nil {
		return err
	}
	log.Info("tls", "target", target, "tls", info)
	return nil
}

func runCmdToolSNMP(args []string) error {
	target := args[0]

	m := server.NewTool()
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

	m := server.NewTool()
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
