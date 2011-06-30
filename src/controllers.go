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
	"fmt"
	"strings"
	"os"
)

// Controllers
type MasterJobController struct {
	master *Master
}

type MasterNodeController struct {
	master *Master
}

// Implementations
func (mc MasterJobController) RetrieveAll(r *http.Request) (json string, numberOfItems int, err os.Error) {
	log("RetrieveAll")
	json = "{ items:[], numberOfItems: 0, uri:'/jobs' }"
	numberOfItems = 0
	err = nil
	return
}
func (mc MasterJobController) Retrieve(jobId string) (json string, err os.Error) {
	log("Retrieve:%v", jobId)
	json = fmt.Sprintf("{ items:[], numberOfItems: 0, uri:'/jobs/%v' }", jobId)
	err = nil
	return
}
func (mc MasterJobController) NewJob(r *http.Request) (jobId string, err os.Error) {
	log("NewJob")
	jobId = UniqueId()
	err = nil
	return
}
func (mc MasterJobController) Stop(jobId string) os.Error {
	log("Stop:%v", jobId)
	return os.NewError("unable to stop")
}
func (mc MasterJobController) Kill(jobId string) os.Error {
	log("Kill:%v", jobId)
	return os.NewError("unable to kill")
}

func (c MasterNodeController) RetrieveAll(r *http.Request) (json string, numberOfItems int, err os.Error) {
	log("RetrieveAll")

	numberOfItems = len(c.master.NodeHandles)
	jsonArray := make([]string, 0, numberOfItems)
	for _, n := range c.master.NodeHandles {
		val, _ := n.MarshalJSON()
		jsonArray = append(jsonArray, string(val))
	}
	json = strings.Join(jsonArray, ",")
	err = nil
	return
}
func (c MasterNodeController) Retrieve(nodeId string) (json string, err os.Error) {
	log("Retrieve:%v", nodeId)
	json = fmt.Sprintf("{ items:[], numberOfItems: 0, uri:'/nodes/%v' }", nodeId)
	err = nil
	return
}
func (c MasterNodeController) Restart(nodeId string) os.Error {
	log("Restart:%v", nodeId)

	c.master.Broadcast(&clientMsg{Type: RESTART})
	log("Restarting node %v in 10 seconds.", nodeId)
	go RestartIn(3000000000)

	return nil
}
func (c MasterNodeController) Resize(nodeId string, numberOfThreads int) os.Error {
	log("Resize:%v,%i", nodeId, numberOfThreads)
	return os.NewError("unable to resize")

	node, isin := c.master.NodeHandles[nodeId]
	if isin {
		node.ReSize(numberOfThreads)
		return nil
	}

	return os.NewError("node not found")
}
func (c MasterNodeController) Kill(nodeId string) os.Error {
	log("Kill:%v", nodeId)

	c.master.Broadcast(&clientMsg{Type: DIE})
	log("Node %v dying in 10 seconds.", nodeId)
	go DieIn(3000000000)

	return nil
}
