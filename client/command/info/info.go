package info

/*
	Sliver Implant Framework
	Copyright (C) 2019  Bishop Fox

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
	"time"

	"github.com/gsmith257-cyber/better-sliver-package/client/command/use"
	"github.com/gsmith257-cyber/better-sliver-package/client/console"
	consts "github.com/gsmith257-cyber/better-sliver-package/client/constants"
	"github.com/gsmith257-cyber/better-sliver-package/protobuf/clientpb"
	"github.com/gsmith257-cyber/better-sliver-package/protobuf/sliverpb"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/proto"
)

// InfoCmd - Display information about the active session.
func InfoCmd(cmd *cobra.Command, con *console.SliverClient, args []string) {
	var err error

	// Check if we have an active target via 'use'
	session, bacon := con.ActiveTarget.Get()

	if len(args) > 0 {
		// ID passed via argument takes priority
		idArg := args[0]
		session, bacon, err = use.SessionOrBaconByID(idArg, con)
		if err != nil {
			con.PrintErrorf("%s\n", err)
			return
		}
	} else if session != nil || bacon != nil {
		currID := ""
		if session != nil {
			currID = session.ID
		} else {
			currID = bacon.ID
		}
		session, bacon, err = use.SessionOrBaconByID(currID, con)
		if err != nil {
			con.PrintErrorf("%s\n", err)
			return
		}
	} else {
		if session == nil && bacon == nil {
			session, bacon, err = use.SelectSessionOrBacon(con)
			if err != nil {
				con.PrintErrorf("%s\n", err)
				return
			}
		}
	}

	if session != nil {

		con.Printf(console.Bold+"        Session ID: %s%s\n", console.Normal, session.ID)
		con.Printf(console.Bold+"              Name: %s%s\n", console.Normal, session.Name)
		con.Printf(console.Bold+"          Hostname: %s%s\n", console.Normal, session.Hostname)
		con.Printf(console.Bold+"              UUID: %s%s\n", console.Normal, session.UUID)
		con.Printf(console.Bold+"          Username: %s%s\n", console.Normal, session.Username)
		con.Printf(console.Bold+"               UID: %s%s\n", console.Normal, session.UID)
		con.Printf(console.Bold+"               GID: %s%s\n", console.Normal, session.GID)
		con.Printf(console.Bold+"               PID: %s%d\n", console.Normal, session.PID)
		con.Printf(console.Bold+"                OS: %s%s\n", console.Normal, session.OS)
		con.Printf(console.Bold+"           Version: %s%s\n", console.Normal, session.Version)
		con.Printf(console.Bold+"            Locale: %s%s\n", console.Normal, session.Locale)
		con.Printf(console.Bold+"              Arch: %s%s\n", console.Normal, session.Arch)
		con.Printf(console.Bold+"         Active C2: %s%s\n", console.Normal, session.ActiveC2)
		con.Printf(console.Bold+"    Remote Address: %s%s\n", console.Normal, session.RemoteAddress)
		con.Printf(console.Bold+"         Proxy URL: %s%s\n", console.Normal, session.ProxyURL)
		con.Printf(console.Bold+"Reconnect Interval: %s%s\n", console.Normal, time.Duration(session.ReconnectInterval).String())
		con.Printf(console.Bold+"     First Contact: %s%s\n", console.Normal, con.FormatDateDelta(time.Unix(session.FirstContact, 0), true, false))
		con.Printf(console.Bold+"      Last Checkin: %s%s\n", console.Normal, con.FormatDateDelta(time.Unix(session.LastCheckin, 0), true, false))

	} else if bacon != nil {

		con.Printf(console.Bold+"         Bacon ID: %s%s\n", console.Normal, bacon.ID)
		con.Printf(console.Bold+"              Name: %s%s\n", console.Normal, bacon.Name)
		con.Printf(console.Bold+"          Hostname: %s%s\n", console.Normal, bacon.Hostname)
		con.Printf(console.Bold+"              UUID: %s%s\n", console.Normal, bacon.UUID)
		con.Printf(console.Bold+"          Username: %s%s\n", console.Normal, bacon.Username)
		con.Printf(console.Bold+"               UID: %s%s\n", console.Normal, bacon.UID)
		con.Printf(console.Bold+"               GID: %s%s\n", console.Normal, bacon.GID)
		con.Printf(console.Bold+"               PID: %s%d\n", console.Normal, bacon.PID)
		con.Printf(console.Bold+"                OS: %s%s\n", console.Normal, bacon.OS)
		con.Printf(console.Bold+"           Version: %s%s\n", console.Normal, bacon.Version)
		con.Printf(console.Bold+"            Locale: %s%s\n", console.Normal, bacon.Locale)
		con.Printf(console.Bold+"              Arch: %s%s\n", console.Normal, bacon.Arch)
		con.Printf(console.Bold+"         Active C2: %s%s\n", console.Normal, bacon.ActiveC2)
		con.Printf(console.Bold+"    Remote Address: %s%s\n", console.Normal, bacon.RemoteAddress)
		con.Printf(console.Bold+"         Proxy URL: %s%s\n", console.Normal, bacon.ProxyURL)
		con.Printf(console.Bold+"          Interval: %s%s\n", console.Normal, time.Duration(bacon.Interval).String())
		con.Printf(console.Bold+"            Jitter: %s%s\n", console.Normal, time.Duration(bacon.Jitter).String())
		con.Printf(console.Bold+"     First Contact: %s%s\n", console.Normal, con.FormatDateDelta(time.Unix(bacon.FirstContact, 0), true, false))
		con.Printf(console.Bold+"      Last Checkin: %s%s\n", console.Normal, con.FormatDateDelta(time.Unix(bacon.LastCheckin, 0), true, false))
		con.Printf(console.Bold+"      Next Checkin: %s%s\n", console.Normal, con.FormatDateDelta(time.Unix(bacon.NextCheckin, 0), true, true))

	} else {
		con.PrintErrorf("No target session, see `help %s`\n", consts.InfoStr)
	}
}

// PIDCmd - Get the active session's PID.
func PIDCmd(cmd *cobra.Command, con *console.SliverClient, args []string) {
	session, bacon := con.ActiveTarget.GetInteractive()
	if session == nil && bacon == nil {
		return
	}
	if session != nil {
		con.Printf("%d\n", session.PID)
	} else if bacon != nil {
		con.Printf("%d\n", bacon.PID)
	}
}

// UIDCmd - Get the active session's UID.
func UIDCmd(cmd *cobra.Command, con *console.SliverClient, args []string) {
	session, bacon := con.ActiveTarget.GetInteractive()
	if session == nil && bacon == nil {
		return
	}
	if session != nil {
		con.Printf("%s\n", session.UID)
	} else if bacon != nil {
		con.Printf("%s\n", bacon.UID)
	}
}

// GIDCmd - Get the active session's GID.
func GIDCmd(cmd *cobra.Command, con *console.SliverClient, args []string) {
	session, bacon := con.ActiveTarget.GetInteractive()
	if session == nil && bacon == nil {
		return
	}
	if session != nil {
		con.Printf("%s\n", session.GID)
	} else if bacon != nil {
		con.Printf("%s\n", bacon.GID)
	}
}

// WhoamiCmd - Displays the current user of the active session.
func WhoamiCmd(cmd *cobra.Command, con *console.SliverClient, args []string) {
	session, bacon := con.ActiveTarget.GetInteractive()
	if session == nil && bacon == nil {
		return
	}

	var isWin bool
	con.Printf("Logon ID: ")
	if session != nil {
		con.Printf("%s\n", session.Username)
		if session.GetOS() == "windows" {
			isWin = true
		}
	} else if bacon != nil {
		con.Printf("%s\n", bacon.Username)
		if bacon.GetOS() == "windows" {
			isWin = true
		}
	}

	if isWin {
		cto, err := con.Rpc.CurrentTokenOwner(context.Background(), &sliverpb.CurrentTokenOwnerReq{
			Request: con.ActiveTarget.Request(cmd),
		})
		if err != nil {
			con.PrintErrorf("%s\n", err)
			return
		}

		if cto.Response != nil && cto.Response.Async {
			con.AddBaconCallback(cto.Response.TaskID, func(task *clientpb.BaconTask) {
				err = proto.Unmarshal(task.Response, cto)
				if err != nil {
					con.PrintErrorf("Failed to decode response %s\n", err)
					return
				}
				PrintTokenOwner(cto, con)
			})
			con.PrintAsyncResponse(cto.Response)
		} else {
			PrintTokenOwner(cto, con)
		}
	}
}

func PrintTokenOwner(cto *sliverpb.CurrentTokenOwner, con *console.SliverClient) {
	if cto.Response != nil && cto.Response.Err != "" {
		con.PrintErrorf("%s\n", cto.Response.Err)
		return
	}
	con.PrintInfof("Current Token ID: %s", cto.Output)
}
