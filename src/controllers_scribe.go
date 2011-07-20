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
	"strconv"
	"time"
	"os"
)

type ScribeJobController struct {
	scribe *Scribe
	proxy  JobController
}

func (c ScribeJobController) RetrieveAll() (json string, err os.Error) {
	log("RetrieveAll")
	items, err := c.scribe.store.All()
	if err != nil {
		return
	}

	jsonArray := make([]string, 0)
	for _, s := range items {
		val, _ := s.MarshalJSON()
		jsonArray = append(jsonArray, string(val))
	}
	json = " { numberOfItems: " + strconv.Itoa(len(jsonArray)) + ", items:[" + strings.Join(jsonArray, ",") + "] }"
	return
}
func (c ScribeJobController) Retrieve(jobId string) (json string, err os.Error) {
	log("Retrieve:%v", jobId)

	item, err := c.scribe.store.Get(jobId)
	if err != nil {
		return
	}

	val, err := item.MarshalJSON()
	if err != nil {
		return
	}

	json = string(val)
	return
}
func (c ScribeJobController) NewJob(r *http.Request) (jobId string, err os.Error) {
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

	err = c.scribe.store.Create(jobPackage)
	return
}

func (c ScribeJobController) Stop(jobId string) os.Error {
	return c.proxy.Stop(jobId)
}
func (c ScribeJobController) Kill(jobId string) os.Error {
	return c.proxy.Kill(jobId)
}
