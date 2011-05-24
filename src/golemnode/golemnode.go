/*
   Copyright (C) 2003-2010 Institute for Systems Biology
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
	"fmt"
	"exec"
)

func runCommand(replyc chan int, jobid int, jobcmd string) {
	//make sure the path to the exec is fully qualified
	cmd, err := exec.LookPath(jobcmd)
	if err != nil {
		fmt.Printf("exec %s: %s", jobcmd, err)
	}
	
	jobidarg := fmt.Sprintf("%v", jobid)
	fmt.Printf("Starting job %v\n", jobidarg)
	
	//start the job in test dir pass all stdio back to main.
	//note that cmd has to be the first thing in the args array
	c, err := exec.Run(cmd, []string{cmd, jobidarg}, nil, "/Users/rbressle/Code/golemtest", exec.PassThrough, exec.PassThrough, exec.PassThrough)
	if err != nil {
		fmt.Printf("%v", err)
	}
	
	//wait for the job to finish
	w, err := c.Wait(0)
	if err != nil {
		fmt.Printf("%v", err)
	}

	fmt.Printf("Finishing job %v\n", jobidarg)
	//send signal back to main
	if w.Exited() && w.ExitStatus() == 0 {
		replyc <- jobid
	} else {
		replyc <- -1*jobid
	}

}

func main() {
	//number og jobs needed
	toRun := 10
	//number to run at once
	atOnce := 3
	//number running
	running := 0
	//number that have been run
	run := 0
	finished := 0
	//chanel to see when jobs finish
	replyc := make(chan int)

	jobcmd := "/Users/rbressle/Code/golem/src/python/samplejob.py"

	//control loop
	for {
		switch {
		default: 
			//wait till a job finishes			
			rv :=<-replyc
			fmt.Printf("Got 'done' signal: %v\n",rv)
			finished++
			running--
		case finished >= toRun: 
			//we are done
			return
		case run < toRun && running < atOnce: 
			//start a job
			go runCommand(replyc, run, jobcmd)
			run++
			running++
		}

	}

}
