// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package commands

import (
	"github.com/spf13/cobra"
)

var (
	cmdDebug = &cobra.Command{
		Use:   "debug",
		Short: "debug/internal tools",
		// Long:  `start server`,
	}

	cmdDebugDumpWSP = &cobra.Command{
		Use:   "dumpwsp [filename]",
		Short: "dump whisper file",
		// Long:  `start server`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmdDebugDumpWSP(args)
		},
	}
)

func init() {
	cmdDebug.AddCommand(cmdDebugDumpWSP)
}

func runCmdDebugDumpWSP([]string) error {
	// cfg := GetConfig()
	// name := args[0]
	//
	// ps := newTimeseriesStore(cfg.Store)
	// w, err := ps.openWhisperFile("data/store/" + name)
	// if err != nil {
	// 	return err
	// }
	// defer w.Close()
	// // w.Dump(true, true)
	// ts, err := w.Fetch(fetchlast(24 * time.Hour))
	// if err != nil {
	// 	return err
	// }
	// points := ts.Points()
	// for _, point := range points {
	// 	fmt.Println(time.Unix(int64(point.Time), 0), point.Value)
	// }
	return nil
}
