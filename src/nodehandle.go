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
	"time"
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
	logger.Debug("NewNodeHandle(%v)", n.isWorker)
	con := *n
	id := UniqueId()
	nh := NodeHandle{NodeId: id,
		Uri:           "/nodes/" + id,
		Hostname:      con.Socket.LocalAddr().String(),
		Master:        m,
		Con:           con,
		MaxJobs:       make(chan int, 1),
		Running:       make(chan int, 1),
		BroadcastChan: make(chan *WorkerMessage, 0)}

	//wait for worker handshake TODO: should this be in monitor???
	nh.Running <- 0
	logger.Debug("NewNodeHandle(%v) waiting for first message", n.isWorker)
	msg := <-nh.Con.InChan

	if msg.Type == HELLO {
		logger.Printf("Node Hello Body:%v",msg.Body)
		val, err := NewHelloMsgBody(msg.Body)
		if err != nil {
			logger.Warn(err)
			return nil
		}
		nh.MaxJobs <- val.JobCapacity
		if val.UniqueId != "" {
			nh.NodeId = val.UniqueId
			nh.Uri = "/nodes/" + val.UniqueId
			<-nh.Running
			nh.Running <- val.RunningJobs
		}
		if val.RunningJobs != 0 {
			<-nh.Running
			nh.Running <- val.RunningJobs
		}
	} else {
		logger.Debug("%v didn't say hello as first message.", nh.Hostname)
		return nil
	}
	logger.Debug("%v says hello and asks for %v jobs.", nh.Hostname, msg.Body)
	return &nh
}

func (nh *NodeHandle) Stats() (processes int, running int) {
	//logger.Debug("Stats()")
	running = <-nh.Running
	nh.Running <- running
	processes = <-nh.MaxJobs
	nh.MaxJobs <- processes
	return
}

func (nh *NodeHandle) ReSize(newMaxJobs int) {
	logger.Debug("ReSize(%d)", newMaxJobs)
	<-nh.MaxJobs
	nh.MaxJobs <- newMaxJobs
}

// turns job into JSON and send to connections outbox. Seems to sleep or deadlock if left alone to long so the worker checks-in every 60 seconds.
func (nh *NodeHandle) SendJob(j *WorkerJob) {
	job := *j
	logger.Debug("SendJob(%v): %v", job, nh.Hostname)
	jobjson, err := json.Marshal(job)
	if err != nil {
		logger.Warn(err)
	}
	msg := WorkerMessage{Type: START, Body: string(jobjson)}
	nh.Con.OutChan <- msg
	running := <-nh.Running
	nh.Running <- running + 1
	logger.Debug("assigning [%v, %d]", nh.Hostname, running)
}

func (nh *NodeHandle) Monitor() {
	logger.Debug("Monitor(): [%v]", nh.Hostname)
	//control loop
	for {
		processes, running := nh.Stats()
		//logger.Debug("[%v %d %d]", nh.Hostname, processes, running)

		switch {
		case running < processes:
			//logger.Debug("waiting for job or message [%v, %d]", nh.Hostname, running)
			select {
			case bcMsg := <-nh.BroadcastChan:
				logger.Debug("broadcasting [%v, %v]", nh.Hostname, *bcMsg)
				nh.Con.OutChan <- *bcMsg
			case job := <-nh.Master.jobChan:
				nh.SendJob(job)
			case  <-time.After(1000):
			}
		default:
			//logger.Debug("waiting for message [%v, %v]", nh.Hostname, running)
			select {
			case bcMsg := <-nh.BroadcastChan:
				logger.Debug("broadcasting [%v, %v]", nh.Hostname, *bcMsg)
				nh.Con.OutChan <- *bcMsg
			case <-time.After(1000):
				
			}
		}
	}
}

func (nh *NodeHandle) MonitorIO() {
	logger.Debug("MonitorIO(): [%v]", nh.Hostname)
	for {
		msg := <-nh.Con.InChan
		nh.HandleWorkerMessage(&msg)
	}
}

//handle worker messages and updates the value in nh.Running if appropriate
func (nh *NodeHandle) HandleWorkerMessage(msg *WorkerMessage) {
	//logger.Debug("message from: %v", nh.Hostname)
	switch msg.Type {
	default:
	case CHECKIN:
		logger.Debug("CHECKIN [%v]", nh.Hostname)
	case COUT:
		//logger.Debug("COUT [%v]", nh.Hostname)
		blocked := true
		for blocked==true {
			select {
			case nh.Master.GetSub(msg.SubId).CoutFileChan <- msg.Body:
				blocked = false
			case <-time.After(1 * second):
				logger.Printf("Sending  COUT to subid %v blocked for more then 1 second.", msg.SubId)
			}
		}
		
	case CERROR:
		//logger.Debug("CERROR [%v]", nh.Hostname)
		blocked := true
		for blocked==true {
			select {
			case nh.Master.GetSub(msg.SubId).CerrFileChan <- msg.Body:
				blocked = false
			case <-time.After(1 * second):
				logger.Printf("Sending  CERROR to subid %v blocked for more then 1 second.", msg.SubId)
			}
		}
		
	case JOBFINISHED:
		go func(){
			logger.Debug("JOBFINISHED [%v]", nh.Hostname)
			running := <-nh.Running
			nh.Running <- running - 1
			logger.Debug("JOBFINISHED [%v, %v, %v]", nh.Hostname, msg.Body, running)
			nh.Master.GetSub(msg.SubId).FinishedChan <- NewWorkerJob(msg.Body)
			logger.Printf("JOBFINISHED [%v, %v, %v]", nh.Hostname, msg.Body, running)
		}()
	case JOBERROR:
		go func(){
			logger.Debug("JOBERROR %v", nh.Hostname)
			running := <-nh.Running
			nh.Running <- running - 1
			logger.Debug("JOBERROR running [%v, %v, %v]", nh.Hostname, msg.Body, running)
			nh.Master.GetSub(msg.SubId).ErrorChan <- NewWorkerJob(msg.Body)
			logger.Debug("JOBERROR finished sent: [%v, %v, %v]", nh.Hostname, msg.Body, running)
		}()
	}
	//logger.Debug("message handled: %v", nh.Hostname)
}
