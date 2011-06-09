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

func startJob(cn *Connection, replyc chan *clientMsg, jsonjob string) {
	log("Starting job from json: %v", jsonjob)
	con := *cn

	job := NewJob(jsonjob)
	jobcmd := job.Args[0]
	//make sure the path to the exec is fully qualified
	cmd, err := exec.LookPath(jobcmd)
	if err != nil {
		con.OutChan <- clientMsg{Type: CERROR, SubId: job.SubId, Body: fmt.Sprintf("Error finding %s: %s\n", jobcmd, err)}
		log("exec %s: %s\n", jobcmd, err)
		replyc <- &clientMsg{Type: JOBERROR, SubId: job.SubId, Body: jsonjob}
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
		log("%v", err)
	}
	go pipeToChan(c.Stdout, COUT, job.SubId, con.OutChan)
	go pipeToChan(c.Stderr, CERROR, job.SubId, con.OutChan)
	//wait for the job to finish
	w, err := c.Wait(0)
	if err != nil {
		log("joberror:%v", err)
	}

	log("Finishing job %v", job.JobId)
	//send signal back to main
	if w.Exited() && w.ExitStatus() == 0 {

		replyc <- &clientMsg{Type: JOBFINISHED, SubId: job.SubId, Body: jsonjob}
	} else {
		replyc <- &clientMsg{Type: JOBERROR, SubId: job.SubId, Body: jsonjob}
	}

}


func RunNode(atOnce int, master string) {
	running := 0
	log("Running as %v process node owned by %v", atOnce, master)

	ws, err := wsDialToMaster(master, useTls)
	if err != nil {
		log("Error connectiong to master:%v", err)
		return
	}
	//ws.Write([]byte("h"))
	mcon := *NewConnection(ws)
	//go sendCio(&mcon)
	mcon.OutChan <- clientMsg{Type: HELLO, Body: fmt.Sprintf("%v", atOnce)}

	replyc := make(chan *clientMsg)

	//control loop
	for {

		select {
		case rv := <-replyc:
			log("Got 'done' signal: %v", *rv)
			mcon.OutChan <- *rv
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
