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
	"time"
	"os"
)

type ScribeJobController struct {
	scribe *Scribe
	store  JobStore
	proxy  JobController
}

func (this ScribeJobController) RetrieveAll() (items []interface{}, err os.Error) {
	log("RetrieveAll")

	retrieved, err := this.store.All()
	if err != nil {
		return
	}

	for _, item := range retrieved {
		items = append(items, item)
	}

	return
}
func (this ScribeJobController) Retrieve(jobId string) (interface{}, os.Error) {
	log("Retrieve:%v", jobId)
	return this.store.Get(jobId)
}
func (this ScribeJobController) NewJob(r *http.Request) (jobId string, err os.Error) {
	tasks := make([]Task, 0, 100)
	if err = loadJson(r, &tasks); err != nil {
		vlog("NewJob:%v", err)
		return
	}

	jobId = UniqueId()
	owner := getHeader(r, "x-golem-job-owner", "Anonymous")
	label := getHeader(r, "x-golem-job-label", jobId)
	now := time.Time{}

	jobPackage := JobPackage{
		Handle: JobHandle{JobId: jobId, Owner: owner, Label: label, FirstCreated: &now, LastModified: &now, Status: JobStatus{}},
		Tasks:  tasks}

	err = this.store.Create(jobPackage)
	return
}

func (c ScribeJobController) Stop(jobId string) os.Error {
	return c.proxy.Stop(jobId)
}
func (c ScribeJobController) Kill(jobId string) os.Error {
	return c.proxy.Kill(jobId)
}
