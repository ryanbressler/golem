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
)

type ScribeJobController struct {
	store JobStore
	proxy ProxyJobController
}

// GET /jobs
func (this ScribeJobController) Index(rw http.ResponseWriter) {
    items, err := this.store.All();
	if err == nil {
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
	tasks := make([]Task, 0, 100)
	if err := loadJson(r, &tasks); err != nil {
        http.Error(rw, err.String(), http.StatusBadRequest)
		return
	}

	jobId := UniqueId()
	owner := getHeader(r, "x-golem-job-owner", "Anonymous")
	label := getHeader(r, "x-golem-job-label", jobId)
	jobtype := getHeader(r, "x-golem-job-type", "Unspecified")

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
    if len(parts) < 2 {
        http.Error(rw, "POST /jobs/id/stop or POST /jobs/id/kill", http.StatusBadRequest)
        return
    }

    jobId := parts[0]
    if parts[1] == "stop" {
        if err := this.proxy.Stop(jobId); err != nil {
            http.Error(rw, err.String(), http.StatusBadRequest)
        }
    } else if parts[1] == "kill" {
        if err := this.proxy.Kill(jobId); err != nil {
            http.Error(rw, err.String(), http.StatusBadRequest)
        }
    }
}
