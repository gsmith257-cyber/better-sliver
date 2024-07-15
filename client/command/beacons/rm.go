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
	"github.com/spf13/cobra"

	"github.com/gsmith257-cyber/better-sliver-package/client/console"
)

// BaconsRmCmd - Display/interact with bacons
func BaconsRmCmd(cmd *cobra.Command, con *console.SliverClient, args []string) {
	bacon, err := SelectBacon(con)
	if err != nil {
		con.PrintErrorf("%s\n", err)
		return
	}
	grpcCtx, cancel := con.GrpcContext(cmd)
	defer cancel()
	_, err = con.Rpc.RmBacon(grpcCtx, bacon)
	if err != nil {
		con.PrintErrorf("%s\n", err)
		return
	}
	con.PrintInfof("Bacon removed (%s)\n", bacon.ID)
}
