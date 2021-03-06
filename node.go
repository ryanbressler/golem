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
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

func PipeToChan(r io.Reader, msgType int, id string, ch chan WorkerMessage, done chan int, prepend string) {
	logger.Debug("PipeToChan(%d,%v)", msgType, id)
	bp := bufio.NewReader(r)
	for {

		line, prefix, err := bp.ReadLine()
		linestr := string(line)
		for prefix != false { //this should almost never happen
			line, prefix, err = bp.ReadLine()
			linestr = linestr + string(line)
		}

		if linestr != "" {
			blocked := true
			for blocked == true {
				select {
				case ch <- WorkerMessage{Type: msgType, SubId: id, Body: prepend + linestr + "\n"}:
					blocked = false
				case <-time.After(time.Second):
					logger.Printf("WARNING PipeToChan() has been blocked for more then 1 second MsgType:%d,id:%v", msgType, id)
				}
			}

		}
		switch {
		case err == io.EOF:
			done <- 1
			return
		case err != nil:
			done <- 1
			logger.Warn(err)
			return
		}
	}
	done <- 1

}

func StartJob(cn *Connection, replyc chan *WorkerMessage, jsonjob string, jk *JobKiller) {
	logger.Debug("StartJob(%v)", jsonjob)
	con := *cn

	job := NewWorkerJob(jsonjob)
	jobcmd := job.Args[0]
	//make sure the path to the exec is fully qualified
	exepath, err := exec.LookPath(jobcmd)
	if err != nil {
		con.OutChan <- WorkerMessage{Type: CERROR, SubId: job.SubId, Body: fmt.Sprintf("Error finding %s: %s\n", jobcmd, err)}
		logger.Printf("exec %s: %s\n", jobcmd, err)
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
		logger.Warn(err)
		return
	}
	coutchan := make(chan int, 0)
	go PipeToChan(outpipe, COUT, job.SubId, con.OutChan, coutchan, "")

	errpipe, err := cmd.StderrPipe()
	if err != nil {
		logger.Warn(err)
		return
	}
	cerrorchan := make(chan int, 0)
	go PipeToChan(errpipe, CERROR, job.SubId, con.OutChan, cerrorchan, "TASK : \""+exepath+strings.Join(args, " ")+"\" ERRORED: \n")

	if err = cmd.Start(); err != nil {
		logger.Warn(err)
		replyc <- &WorkerMessage{Type: JOBERROR, SubId: job.SubId, Body: jsonjob, ErrMsg: err.Error()}
		return
	}

	kb := &Killable{Pid: cmd.Process.Pid, SubId: job.SubId, JobId: job.JobId}
	jk.Registerchan <- kb
	defer func() {
		jk.Donechan <- kb
	}()

	<-coutchan
	<-cerrorchan
	if err = cmd.Wait(); err != nil {
		logger.Warn(err)
		replyc <- &WorkerMessage{Type: JOBERROR, SubId: job.SubId, Body: jsonjob, ErrMsg: err.Error()}
		return
	}

	logger.Printf("finishing job %v", job.JobId)
	replyc <- &WorkerMessage{Type: JOBFINISHED, SubId: job.SubId, Body: jsonjob}
}

func CheckIn(c *Connection) {
	logger.Debug("CheckIn(%v)", c.isWorker)
	con := *c
	for {
		<-time.After(time.Duration(60) * time.Second)
		logger.Debug("CheckIn(%v) after sleep", c.isWorker)
		con.OutChan <- WorkerMessage{Type: CHECKIN}
	}
}

func RunNode(processes int, master string) {
	running := 0

	jk := NewJobKiller()
	logger.Debug("Running as %d process node owned by %v", processes, master)

	ws := OpenWebSocketToMaster(master)

	mcon := *NewConnection(ws, true)
	wm := WorkerMessage{Type: HELLO}
	wm.BodyFromInterface(HelloMsgBody{JobCapacity: processes, RunningJobs: 0})
	logger.Printf("Hello msg body: %v", wm.Body)
	mcon.OutChan <- wm
	go CheckIn(&mcon)
	replyc := make(chan *WorkerMessage)

	for {
		logger.Debug("Waiting for done or msg.")
		select {
		case <-mcon.DiedChan:
			wm = WorkerMessage{Type: HELLO}
			wm.BodyFromInterface(HelloMsgBody{JobCapacity: processes, RunningJobs: running})
			mcon.ReConChan <- wm
		case rv := <-replyc:
			logger.Debug("Got 'done' signal")
			mcon.OutChan <- *rv
			running--

		case msg := <-mcon.InChan:
			logger.Debug("Got master msg")
			switch msg.Type {
			case START:
				logger.Printf("START")
				go StartJob(&mcon, replyc, msg.Body, jk)
				running++
			case KILL:
				logger.Printf("KILL: %v", msg.SubId)
				jk.Killchan <- msg.SubId
			case RESTART:
				logger.Printf("RESTART: %v", msg.SubId)
				RestartIn(8)
			case DIE:
				logger.Printf("DIE: %v", msg.SubId)
				DieIn(0)
			}
		}
	}

}
