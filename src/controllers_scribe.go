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
	"url"
)

type ScribeJobController struct {
	store  JobStore
	target *url.URL
	apikey string
}

// GET /jobs
func (this ScribeJobController) Index(rw http.ResponseWriter) {
	logger.Debug("Index()")
	items, err := this.store.All()
	if err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
		return
	}

	jobDetails := JobDetailsList{Items: items, NumberOfItems: len(items)}
	if err := json.NewEncoder(rw).Encode(jobDetails); err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
	}
}
// POST /jobs
func (this ScribeJobController) Create(rw http.ResponseWriter, r *http.Request) {
	logger.Debug("Create()")
	if CheckApiKey(this.apikey, r) == false {
		http.Error(rw, "api key required in header", http.StatusForbidden)
		return
	}

	tasks := make([]Task, 0, 100)
	if err := LoadTasksFromJson(r, &tasks); err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
		return
	}

	jobId := UniqueId()
	owner := GetHeader(r, "x-golem-job-owner", "Anonymous")
	label := GetHeader(r, "x-golem-job-label", jobId)
	jobtype := GetHeader(r, "x-golem-job-type", "Unspecified")

	job := NewJobDetails(jobId, owner, label, jobtype, TotalTasks(tasks))
	if err := this.store.Create(job, tasks); err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
		return
	}
	if err := json.NewEncoder(rw).Encode(job); err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
	}
}
// GET /jobs/id
func (this ScribeJobController) Find(rw http.ResponseWriter, id string) {
	logger.Debug("Find(%v)", id)
	jd, err := this.store.Get(id)
	if err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
		return
	}
	if err := json.NewEncoder(rw).Encode(jd); err != nil {
		http.Error(rw, err.String(), http.StatusBadRequest)
	}
}
// POST /jobs/id/stop or POST /jobs/id/kill
func (this ScribeJobController) Act(rw http.ResponseWriter, parts []string, r *http.Request) {
	logger.Debug("Act(%v):%v", r.URL.Path, parts)
	if CheckApiKey(this.apikey, r) == false {
		http.Error(rw, "api key required in header", http.StatusForbidden)
		return
	}

	if len(parts) < 2 {
		http.Error(rw, "POST /jobs/id/stop or POST /jobs/id/kill", http.StatusBadRequest)
		return
	}

	preq, _ := http.NewRequest(r.Method, r.URL.Path, r.Body)
	preq.Header.Set("x-golem-apikey", this.apikey)
	proxy := http.NewSingleHostReverseProxy(this.target)
	go proxy.ServeHTTP(rw, preq)
}

type ScribeClusterController struct {
	store  JobStore
	target *url.URL
}

// GET /cluster
func (this ScribeClusterController) Index(rw http.ResponseWriter, params url.Values, header http.Header) {
	logger.Debug("Index():[%v,%v]", params, header)

	var numberOfSecondsSince int64 = 0
	value, err := strconv.Atoi(params.Get("numberOfSecondsSince"))
	if err == nil {
		numberOfSecondsSince = int64(value)
	}

	logger.Printf("Index(): %d", numberOfSecondsSince)

	items, err := this.store.ClusterStats(numberOfSecondsSince)
	if err != nil {
		logger.Warn(err)
		http.Error(rw, err.String(), http.StatusInternalServerError)
		return
	}

	clusterStatList := ClusterStatList{Items: items, NumberOfItems: len(items)}
	// TODO: lookup in storage
	if err := json.NewEncoder(rw).Encode(clusterStatList); err != nil {
		logger.Warn(err)
		http.Error(rw, err.String(), http.StatusBadRequest)
	}
}
