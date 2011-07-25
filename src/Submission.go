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
	"os"
	"time"
)

// TODO: Kill submissions once they finish.
type Submission struct {
	Identity          Identity
	Description    Description
	Status            Status
	Progress        Progress
	Tasks            []Task

	CoutFileChan     chan string
	CerrFileChan     chan string
	ErrorChan        chan *Job
	FinishedChan     chan *Job
	stopChan         chan int
}

func NewSubmission(jd JobDetails, jobChan chan *Job) *Submission {
	s := Submission{
	    Identity: jd.Identity,
	    Description: jd.Description,
	    Status: jd.Status,
	    Progress: jd.Progress,
		Tasks:      jd.Tasks,
		CoutFileChan:     make(chan string, iobuffersize),
		CerrFileChan:     make(chan string, iobuffersize),
		ErrorChan:        make(chan *Job, 1),
		FinishedChan:     make(chan *Job, 1),
		stopChan:         make(chan int, 0) }

	go s.monitorJobs()
	go s.writeIo()
	go s.submitJobs(jobChan)

	return &s
}

func (s *Submission) Stop() bool {
	if s.Status.Running {
		select {
		case s.stopChan <- 1:
			log("stopped: %v", s.Identity)
			s.Status.Running = false
		case <-time.After(250000000):
			log("timeout stopping: %v", s.Identity)
		}
	}
	return s.Status.Running
}

func (s Submission) monitorJobs() {
	vlog("monitorJobs")

	for {
		select {
		case <-s.ErrorChan:
			s.Progress.Errored = 1 + s.Progress.Errored
		case <-s.FinishedChan:
            s.Progress.Finished = 1 + s.Progress.Finished
		}

		vlog("updating job: %v, %v, %v", s.Identity, s.Progress, s.Status)
		if s.Progress.isComplete() {
			s.Status.Running = false
			log("job completed: %v, %v, %v", s.Identity, s.Progress, s.Status)
			return
		}
	}
}

func (s Submission) submitJobs(jobChan chan *Job) {
	vlog("submitJobs")

	taskId := 0
	for lineId, vals := range s.Tasks {
		vlog("submitJobs:[%d,%v]", lineId, vals)
		for i := 0; i < vals.Count; i++ {
			select {
			case jobChan <- &Job{SubId: s.Identity.JobId, LineId: lineId, JobId: taskId, Args: vals.Args}:
				taskId++
			case <-s.stopChan:
				return //TODO: add indication that we stopped
			}
		}
	}
	log("[%d] tasks submitted for [%v]", taskId, s.Identity)
}

func (s Submission) writeIo() {
	vlog("Submission.writeIo(%v)", s.Identity)

	outf, err := os.Create(fmt.Sprintf("%v.out.txt", s.Identity.JobId))
	if err != nil {
		log("writeIo: %v", err)
	}
	defer outf.Close()

	errf, err := os.Create(fmt.Sprintf("%v.err.txt", s.Identity.JobId))
	if err != nil {
		log("writeIo: %v", err)
	}
	defer errf.Close()

	for {
		select {
		case msg := <-s.CoutFileChan:
			fmt.Fprint(outf, msg)
		case errmsg := <-s.CerrFileChan:
			fmt.Fprint(errf, errmsg)
		}
	}
}
