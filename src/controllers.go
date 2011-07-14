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
	"strings"
	"os"
)

type MasterJobController struct {
	master *Master
}

func (c MasterJobController) RetrieveAll() (json string, numberOfItems int, err os.Error) {
	log("RetrieveAll")

	jsonArray := make([]string, 0)
	for _, s := range c.master.subMap {
		val, _ := s.MarshalJSON()
		jsonArray = append(jsonArray, string(val))
	}
	numberOfItems = len(jsonArray)
	json = strings.Join(jsonArray, ",")
	err = nil
	return
}
func (c MasterJobController) Retrieve(jobId string) (json string, err os.Error) {
	log("Retrieve:%v", jobId)

	job, isin := c.master.subMap[jobId]
	if isin {
		val, err := job.MarshalJSON()
		if err != nil { return }

        json = string(val)
	} else {
        err = os.NewError("job not found")
	}

	return
}
func (c MasterJobController) NewJob(r *http.Request) (jobId string, err os.Error) {
	log("NewJob")

	rJobs := make([]RequestedJob, 0, 100)
	if err = loadJson(r, rJobs); err != nil { return }

	s := NewSubmission(&rJobs, c.master.jobChan)
	jobId = s.SubId
	c.master.subMap[jobId] = s
	log("NewJob: %v", jobId)

	return
}
func (c MasterJobController) Stop(jobId string) os.Error {
	log("Stop:%v", jobId)

	job, isin := c.master.subMap[jobId]
	if isin {
		if job.Stop() {
			return nil
		}
		return os.NewError("unable to stop")
	}
	return os.NewError("job not found")
}
func (c MasterJobController) Kill(jobId string) os.Error {
	log("Kill:%v", jobId)

	job, isin := c.master.subMap[jobId]
	if isin {
		c.master.Broadcast(&clientMsg{Type: KILL, SubId: jobId})
		if job.Stop() {
			return nil
		}
		return os.NewError("unable to stop/kill")
	}
	return os.NewError("job not found")
}

type MasterNodeController struct {
	master *Master
}

func (c MasterNodeController) RetrieveAll() (json string, numberOfItems int, err os.Error) {
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
	node := c.master.NodeHandles[nodeId]
	val, err := node.MarshalJSON()
	json = string(val)
	return
}
func (c MasterNodeController) RestartAll() os.Error {
	log("Restart")

	c.master.Broadcast(&clientMsg{Type: RESTART})
	log("Restarting in 10 seconds.")
	go RestartIn(3000000000)

	return nil
}
func (c MasterNodeController) Resize(nodeId string, numberOfThreads int) os.Error {
	log("Resize:%v,%i", nodeId, numberOfThreads)

	node, isin := c.master.NodeHandles[nodeId]
	if isin {
		node.ReSize(numberOfThreads)
		return nil
	}

	return os.NewError("node not found")
}
func (c MasterNodeController) KillAll() os.Error {
	log("Kill")

	c.master.Broadcast(&clientMsg{Type: DIE})
	log("dying in 10 seconds.")
	go DieIn(3000000000)

	return nil
}
