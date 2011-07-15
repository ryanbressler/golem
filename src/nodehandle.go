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
	"json"
	"strconv"
	"fmt"
	"os"
)

type NodeHandle struct {
	NodeId        string
	Uri           string
	Hostname      string
	Master        *Master
	Con           Connection
	MaxJobs       chan int
	Running       chan int
	BroadcastChan chan *WorkerMessage
}

func NewNodeHandle(n *Connection, m *Master) *NodeHandle {
	con := *n
	id := UniqueId()
	nh := NodeHandle{NodeId: id,
		Uri:           "/admin/" + id,
		Hostname:      con.Socket.LocalAddr().String(),
		Master:        m,
		Con:           con,
		MaxJobs:       make(chan int, 1),
		Running:       make(chan int, 1),
		BroadcastChan: make(chan *WorkerMessage, 0)}

	//wait for worker handshake TODO: should this be in monitor???
	nh.Running <- 0
	msg := <-nh.Con.InChan

	if msg.Type == HELLO {
		val, err := strconv.Atoi(msg.Body)
		if err != nil {
			log("error parsing worker hello: %v", err)
			return nil
		}
		nh.MaxJobs <- val
	} else {
		log("%v didn't say hello as first message.", nh.Hostname)
		return nil
	}
	log("%v says hello and asks for %v jobs.", nh.Hostname, msg.Body)
	return &nh
}

func (nh *NodeHandle) MarshalJSON() ([]byte, os.Error) {
	running := <-nh.Running
	nh.Running <- running
	atOnce := <-nh.MaxJobs
	nh.MaxJobs <- atOnce

	rv := fmt.Sprintf("{\"uri\":\"%v\",\"NodeId\":\"%v\",\"Hostname\":\"%v\", \"MaxJobs\":%v,\"Running\":%v}", nh.Uri, nh.NodeId, nh.Hostname, atOnce, running)

	return []byte(rv), nil
}

func (nh *NodeHandle) ReSize(NewMaxJobs int) {
	<-nh.MaxJobs
	nh.MaxJobs <- NewMaxJobs
}

// turns job into JSON and send to connections outbox. Seems to sleep or deadlock if left alone to long so the worker checks-in every 60 seconds.
func (nh *NodeHandle) SendJob(j *Job) {
	job := *j
	log("Sending job %v to %v", job, nh.Hostname)
	jobjson, err := json.Marshal(job)
	if err != nil {
		log("error json.Marshaling job: %v", err)
	}
	msg := WorkerMessage{Type: START, Body: string(jobjson)}
	nh.Con.OutChan <- msg
}

func (nh *NodeHandle) Monitor() {
	//control loop
	for {
		running := <-nh.Running
		nh.Running <- running
		atOnce := <-nh.MaxJobs
		nh.MaxJobs <- atOnce

		switch {
		case running < atOnce:
			vlog("%v has %v running. Waiting for job or message.", nh.Hostname, running)
			select {
			case bcMsg := <-nh.BroadcastChan:
				log("%v sending broadcast message %v", nh.Hostname, *bcMsg)
				nh.Con.OutChan <- *bcMsg
			case job := <-nh.Master.jobChan:
				nh.SendJob(job)
				running := <-nh.Running
				nh.Running <- running + 1
				vlog("%v got job, %v running.", nh.Hostname, running)
			case msg := <-nh.Con.InChan:
				nh.HandleWorkerMessage(&msg)
			}
		default:
			vlog("%v has %v running. Waiting for message.", nh.Hostname, running)
			select {
			case bcMsg := <-nh.BroadcastChan:
				log("%v sending broadcast message %v", nh.Hostname, *bcMsg)
				nh.Con.OutChan <- *bcMsg
			case msg := <-nh.Con.InChan:
				nh.HandleWorkerMessage(&msg)
			}
		}

	}

}

//handle worker messages and updates the value in nh.Running if appropriate
func (nh *NodeHandle) HandleWorkerMessage(msg *WorkerMessage) {
	vlog("Got msg from %v", nh.Hostname)
	switch msg.Type {
	default:
		//cout <- msg.Body
	case CHECKIN:
		vlog("%v checks in", nh.Hostname)
	case COUT:
		vlog("%v got cout", nh.Hostname)
		nh.Master.subMap[msg.SubId].CoutFileChan <- msg.Body
	case CERROR:
		vlog("%v got cerror", nh.Hostname)
		nh.Master.subMap[msg.SubId].CerrFileChan <- msg.Body
	case JOBFINISHED:

		vlog("JOBFINISHED getting running value.")
		running := <-nh.Running
		nh.Running <- running - 1
		log("%v says job finished: %v running: %v", nh.Hostname, msg.Body, running)
		nh.Master.subMap[msg.SubId].FinishedChan <- NewJob(msg.Body)
		vlog("%v finished sent to Sub: %v running: %v", nh.Hostname, msg.Body, running)

	case JOBERROR:
		vlog("JOBERROR getting running value.")
		running := <-nh.Running
		nh.Running <- running - 1
		log("%v says job error: %v running: %v", nh.Hostname, msg.Body, running)
		nh.Master.subMap[msg.SubId].ErrorChan <- NewJob(msg.Body)
		vlog("%v finished sent to Sub: %v running: %v", nh.Hostname, msg.Body, running)

	}

	vlog("%v msg handled", nh.Hostname)
}
