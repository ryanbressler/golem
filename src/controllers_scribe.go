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

type ScribeJobController struct {
	store  JobStore
	proxy  JobController
}

func (this ScribeJobController) RetrieveAll() (items []interface{}, err os.Error) {
	retrieved, err := this.store.All()
	if err == nil {
        for _, item := range retrieved {
            items = append(items, item)
        }
	}
	return
}
func (this ScribeJobController) Retrieve(jobId string) (interface{}, os.Error) {
	return this.store.Get(jobId)
}
func (this ScribeJobController) NewJob(r *http.Request) (jobId string, err os.Error) {
	tasks := make([]Task, 0, 100)
	if err = loadJson(r, &tasks); err != nil {
		vlog("ScribeJobController.NewJob:%v", err)
		return
	}

	jobId = UniqueId()
	owner := getHeader(r, "x-golem-job-owner", "Anonymous")
	label := getHeader(r, "x-golem-job-label", jobId)
	jobtype := getHeader(r, "x-golem-job-type", "Unspecified")

    job := JobDetails{
        Identity: Identity{ JobId: jobId, Uri: "/jobs/" + jobId },
        Description: Description{Owner: owner, Label: label, Type: jobtype},
        Status: InitialStatus(), Progress: InitialProgress(tasks), Timing: InitialTiming(),
        Tasks: tasks }
	err = this.store.Create(job)
	return
}

func (c ScribeJobController) Stop(jobId string) os.Error {
	return c.proxy.Stop(jobId)
}
func (c ScribeJobController) Kill(jobId string) os.Error {
	return c.proxy.Kill(jobId)
}
