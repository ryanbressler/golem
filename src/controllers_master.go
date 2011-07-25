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
	"os"
)

type MasterJobController struct {
	master *Master
}

func (c MasterJobController) RetrieveAll() (items []interface{}, err os.Error) {
	log("MasterJobController.RetrieveAll")

	for _, s := range c.master.subMap {
		items = append(items, s.Details)
	}

	return
}
func (c MasterJobController) Retrieve(jobId string) (item interface{}, err os.Error) {
	log("Retrieve:%v", jobId)

	s, isin := c.master.subMap[jobId]
	if isin == false {
		err = os.NewError("job " + jobId + " not found")
		return
	}

	item = s.Details
	return
}
func (c MasterJobController) NewJob(r *http.Request) (jobId string, err os.Error) {
	tasks := make([]Task, 0, 100)
	if err = loadJson(r, &tasks); err != nil {
		vlog("MasterJobController.NewJob: %v", err)
		return
	}

    jobId = getHeader(r, "x-golem-job-preassigned-id", "")
    if jobId == "" { jobId = UniqueId() }

	owner := getHeader(r, "x-golem-job-owner", "Anonymous")
	label := getHeader(r, "x-golem-job-label", jobId)
	jobtype := getHeader(r, "x-golem-job-type", "Unspecified")

    jd := NewJobDetails(jobId, owner, label, jobtype, tasks)

	c.master.subMap[jobId] = NewSubmission(jd, c.master.jobChan)
	log("NewJob: %v", jd.JobId)

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
		c.master.Broadcast(&WorkerMessage{Type: KILL, SubId: jobId})
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

func (c MasterNodeController) RetrieveAll() (items []interface{}, err os.Error) {
	log("MasterJobController.RetrieveAll")

	for _, n := range c.master.NodeHandles {
		items = append(items, NewWorkerNode(n))
	}
	return
}
func (c MasterNodeController) Retrieve(nodeId string) (item interface{}, err os.Error) {
	log("Retrieve:%v", nodeId)
	nh, isin := c.master.NodeHandles[nodeId]
	if isin == false {
		err = os.NewError("node " + nodeId + " not found")
		return
	}
	item = NewWorkerNode(nh)
	return
}
func (c MasterNodeController) RestartAll() os.Error {
	log("Restart")

	c.master.Broadcast(&WorkerMessage{Type: RESTART})
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

	c.master.Broadcast(&WorkerMessage{Type: DIE})
	log("dying in 10 seconds.")
	go DieIn(3000000000)

	return nil
}
