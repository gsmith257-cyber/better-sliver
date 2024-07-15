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
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/gsmith257-cyber/better-sliver-package/client/console"
	"github.com/gsmith257-cyber/better-sliver-package/protobuf/clientpb"
	"github.com/gsmith257-cyber/better-sliver-package/protobuf/commonpb"
)

// BaconsPruneCmd - Prune stale bacons automatically
func BaconsPruneCmd(cmd *cobra.Command, con *console.SliverClient, args []string) {
	duration, _ := cmd.Flags().GetString("duration")
	pruneDuration, err := time.ParseDuration(duration)
	if err != nil {
		con.PrintErrorf("%s\n", err)
		return
	}
	con.PrintInfof("Pruning bacons that missed their last checking by %s or more...\n\n", pruneDuration)
	grpcCtx, cancel := con.GrpcContext(cmd)
	defer cancel()
	bacons, err := con.Rpc.GetBacons(grpcCtx, &commonpb.Empty{})
	if err != nil {
		con.PrintErrorf("%s\n", err)
		return
	}
	pruneBacons := []*clientpb.Bacon{}
	for _, bacon := range bacons.Bacons {
		nextCheckin := time.Unix(bacon.NextCheckin, 0)
		if time.Now().Before(nextCheckin) {
			continue
		}
		delta := time.Since(nextCheckin)
		if pruneDuration <= delta {
			pruneBacons = append(pruneBacons, bacon)
		}
	}
	if len(pruneBacons) == 0 {
		con.PrintInfof("No bacons to prune.\n")
		return
	}
	con.PrintWarnf("The following bacons and their tasks will be removed:\n")
	for index, bacon := range pruneBacons {
		bacon, err := con.Rpc.GetBacon(grpcCtx, &clientpb.Bacon{ID: bacon.ID})
		if err != nil {
			con.PrintErrorf("%s\n", err)
			continue
		}
		con.Printf("\t%d. %s (%s)\n", (index + 1), bacon.Name, bacon.ID)
	}
	con.Println()
	confirm := false
	prompt := &survey.Confirm{Message: "Prune these bacons?"}
	survey.AskOne(prompt, &confirm)
	if !confirm {
		return
	}
	errCount := 0
	for _, bacon := range pruneBacons {
		_, err := con.Rpc.RmBacon(grpcCtx, &clientpb.Bacon{ID: bacon.ID})
		if err != nil {
			con.PrintErrorf("%s\n", err)
			errCount++
		}
	}
	con.PrintInfof("Pruned %d bacon(s)\n", len(pruneBacons)-errCount)
}
