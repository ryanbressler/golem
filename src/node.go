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
	"os"
	"bufio"
	"json"
	"fmt"
	"exec"
)


//////////////////////////////////////////////
//node 


func pipeToChan(p *os.File, msgType int, Id int, ch chan clientMsg) {
	bp := bufio.NewReader(p)

	for {
		line, err := bp.ReadString('\n')
		if err != nil {
			return
		} else {
			ch <- clientMsg{Type: msgType, SubId: Id, Body: line} //string(buffer[0:n])
		}
	}

}

func startJob(cn *Connection, replyc chan int, jsonjob string) {
	var job Job
	con := *cn
	err := json.Unmarshal([]byte(jsonjob), &job)
	if err != nil {
		fmt.Printf("error parseing job json: %v\n", err)
	}
	jobcmd := job.Args[0]
	//make sure the path to the exec is fully qualified
	cmd, err := exec.LookPath(jobcmd)
	if err != nil {
		con.OutChan <- clientMsg{Type: CERROR, SubId: job.SubId, Body: fmt.Sprintf("Error finding %s: %s\n", jobcmd, err)}
		fmt.Printf("exec %s: %s\n", jobcmd, err)
		replyc <- -1 * job.JobId
		return
	}

	args := job.Args[:]
	args = append(args, fmt.Sprintf("%v", job.SubId))
	args = append(args, fmt.Sprintf("%v", job.LineId))
	args = append(args, fmt.Sprintf("%v", job.JobId))

	//start the job in test dir pass all stdio back to main.
	//note that cmd has to be the first thing in the args array
	c, err := exec.Run(cmd, args, nil, "./", exec.DevNull, exec.Pipe, exec.Pipe)
	if err != nil {
		fmt.Printf("%v", err)
	}
	go pipeToChan(c.Stdout, COUT, job.SubId, con.OutChan)
	go pipeToChan(c.Stderr, CERROR, job.SubId, con.OutChan)
	//wait for the job to finish
	w, err := c.Wait(0)
	if err != nil {
		fmt.Printf("joberror:%v", err)
	}

	fmt.Printf("Finishing job %v\n", job.JobId)
	//send signal back to main
	if w.Exited() && w.ExitStatus() == 0 {
		replyc <- job.JobId
	} else {
		replyc <- -1 * job.JobId
	}

}



func RunNode(atOnce int, master string,useTls bool) {
	running := 0
	fmt.Printf("Running as %v process node owned by %v\n", atOnce, master)


	ws, err := wsDialToMaster(master, useTls)
	if err != nil {
		fmt.Printf("Error connectiong to master:%v\n", err)
		return
	}
	//ws.Write([]byte("h"))
	mcon := *NewConnection(ws)
	//go sendCio(&mcon)
	mcon.OutChan <- clientMsg{Type: HELLO, Body: fmt.Sprintf("%v", atOnce)}

	replyc := make(chan int)

	//control loop
	for {

		select {
		case rv := <-replyc:
			fmt.Printf("Got 'done' signal: %v\n", rv)
			mcon.OutChan <- clientMsg{Type: DONE, Body: fmt.Sprintf("%v", rv)}
			running--

		case msg := <-mcon.InChan:
			switch msg.Type {
			case START:
				go startJob(&mcon, replyc, msg.Body)
				running++
			}
		}

	}

}