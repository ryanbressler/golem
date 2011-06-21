/*
   Copyright (C) 2003-2011 Institute for Systems Biology
                           Seattle, Washington, USA.

   This library is free software; you can redistribute it and/or
   modify it under the terms of the GNU Lesser General Public
   License as published by the Free Software Foundation; either
   version 2.1 of the License, or (at your option) any later version.

   This library is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
   Lesser General Public License for more details.

   You should have received a copy of the GNU Lesser General Public
   License along with this library; if not, write to the Free Software
   Foundation, Inc., 59 Temple Place, Suite 330, Boston, MA 02111-1307  USA

*/
package main

import (
	"os"
	"exec"
	"time"
)

//restarts his process using the origional commands after waitn nanoseconds then die
func RestartIn(waitn int64) {
	log("Restart in %v nanoseconds", waitn)
	time.Sleep(waitn)
	log("Restarting.")
	cmd, err := exec.LookPath(os.Args[0])
	if err != nil {
		log("exec %s: %s\n", os.Args, err)
		return
	}
	_, err = exec.Run(cmd, os.Args, nil, "./", exec.DevNull, exec.PassThrough, exec.PassThrough)
	if err != nil {
		log("%v", err)
	}
	log("Exiting.")
	os.Exit(0)
}

//exit this proccess in waitn n nanoseconds
func DieIn(waitn int64) {
	log("Die in %v nanoseconds", waitn)
	time.Sleep(waitn)
	log("Exiting.")
	os.Exit(0)
}

//kill the supplied pid... uses the kill command, there must be a better way to do this
//TODO: find a better way to do this
func KillPid(pid string) {
	killcmd, err := exec.LookPath("kill")
	if err != nil {
		log("kill not found")
		return
	}
	log("Killing Pid: %v", pid)
	_, err = exec.Run(killcmd, []string{killcmd, pid}, nil, "./", exec.DevNull, exec.PassThrough, exec.PassThrough)
	if err != nil {
		log("%v", err)
	}
}
