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
	"json"
	"strconv"
)

type MasterJobController struct {
	master *Master
	apikey string
}
// GET /jobs
func (this MasterJobController) Index(rw http.ResponseWriter) {
	vlog("MasterJobController.Index()")
	items := make([]JobDetails, 0, 0)

	vlog("MasterJobController.Index(): for loop")
	this.master.subMu.RLock()
	for _, s := range this.master.subMap {
		items = append(items, s.SniffDetails())
	}
	this.master.subMu.RUnlock()
	vlog("MasterJobController.Index(): for loop done")

	jobDetails := JobDetailsList{Items: items, NumberOfItems: len(items)}
	if err := json.NewEncoder(rw).Encode(jobDetails); err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
	}
}
// POST /jobs
func (this MasterJobController) Create(rw http.ResponseWriter, r *http.Request) {
	vlog("MasterJobController.Create()")
	if CheckApiKey(this.apikey, r) == false {
		http.Error(rw, "api key required in header", http.StatusForbidden)
		return
	}

	tasks := make([]Task, 0, 100)
	if err := loadJson(r, &tasks); err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
		return
	}

	jobId := getHeader(r, "x-golem-job-preassigned-id", "")
	if jobId == "" {
		jobId = UniqueId()
	}

	this.master.subMu.RLock()
	_, isin := this.master.subMap[jobId]
	this.master.subMu.RUnlock()
	if isin {
		vlog("MasterJobController.Create(): Exists: %v", jobId)
		return
	}

	owner := getHeader(r, "x-golem-job-owner", "Anonymous")
	label := getHeader(r, "x-golem-job-label", jobId)
	jobtype := getHeader(r, "x-golem-job-type", "Unspecified")

	jd := NewJobDetails(jobId, owner, label, jobtype, TotalTasks(tasks))

	vlog("MasterJobController.Create(): creating: %v", jobId)
	this.master.subMu.Lock()
	this.master.subMap[jobId] = NewSubmission(jd, tasks, this.master.jobChan)
	this.master.subMu.Unlock()
	vlog("MasterJobController.Create(): created: %v", jobId)

	if err := json.NewEncoder(rw).Encode(jd); err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
	}
}
// GET /jobs/id
func (this MasterJobController) Find(rw http.ResponseWriter, id string) {
	vlog("MasterJobController.Find(%v)", id)
	this.master.subMu.RLock()
	s, isin := this.master.subMap[id]
	this.master.subMu.RUnlock()

	if isin == false {
		vlog("MasterJobController.Find(%v): not found", id)
		http.Error(rw, "job "+id+" not found", http.StatusNotFound)
		return
	}
	vlog("MasterJobController.Find(%v): found", id)
	if err := json.NewEncoder(rw).Encode(s.SniffDetails()); err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
	}
}
// POST /jobs/id/stop or POST /jobs/id/kill
func (this MasterJobController) Act(rw http.ResponseWriter, parts []string, r *http.Request) {
	vlog("MasterJobController.Act(%v)", r.URL.Path)
	if CheckApiKey(this.apikey, r) == false {
		http.Error(rw, "api key required in header", http.StatusForbidden)
		return
	}

	if len(parts) < 2 {
		http.Error(rw, "POST /jobs/id/stop or POST /jobs/id/kill", http.StatusBadRequest)
		return
	}

	jobId := parts[0]
	vlog("MasterJobController.Act(%v): Finding", jobId)
	this.master.subMu.RLock()
	job, isin := this.master.subMap[jobId]
	this.master.subMu.RUnlock()
	if isin == false {
		vlog("MasterJobController.Act(%v): Not Found", jobId)
		http.Error(rw, "job "+jobId+" not found", http.StatusNotFound)
		return
	}

	if parts[1] == "stop" {
		vlog("MasterJobController.Act(%v): Stopping", jobId)
		if job.Stop() == false {
			http.Error(rw, "unable to stop", http.StatusExpectationFailed)
		}
	} else if parts[1] == "kill" {
		vlog("MasterJobController.Act(%v): Killing", jobId)
		if job.Stop() == false {
			http.Error(rw, "unable to stop", http.StatusExpectationFailed)
		}
		this.master.Broadcast(&WorkerMessage{Type: KILL, SubId: jobId})
	}
	vlog("MasterJobController.Act(): Returning")
}

type MasterNodeController struct {
	master *Master
	apikey string
}
// GET /nodes
func (this MasterNodeController) Index(rw http.ResponseWriter) {
	vlog("MasterNodeController.Index()")
	items := make([]WorkerNode, 0, 0)
	vlog("MasterNodeController.Index(): for loop")
	this.master.nodeMu.RLock()
	for _, n := range this.master.NodeHandles {
		items = append(items, NewWorkerNode(n))
	}
	this.master.nodeMu.RUnlock()
	vlog("MasterNodeController.Index(): for loop done")
	workerNodes := WorkerNodeList{Items: items, NumberOfItems: len(items)}
	if err := json.NewEncoder(rw).Encode(workerNodes); err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
	}
}
// GET /nodes/id
func (this MasterNodeController) Find(rw http.ResponseWriter, nodeId string) {
	vlog("MasterNodeController.Find(%v)", nodeId)
	this.master.nodeMu.RLock()
	nh, isin := this.master.NodeHandles[nodeId]
	this.master.nodeMu.RUnlock()
	if isin == false {
		vlog("MasterNodeController.Find(%v): not found", nodeId)
		http.Error(rw, "node "+nodeId+" not found", http.StatusNotFound)
		return
	}

	vlog("MasterNodeController.Find(%v): found", nodeId)
	if err := json.NewEncoder(rw).Encode(NewWorkerNode(nh)); err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
	}
}
// POST /nodes/restart or POST /nodes/die or POST /nodes/id/resize/new-size
func (this MasterNodeController) Act(rw http.ResponseWriter, parts []string, r *http.Request) {
	vlog("MasterNodeController.Act(%v):%v", r.URL.Path, parts)

	if CheckApiKey(this.apikey, r) == false {
		http.Error(rw, "api key required in header", http.StatusForbidden)
		return
	}

	if parts[0] == "restart" {
		this.master.Broadcast(&WorkerMessage{Type: RESTART})
		go RestartIn(10)
		return
	}

	if parts[0] == "die" {
		this.master.Broadcast(&WorkerMessage{Type: DIE})
		go DieIn(10)
		return
	}

	if parts[1] == "resize" {
		nodeId := parts[0]
		numberOfThreads, err := strconv.Atoi(parts[2])
		if err != nil {
			http.Error(rw, err.String(), http.StatusBadRequest)
			return
		}

		this.master.nodeMu.RLock()
		node, isin := this.master.NodeHandles[nodeId]
		this.master.nodeMu.RUnlock()

		if isin == false {
			http.Error(rw, "node "+nodeId+" not found", http.StatusNotFound)
			return
		}

		node.ReSize(numberOfThreads)
	}
}
