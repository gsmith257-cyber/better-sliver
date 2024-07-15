package bacons

import (
	"context"
	"fmt"
	"strings"

	"github.com/rsteube/carapace"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/gsmith257-cyber/better-sliver-package/client/command/flags"
	"github.com/gsmith257-cyber/better-sliver-package/client/command/help"
	"github.com/gsmith257-cyber/better-sliver-package/client/console"
	consts "github.com/gsmith257-cyber/better-sliver-package/client/constants"
	"github.com/gsmith257-cyber/better-sliver-package/protobuf/commonpb"
)

// Commands returns the “ command and its subcommands.
func Commands(con *console.SliverClient) []*cobra.Command {
	baconsCmd := &cobra.Command{
		Use:     consts.BaconsStr,
		Short:   "Manage bacons",
		Long:    help.GetHelpFor([]string{consts.BaconsStr}),
		GroupID: consts.SliverHelpGroup,
		Run: func(cmd *cobra.Command, args []string) {
			BaconsCmd(cmd, con, args)
		},
	}
	flags.Bind("bacons", true, baconsCmd, func(f *pflag.FlagSet) {
		f.IntP("timeout", "t", flags.DefaultTimeout, "grpc timeout in seconds")
	})
	flags.Bind("bacons", false, baconsCmd, func(f *pflag.FlagSet) {
		f.StringP("kill", "k", "", "kill the designated bacon")
		f.BoolP("kill-all", "K", false, "kill all bacons")
		f.BoolP("force", "F", false, "force killing the bacon")

		f.StringP("filter", "f", "", "filter bacons by substring")
		f.StringP("filter-re", "e", "", "filter bacons by regular expression")
	})
	flags.BindFlagCompletions(baconsCmd, func(comp *carapace.ActionMap) {
		(*comp)["kill"] = BaconIDCompleter(con)
	})
	baconsRmCmd := &cobra.Command{
		Use:   consts.RmStr,
		Short: "Remove a bacon",
		Long:  help.GetHelpFor([]string{consts.BaconsStr, consts.RmStr}),
		Run: func(cmd *cobra.Command, args []string) {
			BaconsRmCmd(cmd, con, args)
		},
	}
	carapace.Gen(baconsRmCmd).PositionalCompletion(BaconIDCompleter(con))
	baconsCmd.AddCommand(baconsRmCmd)

	baconsWatchCmd := &cobra.Command{
		Use:   consts.WatchStr,
		Short: "Watch your bacons",
		Long:  help.GetHelpFor([]string{consts.BaconsStr, consts.WatchStr}),
		Run: func(cmd *cobra.Command, args []string) {
			BaconsWatchCmd(cmd, con, args)
		},
	}
	baconsCmd.AddCommand(baconsWatchCmd)

	baconsPruneCmd := &cobra.Command{
		Use:   consts.PruneStr,
		Short: "Prune stale bacons automatically",
		Long:  help.GetHelpFor([]string{consts.BaconsStr, consts.PruneStr}),
		Run: func(cmd *cobra.Command, args []string) {
			BaconsPruneCmd(cmd, con, args)
		},
	}
	flags.Bind("bacons", false, baconsPruneCmd, func(f *pflag.FlagSet) {
		f.StringP("duration", "d", "1h", "duration to prune bacons that have missed their last checkin")
	})
	baconsCmd.AddCommand(baconsPruneCmd)

	return []*cobra.Command{baconsCmd}
}

// BaconIDCompleter completes bacon IDs
func BaconIDCompleter(con *console.SliverClient) carapace.Action {
	callback := func(_ carapace.Context) carapace.Action {
		results := make([]string, 0)

		bacons, err := con.Rpc.GetBacons(context.Background(), &commonpb.Empty{})
		if err == nil {
			for _, b := range bacons.Bacons {
				link := fmt.Sprintf("[%s <- %s]", b.ActiveC2, b.RemoteAddress)
				id := fmt.Sprintf("%s (%d)", b.Name, b.PID)
				userHost := fmt.Sprintf("%s@%s", b.Username, b.Hostname)
				desc := strings.Join([]string{id, userHost, link}, " ")

				results = append(results, b.ID[:8])
				results = append(results, desc)
			}
		}
		return carapace.ActionValuesDescribed(results...).Tag("bacons")
	}

	return carapace.ActionCallback(callback)
}
