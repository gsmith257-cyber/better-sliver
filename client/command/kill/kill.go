package kill

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
	"context"
	"errors"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/gsmith257-cyber/better-sliver-package/client/console"
	"github.com/gsmith257-cyber/better-sliver-package/client/core"
	"github.com/gsmith257-cyber/better-sliver-package/protobuf/clientpb"
	"github.com/gsmith257-cyber/better-sliver-package/protobuf/commonpb"
	"github.com/gsmith257-cyber/better-sliver-package/protobuf/sliverpb"
)

// KillCmd - Kill the active session (not to be confused with TerminateCmd)
func KillCmd(cmd *cobra.Command, con *console.SliverClient, args []string) {
	session, bacon := con.ActiveTarget.GetInteractive()
	// Confirm with the user, just in case they confused kill with terminate
	confirm := false
	con.PrintWarnf("WARNING: This will kill the remote implant process\n\n")
	if session != nil {
		survey.AskOne(&survey.Confirm{Message: "Kill the active session?"}, &confirm, nil)
		if !confirm {
			return
		}
		err := KillSession(session, cmd, con)
		if err != nil {
			con.PrintErrorf("%s\n", err)
			return
		}
		con.PrintInfof("Killed %s (%s)\n", session.Name, session.ID)
		con.ActiveTarget.Background()
		return
	} else if bacon != nil {
		survey.AskOne(&survey.Confirm{Message: "Kill the active bacon?"}, &confirm, nil)
		if !confirm {
			return
		}
		err := KillBacon(bacon, cmd, con)
		if err != nil {
			con.PrintErrorf("%s\n", err)
			return
		}
		con.PrintInfof("Killed %s (%s)\n", bacon.Name, bacon.ID)
		con.ActiveTarget.Background()
		return
	}
	con.PrintErrorf("No active session or bacon\n")
}

func KillSession(session *clientpb.Session, cmd *cobra.Command, con *console.SliverClient) error {
	if session == nil {
		return errors.New("session does not exist")
	}
	timeout, _ := cmd.Flags().GetInt64("timeout")
	force, _ := cmd.Flags().GetBool("force")

	// remove any active socks proxies
	socks := core.SocksProxies.List()
	if len(socks) != 0 {
		for _, p := range socks {
			if p.SessionID == session.ID {
				core.SocksProxies.Remove(p.ID)
			}
		}
	}

	_, err := con.Rpc.Kill(context.Background(), &sliverpb.KillReq{
		Request: &commonpb.Request{
			SessionID: session.ID,
			Timeout:   timeout,
		},
		Force: force,
	})
	return err
}

func KillBacon(bacon *clientpb.Bacon, cmd *cobra.Command, con *console.SliverClient) error {
	if bacon == nil {
		return errors.New("session does not exist")
	}

	timeout, _ := cmd.Flags().GetInt64("timeout")
	force, _ := cmd.Flags().GetBool("force")

	_, err := con.Rpc.Kill(context.Background(), &sliverpb.KillReq{
		Request: &commonpb.Request{
			BaconID: bacon.ID,
			Timeout:  timeout,
		},
		Force: force,
	})
	return err
}
