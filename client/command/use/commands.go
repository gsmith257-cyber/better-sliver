package use

import (
	"github.com/rsteube/carapace"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/gsmith257-cyber/better-sliver-package/client/command/bacons"
	"github.com/gsmith257-cyber/better-sliver-package/client/command/flags"
	"github.com/gsmith257-cyber/better-sliver-package/client/command/help"
	"github.com/gsmith257-cyber/better-sliver-package/client/console"
	consts "github.com/gsmith257-cyber/better-sliver-package/client/constants"
)

// Commands returns the “ command and its subcommands.
func Commands(con *console.SliverClient) []*cobra.Command {
	useCmd := &cobra.Command{
		Use:   consts.UseStr,
		Short: "Switch the active session or bacon",
		Long:  help.GetHelpFor([]string{consts.UseStr}),
		Run: func(cmd *cobra.Command, args []string) {
			UseCmd(cmd, con, args)
		},
		GroupID: consts.SliverHelpGroup,
	}
	flags.Bind("use", true, useCmd, func(f *pflag.FlagSet) {
		f.Int64P("timeout", "t", flags.DefaultTimeout, "grpc timeout in seconds")
	})
	carapace.Gen(useCmd).PositionalCompletion(BaconAndSessionIDCompleter(con))

	useSessionCmd := &cobra.Command{
		Use:   consts.SessionsStr,
		Short: "Switch the active session",
		Long:  help.GetHelpFor([]string{consts.UseStr, consts.SessionsStr}),
		Run: func(cmd *cobra.Command, args []string) {
			UseSessionCmd(cmd, con, args)
		},
	}
	carapace.Gen(useSessionCmd).PositionalCompletion(SessionIDCompleter(con))
	useCmd.AddCommand(useSessionCmd)

	useBaconCmd := &cobra.Command{
		Use:   consts.BaconsStr,
		Short: "Switch the active bacon",
		Long:  help.GetHelpFor([]string{consts.UseStr, consts.BaconsStr}),
		Run: func(cmd *cobra.Command, args []string) {
			UseBaconCmd(cmd, con, args)
		},
	}
	carapace.Gen(useBaconCmd).PositionalCompletion(bacons.BaconIDCompleter(con))
	useCmd.AddCommand(useBaconCmd)

	return []*cobra.Command{useCmd}
}
