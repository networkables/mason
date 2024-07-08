// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package netflows

import (
	"context"
	"net"

	"github.com/charmbracelet/log"

	"github.com/networkables/mason/internal/model"
)

/*

- listener passes bytes out a channel
- workerpool processes into flow struct, output on channel
- persist flow to db

*/

func Listen(ctx context.Context, cfg *Config) chan []byte {
	output := make(chan []byte)
	listenaddy, err := net.ResolveUDPAddr("udp", cfg.ListenAddress)
	if err != nil {
		log.Fatalf("resolveudpaddr: %v", err)
	}
	conn, err := net.ListenUDP("udp", listenaddy)
	if err != nil {
		log.Fatalf("listenudp: %v", err)
	}
	log.Info("starting netflow server", "addr", cfg.ListenAddress)

	go func(pktsize int) {
		defer conn.Close()
		defer close(output)
		for {
			if ctx.Err() != nil {
				log.Info("netflow listener shutdown")
				return
			}
			buff := make([]byte, pktsize)
			size, _, err := conn.ReadFromUDP(buff)
			if err != nil {
				if size == 0 {
					return
				}
				log.Fatalf("readfromudp: %v", err)
			}
			output <- buff
		}
	}(cfg.PacketSize)

	return output

	// ipflows := rawsToIpFlows(flows)
	// for _, ipflow := range ipflows {
	// 	err = db.CreateFlow(ctx, ipflow)
	// 	if err != nil {
	// 		log.Printf("createflow: %v\n", err)
	// 		return
	// 	}
	// }
}

func parser(ctx context.Context, pkt []byte) ([]model.IpFlow, error) {
	if ctx.Err() != nil {
		return nil, nil
	}
	rawflows, err := handlePacket(pkt)
	if err != nil {
		log.Errorf("handlepacket: %v", err)
		return nil, err
	}
	ipflows := rawsToIpFlows(rawflows)
	return ipflows, nil
}
