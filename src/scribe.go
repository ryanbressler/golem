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
	"fmt"
	"http"
	"os"
	"io"
	"json"
	"strings"
	"time"
	"mime/multipart"
)

type Scribe struct {
	store     JobStore
	masterUrl string
	apikey    string
}

func LaunchScribe(store JobStore, target string, apikey string) {
	s := Scribe{store: store, masterUrl: target, apikey: apikey}

	for {
		s.PollJobs()
		time.Sleep(10 * second)
	}
}

func MonitorClusterStats(store JobStore, target string, numberOfSeconds int64) {
	s := Scribe{store: store, masterUrl: target}

	for {
		time.Sleep(numberOfSeconds * second)

		workerNodes, err := s.GetWorkerNodes()
		if err != nil {
			logger.Warn(err)
			continue
		}

		totalJobsRunning, _ := s.store.CountActive()
		totalJobsPending, _ := s.store.CountPending()
		logger.Debug("MonitorClusterStats jobs:[%d,%d]", totalJobsRunning, totalJobsPending)

		totalWorkersRunning := 0
		totalWorkersAvailable := 0
		for _, wn := range workerNodes {
			totalWorkersRunning += wn.RunningJobs
			totalWorkersAvailable += (wn.MaxJobs - wn.RunningJobs)
		}

		logger.Debug("MonitorClusterStats [%d,%d,%d,%d]", totalJobsRunning, totalJobsPending, totalWorkersRunning, totalWorkersAvailable)
		clusterStat := NewClusterStat(totalJobsRunning, totalJobsPending, totalWorkersRunning, totalWorkersAvailable)
		logger.Debug("clusterStat=%v", clusterStat)

		s.store.SnapshotCluster(clusterStat)
	}
}

func (this *Scribe) PollJobs() {
	logger.Debug("PollJobs")
	for _, jd := range this.GetJobs() {
		this.store.Update(jd)
		this.ArchiveJob(jd)
	}

	unscheduled, _ := this.store.Unscheduled()
	logger.Debug("unscheduled=%d", len(unscheduled))
	for _, u := range unscheduled {
		this.PostJob(u)
	}
}

func (this *Scribe) GetJobs() []JobDetails {
	logger.Debug("GetJobs()")
	resp, err := http.Get(this.masterUrl + "/jobs/")
	if err != nil {
		return nil
	}

	rb := resp.Body
	defer rb.Close()
	lst := JobDetailsList{Items: make([]JobDetails, 0, 0)}
	if err = json.NewDecoder(rb).Decode(&lst); err != nil {
		logger.Warn(err)
	}
	return lst.Items
}

func (this *Scribe) GetWorkerNodes() (items []WorkerNode, err os.Error) {
	logger.Debug("GetWorkerNodes()")
	resp, err := http.Get(this.masterUrl + "/nodes")
	if err != nil {
		logger.Warn(err)
		return
	}

	rb := resp.Body
	defer rb.Close()

	workerNodes := WorkerNodeList{Items: make([]WorkerNode, 0, 0)}
	if err := json.NewDecoder(rb).Decode(&workerNodes); err != nil {
		logger.Warn(err)
		return
	}

	items = workerNodes.Items
	logger.Debug("GetWorkerNodes():%d", len(items))
	return
}

func (this *Scribe) PostJob(jd JobDetails) (err os.Error) {
	logger.Debug("PostJob(%v)", jd.JobId)

	tasks, err := this.store.Tasks(jd.JobId)
	if err != nil {
		logger.Warn(err)
		return
	}

	logger.Debug("PostJob(%v):tasks=%d", jd.JobId, len(tasks))
	preader, pwriter := io.Pipe()

	r, err := http.NewRequest("POST", this.masterUrl+"/jobs/", preader)
	if err != nil {
		logger.Warn(err)
		return
	}

	multipartWriter := multipart.NewWriter(pwriter)

	r.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	r.Header.Set("x-golem-job-preassigned-id", jd.JobId)
	r.Header.Set("x-golem-apikey", this.apikey)

	if jd.Owner != "" {
		r.Header.Set("x-golem-job-owner", jd.Owner)
	}
	if jd.Label != "" {
		r.Header.Set("x-golem-job-label", jd.Label)
	}
	if jd.Type != "" {
		r.Header.Set("x-golem-job-type", jd.Type)
	}

	go func() {
		logger.Debug("encoding tasks")
		jsonFileWriter, _ := multipartWriter.CreateFormFile("jsonfile", "data.json")
		if err := json.NewEncoder(jsonFileWriter).Encode(tasks); err != nil {
			logger.Warn(err)
		}

		multipartWriter.Close()
		pwriter.Close()
	}()

	logger.Debug("submitting POST to %v/jobs: %v", this.masterUrl, r)
	client := http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		logger.Warn(err)
	}

	respcode := 0
	if resp != nil {
		respcode = resp.StatusCode
		resp.Body.Close()
	}

	logger.Debug("completed POST to %v/jobs: %d", this.masterUrl, respcode)
	return
}

func (this *Scribe) ArchiveJob(jd JobDetails) (err os.Error) {
	logger.Debug("ArchiveJob(%v)", jd)
	if jd.State != COMPLETE {
		logger.Debug("ArchiveJob(%v): NOT COMPLETE", jd)
		return
	}

	jobUri := fmt.Sprintf("%v/jobs/%v/archive", this.masterUrl, jd.JobId)
	r, err := http.NewRequest("POST", jobUri, strings.NewReader(""))
	if err != nil {
		logger.Warn(err)
		return
	}

	r.Header.Set("x-golem-apikey", this.apikey)

	logger.Debug("submitting POST to %v: %v", jobUri, r)
	client := http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		logger.Warn(err)
	}

	respcode := 0
	if resp != nil {
		respcode = resp.StatusCode
		resp.Body.Close()
	}

	logger.Debug("completed POST to %v: %d", jobUri, respcode)
	return
}
