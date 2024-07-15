package bacons

/*
	Sliver Implant Framework
	Copyright (C) 2021  Bishop Fox

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/gsmith257-cyber/better-sliver-package/client/command/kill"
	"github.com/gsmith257-cyber/better-sliver-package/client/command/settings"
	"github.com/gsmith257-cyber/better-sliver-package/client/console"
	"github.com/gsmith257-cyber/better-sliver-package/protobuf/clientpb"
	"github.com/gsmith257-cyber/better-sliver-package/protobuf/commonpb"
)

// BaconsCmd - Display/interact with bacons
func BaconsCmd(cmd *cobra.Command, con *console.SliverClient, args []string) {
	killFlag, _ := cmd.Flags().GetString("kill")
	killAll, _ := cmd.Flags().GetBool("kill-all")

	// Handle kill
	if killFlag != "" {
		bacon, err := GetBacon(con, killFlag)
		if err != nil {
			con.PrintErrorf("%s\n", err)
			return
		}
		err = kill.KillBacon(bacon, cmd, con)
		if err != nil {
			con.PrintErrorf("%s\n", err)
			return
		}
		con.Println()
		con.PrintInfof("Killed %s (%s)\n", bacon.Name, bacon.ID)
	}

	if killAll {
		bacons, err := GetBacons(con)
		if err != nil {
			con.PrintErrorf("%s\n", err)
			return
		}
		for _, bacon := range bacons.Bacons {
			err = kill.KillBacon(bacon, cmd, con)
			if err != nil {
				con.PrintErrorf("%s\n", err)
				return
			}
			con.Println()
			con.PrintInfof("Killed %s (%s)\n", bacon.Name, bacon.ID)
		}
	}
	filter, _ := cmd.Flags().GetString("filter")
	var filterRegex *regexp.Regexp
	if filterRe, _ := cmd.Flags().GetString("filter-re"); filterRe != "" {
		var err error
		filterRegex, err = regexp.Compile(filterRe)
		if err != nil {
			con.PrintErrorf("%s\n", err)
			return
		}
	}

	grpcCtx, cancel := con.GrpcContext(cmd)
	defer cancel()
	bacons, err := con.Rpc.GetBacons(grpcCtx, &commonpb.Empty{})
	if err != nil {
		con.PrintErrorf("%s\n", err)
		return
	}
	PrintBacons(bacons.Bacons, filter, filterRegex, con)
}

// PrintBacons - Display a list of bacons
func PrintBacons(bacons []*clientpb.Bacon, filter string, filterRegex *regexp.Regexp, con *console.SliverClient) {
	if len(bacons) == 0 {
		con.PrintInfof("No bacons 🙁\n")
		return
	}
	tw := renderBacons(bacons, filter, filterRegex, con)
	con.Printf("%s\n", tw.Render())
}

func renderBacons(bacons []*clientpb.Bacon, filter string, filterRegex *regexp.Regexp, con *console.SliverClient) table.Writer {
	width, _, err := term.GetSize(0)
	if err != nil {
		width = 999
	}

	tw := table.NewWriter()
	tw.SetStyle(settings.GetTableStyle(con))
	wideTermWidth := con.Settings.SmallTermWidth < width
	windowsBaconInList := false
	for _, bacon := range bacons {
		if bacon.OS == "windows" {
			windowsBaconInList = true
		}
	}
	if wideTermWidth {
		if windowsBaconInList {
			tw.AppendHeader(table.Row{
				"ID",
				"Name",
				"Tasks",
				"Transport",
				"Remote Address",
				"Hostname",
				"Username",
				"Process (PID)",
				"Integrity",
				"Operating System",
				"Locale",
				"Last Check-in",
				"Next Check-in",
			})
		} else {
			tw.AppendHeader(table.Row{
				"ID",
				"Name",
				"Tasks",
				"Transport",
				"Remote Address",
				"Hostname",
				"Username",
				"Process (PID)",
				"Operating System",
				"Locale",
				"Last Check-in",
				"Next Check-in",
			})
		}
	} else {
		tw.AppendHeader(table.Row{
			"ID",
			"Name",
			"Transport",
			"Hostname",
			"Username",
			"Operating System",
			"Last Check-in",
			"Next Check-in",
		})
	}

	for _, bacon := range bacons {
		color := console.Normal
		activeBacon := con.ActiveTarget.GetBacon()
		if activeBacon != nil && activeBacon.ID == bacon.ID {
			color = console.Green
		}
		if bacon.Integrity == "" {
			bacon.Integrity = "-"
		}

		// We need a slice of strings so we can apply filters
		var rowEntries []string

		if wideTermWidth {
			rowEntries = []string{
				fmt.Sprintf(color+"%s"+console.Normal, strings.Split(bacon.ID, "-")[0]),
				fmt.Sprintf(color+"%s"+console.Normal, bacon.Name),
				fmt.Sprintf(color+"%d/%d"+console.Normal, bacon.TasksCountCompleted, bacon.TasksCount),
				fmt.Sprintf(color+"%s"+console.Normal, bacon.Transport),
				fmt.Sprintf(color+"%s"+console.Normal, bacon.RemoteAddress),
				fmt.Sprintf(color+"%s"+console.Normal, bacon.Hostname),
				fmt.Sprintf(color+"%s"+console.Normal, strings.TrimPrefix(bacon.Username, bacon.Hostname+"\\")),
				fmt.Sprintf(color+"%s (%d)"+console.Normal, bacon.Filename, bacon.PID),
			}

			if windowsBaconInList {
				rowEntries = append(rowEntries, fmt.Sprintf(color+"%s"+console.Normal, bacon.Integrity))
			}

			rowEntries = append(rowEntries, []string{
				fmt.Sprintf(color+"%s/%s"+console.Normal, bacon.OS, bacon.Arch),
				fmt.Sprintf(color+"%s"+console.Normal, bacon.Locale),
				con.FormatDateDelta(time.Unix(bacon.LastCheckin, 0), wideTermWidth, false),
				con.FormatDateDelta(time.Unix(bacon.NextCheckin, 0), wideTermWidth, true),
			}...)
		} else {
			rowEntries = []string{
				fmt.Sprintf(color+"%s"+console.Normal, strings.Split(bacon.ID, "-")[0]),
				fmt.Sprintf(color+"%s"+console.Normal, bacon.Name),
				fmt.Sprintf(color+"%s"+console.Normal, bacon.Transport),
				fmt.Sprintf(color+"%s"+console.Normal, bacon.Hostname),
				fmt.Sprintf(color+"%s"+console.Normal, strings.TrimPrefix(bacon.Username, bacon.Hostname+"\\")),
				fmt.Sprintf(color+"%s/%s"+console.Normal, bacon.OS, bacon.Arch),
				con.FormatDateDelta(time.Unix(bacon.LastCheckin, 0), wideTermWidth, false),
				con.FormatDateDelta(time.Unix(bacon.NextCheckin, 0), wideTermWidth, true),
			}
		}
		// Build the row struct
		row := table.Row{}
		for _, entry := range rowEntries {
			row = append(row, entry)
		}
		// Apply filters if any
		if filter == "" && filterRegex == nil {
			tw.AppendRow(row)
		} else {
			for _, rowEntry := range rowEntries {
				if filter != "" {
					if strings.Contains(rowEntry, filter) {
						tw.AppendRow(row)
						break
					}
				}
				if filterRegex != nil {
					if filterRegex.MatchString(rowEntry) {
						tw.AppendRow(row)
						break
					}
				}
			}
		}
	}
	return tw
}
