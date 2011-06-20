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
/*Node handle contains the struct used by the master to keep track of nodes*/
package main

import (
	"json"
	"strconv"
)


/////////////////////////////////////////////////
//master


//THe master struct contains things that the master node needs but other nodes don't
type NodeHandle struct {
	NodeId string
	Hostname string
	Master * Master
	Con Connection
	MaxJobs chan int
	Running chan int
	BroadcastChan chan *clientMsg
}

func NewNodeHandle(n *Connection, m * Master) * NodeHandle {
	con := *n
	nh := NodeHandle{NodeId:UniqueId(),
		Hostname:con.Socket.LocalAddr().String(),
		Master:m,
		Con:con,
		MaxJobs:make(chan int, 1),
		Running:make(chan int, 1),
		BroadcastChan: make(chan *clientMsg, 0) }

	//wait for client handshake TODO: should this be in monitor???
	nh.Running<-0
	msg := <-nh.Con.InChan

	if msg.Type == HELLO {
		val, err := strconv.Atoi(msg.Body)
		if err != nil {
			log("error parsing client hello: %v", err)
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

//takes a  job, turns job into a json messags and send it into the connections out box.
//This seems to sleep or deadlock if left alone to long so the client checks in every 
//60 seconds.
func (nh * NodeHandle) SendJob( j *Job) {


	job := *j
	log("Sending job %v to %v", job, nh.Hostname)
	jobjson, err := json.Marshal(job)
	if err != nil {
		log("error json.Marshaling job: %v", err)

	}
	msg := clientMsg{Type: START, Body: string(jobjson)}
	nh.Con.OutChan <- msg

}

func (nh * NodeHandle) Monitor() {
	//control loop
	m:= * nh.Master
	for {
		running :=<-nh.Running
		nh.Running<-running
		atOnce :=<-nh.MaxJobs
		nh.MaxJobs<-atOnce
		
		switch {
		case running < atOnce:
			vlog("%v has %v running. Waiting for job or message.", nh.Hostname, running)
			select {
			case bcMsg := <-nh.BroadcastChan:
				log("%v sending broadcast message %v", nh.Hostname, *bcMsg)
				nh.Con.OutChan <- *bcMsg
			case job := <-m.jobChan:
				nh.SendJob(job)
				running :=<-nh.Running
				nh.Running<-running+1
				vlog("%v got job, %v running.", nh.Hostname, running)
			case msg := <-nh.Con.InChan:
				vlog("%v Got msg", nh.Hostname)
				running :=<-nh.Running
				nh.Running <- nh.Master.clientMsgSwitch(nh.Hostname, &msg, running)
				vlog("%v msg handled", nh.Hostname)
			}
		default:
			vlog("%v has %v running. Waiting for message.", nh.Hostname, running)
			select {
			case bcMsg := <-nh.BroadcastChan:
				log("%v sending broadcast message %v", nh.Hostname, *bcMsg)
				nh.Con.OutChan <- *bcMsg
			case msg := <-nh.Con.InChan:
				//log("Got msg from %v", nh.Hostname)
				running :=<-nh.Running
				nh.Running <- nh.Master.clientMsgSwitch(nh.Hostname, &msg, running)
			}
		}

	}

}

