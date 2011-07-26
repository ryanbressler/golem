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
	"strings"
	"time"
)

type Scribe struct {
	store         JobStore
	masterJobsUrl string
	apikey string
}

func LaunchScribe(store JobStore) {
	target := ConfigFile.GetRequiredString("scribe", "target")
	apikey := ConfigFile.GetRequiredString("default", "password")
	s := Scribe{store: store, masterJobsUrl: target + "/jobs/", apikey: apikey}

	for {
		s.PollJobs()
		time.Sleep(10 * second)
	}
}

func (this *Scribe) PollJobs() {
	vlog("Scribe.PollJobs")
	for _, jd := range this.GetJobs() {
		this.store.Update(jd)
	}

	unscheduled, _ := this.store.Unscheduled()
	vlog("Scribe.PollJobs:unsheduled:%d", len(unscheduled))
	for _, u := range unscheduled {
		this.PostJob(u)
	}
}

func (this *Scribe) GetJobs() []JobDetails {
	vlog("Scribe.GetJobs()")
	resp, err := http.Get(this.masterJobsUrl)
	if err != nil {
		vlog("Scribe.GetJobs:%v", err)
		return nil
	}

	b := make([]byte, 100)
	resp.Body.Read(b)

	lst := JobDetailsList{}
	json.Unmarshal(b, &lst)
	return lst.Items
}

func (this *Scribe) PostJob(jd JobDetails) (err os.Error) {
	log("Scribe.PostJob(%v)", jd.JobId)

	tasks, err := this.store.Tasks(jd.JobId)
	if err != nil {
		vlog("Scribe.PostJob(%v):%v", jd.JobId, err)
		return
	}

	taskJson, err := json.Marshal(tasks)
	if err != nil {
		vlog("Scribe.PostJob(%v):%v", jd.JobId, err)
		return
	}

	data := make(http.Values)
	data.Set("jsonfile", string(taskJson))

	r, err := http.NewRequest("POST", this.masterJobsUrl, strings.NewReader(data.Encode()))
	if err != nil {
	    return
	}

	r.Header.Set("x-golem-job-preassigned-id", jd.JobId)
	r.Header.Set("x-golem-apikey", this.apikey)

    client := http.Client{}
    resp, err := client.Do(r)

    log("Client.Do:%v;%v",resp, err)

	return
}
