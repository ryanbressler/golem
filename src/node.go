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
	"io"
	"bufio"
	"fmt"
	"exec"
	"time"
)

func pipeToChan(r io.Reader, msgType int, Id string, ch chan WorkerMessage) {
	bp := bufio.NewReader(r)

	for {
		line, err := bp.ReadString('\n')
		if err != nil {
			return
		} else {
			ch <- WorkerMessage{Type: msgType, SubId: Id, Body: line} //string(buffer[0:n])
		}
	}

}


func startJob(cn *Connection, replyc chan *WorkerMessage, jsonjob string, jk *JobKiller) {
	log("Starting job from json: %v", jsonjob)
	con := *cn

	job := NewJob(jsonjob)
	jobcmd := job.Args[0]
	//make sure the path to the exec is fully qualified
	exepath, err := exec.LookPath(jobcmd)
	if err != nil {
		con.OutChan <- WorkerMessage{Type: CERROR, SubId: job.SubId, Body: fmt.Sprintf("Error finding %s: %s\n", jobcmd, err)}
		log("exec %s: %s\n", jobcmd, err)
		replyc <- &WorkerMessage{Type: JOBERROR, SubId: job.SubId, Body: jsonjob}
		return
	}

	args := job.Args[1:]
	args = append(args, fmt.Sprintf("%v", job.SubId))
	args = append(args, fmt.Sprintf("%v", job.LineId))
	args = append(args, fmt.Sprintf("%v", job.JobId))

	//start the job in test dir pass all stdio back to main.  note that cmd has to be the first thing in the args array
	cmd := exec.Command(exepath, args...)

	outpipe, err := cmd.StdoutPipe()
	if err != nil {
		log("startJob.StdoutPipe:%v", err)
		return
	}
	errpipe, err := cmd.StderrPipe()
	if err != nil {
		log("startJob.StderrPipe:%v", err)
		return
	}

	go pipeToChan(outpipe, COUT, job.SubId, con.OutChan)
	go pipeToChan(errpipe, CERROR, job.SubId, con.OutChan)

	err = cmd.Start()
	if err != nil {
		log("startJob.Start:%v", err)
		replyc <- &WorkerMessage{Type: JOBERROR, SubId: job.SubId, Body: jsonjob, ErrMsg: err.String()}
		return
	}

	kb := &Killable{Pid: cmd.Process.Pid, SubId: job.SubId, JobId: job.JobId}
	jk.Registerchan <- kb
	defer func() {
		jk.Donechan <- kb
	}()

	//wait for the job to finish
	err = cmd.Wait()
	if err != nil {
		log("joberror:%v", err)
		replyc <- &WorkerMessage{Type: JOBERROR, SubId: job.SubId, Body: jsonjob, ErrMsg: err.String()}
		return
	}

	log("Finishing job %v", job.JobId)
	replyc <- &WorkerMessage{Type: JOBFINISHED, SubId: job.SubId, Body: jsonjob}
}

func CheckIn(c *Connection) {
	con := *c
	for {
		time.Sleep(60000000000)
		con.OutChan <- WorkerMessage{Type: CHECKIN}

	}
}


func RunNode(processes int, master string) {
	running := 0
	jk := NewJobKiller()
	log("Running as %v process node owned by %v", processes, master)

	ws, err := wsDialToMaster(master, useTls)
	if err != nil {
		log("Error connecting to master:%v", err)
		panic(err)
	}

	mcon := *NewConnection(ws, true)
	mcon.OutChan <- WorkerMessage{Type: HELLO, Body: fmt.Sprintf("%v", processes)}
	go CheckIn(&mcon)
	replyc := make(chan *WorkerMessage)

	for {
		log("Waiting for done or msg.")
		select {
		case rv := <-replyc:
			log("Got 'done' signal: %v", *rv)
			mcon.OutChan <- *rv
			running--

		case msg := <-mcon.InChan:
			log("Got master msg")
			switch msg.Type {
			case START:
				go startJob(&mcon, replyc, msg.Body, jk)
				running++
			case KILL:
				log("Got KILL message for subit : %v", msg.SubId)
				jk.Killchan <- msg.SubId
			case RESTART:
				log("Got restart message: %v", msg)
				RestartIn(8000000000)
			case DIE:
				log("Got die message: %v", msg)
				DieIn(0)
			}
		}
	}

}
