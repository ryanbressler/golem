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
	"http"
	"websocket"
)


/////////////////////////////////////////////////
//master


//THe master struct contains things that the master node needs but other nodes don't
type Master struct {
	subMap      map[string]*Submission //buffered channel for creating jobs TODO: verify thread safety... should be okay since we only set once
	jobChan     chan *Job              //buffered channel for creating jobs
	subidChan   chan int               //buffered channel for use as an incrementer to keep track of submissions
	NodeHandles map[string]*NodeHandle
}

//create a master node and initalize its channels
func NewMaster() *Master {
	m := Master{
		subMap:      map[string]*Submission{},
		jobChan:     make(chan *Job, 0),
		NodeHandles: map[string]*NodeHandle{}}
    http.Handle("/master/", websocket.Handler(func(ws *websocket.Conn) { m.Listen(ws) }))
	return &m
}

func (m *Master) Listen(ws *websocket.Conn) {
        log("Node connecting from %v.", ws.LocalAddr().String())

        nh := NewNodeHandle(NewConnection(ws), m)
        m.NodeHandles[nh.NodeId] = nh
        go m.RemoveNodeOnDeath(nh)
        nh.Monitor()
}

//Broadcast sends a message to ever connected client
func (m *Master) Broadcast(msg *clientMsg) {
	log("Broadcasting message %v to %v nodes.", *msg, len(m.NodeHandles))
	for _, nh := range m.NodeHandles {
		nh.BroadcastChan <- msg
	}
	vlog("Broadcasting done.")

}

//goroutine to remove nodehandles from the map used to store them as they disconect
func (m *Master) RemoveNodeOnDeath(nh *NodeHandle) {
	<-nh.Con.DiedChan
	m.NodeHandles[nh.NodeId] = nh, false
}
