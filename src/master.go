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
	"sync"
)

type Master struct {
	subMu       sync.RWMutex
	subMap      map[string]*Submission //buffered channel for creating jobs TODO: verify thread safety... should be okay since we only set once
	jobChan     chan *WorkerJob        //buffered channel for creating jobs
	subidChan   chan int               //buffered channel used to keep track of submissions
	nodeMu      sync.RWMutex
	NodeHandles map[string]*NodeHandle
}

//create a master node and initialize its channels
func NewMaster() *Master {
	m := &Master{
		subMap:      map[string]*Submission{},
		jobChan:     make(chan *WorkerJob, 0),
		NodeHandles: map[string]*NodeHandle{}}
	http.Handle("/master/", websocket.Handler(func(ws *websocket.Conn) { m.Listen(ws) }))
	return m
}

func (m *Master) GetSub(subId string) *Submission {
	logger.Debug("GetSub(%v)", subId)
	m.subMu.RLock()
	defer m.subMu.RUnlock()
	return m.subMap[subId]
}

func (m *Master) Listen(ws *websocket.Conn) {
	logger.Printf("Listen(%v): node connecting", ws.LocalAddr().String())
	nh := NewNodeHandle(NewConnection(ws, false), m)
	m.nodeMu.Lock()
	m.NodeHandles[nh.NodeId] = nh
	m.nodeMu.Unlock()
	go m.RemoveNodeOnDeath(nh)
	nh.Monitor()
}

// sends a message to every connected worker
func (m *Master) Broadcast(msg *WorkerMessage) {
	m.nodeMu.RLock()
	logger.Debug("Broadcast(%v): to %v nodes", *msg, len(m.NodeHandles))
	for _, nh := range m.NodeHandles {
		nh.BroadcastChan <- msg
	}
	m.nodeMu.RUnlock()
	logger.Debug("Broadcast(): done")
}

// remove node handles from the map used to store them as they disconnect
func (m *Master) RemoveNodeOnDeath(nh *NodeHandle) {
	logger.Debug("RemoveNodeOnDeath(%v)", nh.NodeId)
	<-nh.Con.DiedChan
	m.nodeMu.Lock()
	m.NodeHandles[nh.NodeId] = nh, false
	m.nodeMu.Unlock()
}
