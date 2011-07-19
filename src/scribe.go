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
	"os"
	"time"
)

type Scribe struct {
	store         JobStore
	masterJobsUrl string
}

func NewScribe(store JobStore) *Scribe {
	target, err := ConfigFile.GetString("scribe", "target")
	if err != nil {
		panic(err)
	}

	s := Scribe{store: store, masterJobsUrl: target + "/jobs/"}

	ticker := time.NewTicker(3 * second)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.PollJobs()
			}
		}
	}()

	return &s
}

func (this *Scribe) PollJobs() {
	for _, jobHandle := range this.GetJobs() {
		this.store.Update(jobHandle.JobId, jobHandle.Status)
	}

	unscheduled, _ := this.store.Unscheduled()
	for _, jobHandle := range unscheduled {
		this.PostJob(jobHandle)
	}
}

func (this *Scribe) GetJobs() []JobHandle {
	resp, _, err := http.Get(this.masterJobsUrl)
	if err != nil {
		log("Error [%v]: %v", this.masterJobsUrl, err)
		return nil
	}

	b := make([]byte, 100)
	resp.Body.Read(b)

	jobHandles := make([]JobHandle, 100)
	json.Unmarshal(b, &jobHandles)
	return jobHandles
}

func (this *Scribe) PostJob(jobHandle JobHandle) (err os.Error) {
    log("PostJob(%v)", jobHandle)

    jobPkg, err := this.store.Get(jobHandle.JobId)
    if err != nil {
        return
    }

    taskJson, err := json.Marshal(jobPkg.Tasks)
    if err != nil {
        return
    }

    data := make(map[string]string)
    data["jsonfile"] = string(taskJson)

    header := http.Header{}
    header.Set("x-golem-job-preassigned-id", jobHandle.JobId)
    http.PostForm(this.masterJobsUrl, data)

	return
}
