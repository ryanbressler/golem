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

func RestartIn(waitn int64) {
	log("Restart in %v nanoseconds", waitn)
	time.Sleep(waitn)
	log("Restarting.")
	_, err := exec.Run(os.Args[0], os.Args, nil, "./", exec.DevNull, exec.PassThrough, exec.PassThrough)
	if err != nil {
		log("%v", err)
	}
	log("Exiting.")
	os.Exit(0)
}

func DieIn(waitn int64) {
	log("Die in %v nanoseconds", waitn)
	time.Sleep(waitn)
	log("Exiting.")
	os.Exit(0)
}