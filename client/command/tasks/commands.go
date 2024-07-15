package tasks

import (
	"github.com/gsmith257-cyber/better-sliver-package/client/command/flags"
	"github.com/gsmith257-cyber/better-sliver-package/client/command/help"
	"github.com/gsmith257-cyber/better-sliver-package/client/console"
	consts "github.com/gsmith257-cyber/better-sliver-package/client/constants"
	"github.com/rsteube/carapace"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Commands returns the “ command and its subcommands.
func Commands(con *console.SliverClient) []*cobra.Command {
	tasksCmd := &cobra.Command{
		Use:   consts.TasksStr,
		Short: "Bacon task management",
		Long:  help.GetHelpFor([]string{consts.TasksStr}),
		Run: func(cmd *cobra.Command, args []string) {
			TasksCmd(cmd, con, args)
		},
		GroupID:     consts.SliverCoreHelpGroup,
		Annotations: flags.RestrictTargets(consts.BaconCmdsFilter),
	}
	flags.Bind("tasks", true, tasksCmd, func(f *pflag.FlagSet) {
		f.IntP("timeout", "t", flags.DefaultTimeout, "grpc timeout in seconds")
		f.BoolP("overflow", "O", false, "overflow terminal width (display truncated rows)")
		f.IntP("skip-pages", "S", 0, "skip the first n page(s)")
		f.StringP("filter", "f", "", "filter based on task type (case-insensitive prefix matching)")
	})

	fetchCmd := &cobra.Command{
		Use:   consts.FetchStr,
		Short: "Fetch the details of a bacon task",
		Long:  help.GetHelpFor([]string{consts.TasksStr, consts.FetchStr}),
		Args:  cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			TasksFetchCmd(cmd, con, args)
		},
	}
	tasksCmd.AddCommand(fetchCmd)
	carapace.Gen(fetchCmd).PositionalCompletion(BaconTaskIDCompleter(con).Usage("bacon task ID"))

	cancelCmd := &cobra.Command{
		Use:   consts.CancelStr,
		Short: "Cancel a pending bacon task",
		Long:  help.GetHelpFor([]string{consts.TasksStr, consts.CancelStr}),
		Args:  cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			TasksCancelCmd(cmd, con, args)
		},
	}
	tasksCmd.AddCommand(cancelCmd)
	carapace.Gen(cancelCmd).PositionalCompletion(BaconPendingTasksCompleter(con).Usage("bacon task ID"))

	return []*cobra.Command{tasksCmd}
}
