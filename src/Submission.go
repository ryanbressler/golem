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
	Details chan JobDetails
	Tasks   []Task

	CoutFileChan chan string
	CerrFileChan chan string
	ErrorChan    chan *Job
	FinishedChan chan *Job
	stopChan     chan int
}

func NewSubmission(jd JobDetails, tasks []Task, jobChan chan *Job) *Submission {
	s := Submission{
		Details:      make(chan JobDetails, 1),
		Tasks:        tasks,
		CoutFileChan: make(chan string, iobuffersize),
		CerrFileChan: make(chan string, iobuffersize),
		ErrorChan:    make(chan *Job, 1),
		FinishedChan: make(chan *Job, 1),
		stopChan:     make(chan int, 0)}

	s.Details <- jd

	go s.MonitorWorkTasks()
	go s.writeIo()
	go s.submitJobs(jobChan)

	return &s
}

func (s *Submission) Stop() bool {
	dtls := s.SniffDetails()

	if dtls.Running {
		select {
		case s.stopChan <- 1:
			dtls := <-s.Details
			dtls.Running = false
			s.Details <- dtls

			log("Submission.Stop(): %v", dtls.JobId)
		case <-time.After(250000000):
			log("timeout stopping: %v", dtls.JobId)
		}
	}

	return s.SniffDetails().Running
}

func (this *Submission) SniffDetails() JobDetails {
	dtls := <-this.Details
	this.Details <- dtls
	return dtls
}

func (s Submission) MonitorWorkTasks() {
	for {
		select {
		case <-s.ErrorChan:
			Det := <-s.Details
			Det.Progress.Errored++
			s.Details <- Det

			vlog("MonitorWorkTasks [ERROR]", Det.JobId)

		case <-s.FinishedChan:
			Det := <-s.Details
			Det.Progress.Finished++
			s.Details <- Det

			vlog("MonitorWorkTasks [FINISHED]", Det.JobId)
		}

		dtls := s.SniffDetails()
		if dtls.Progress.isComplete() {
			vlog("MonitorWorkTasks [COMPLETED]: %v", dtls.JobId)
			dtls = <-s.Details
			dtls.Running = false
			s.Details <- dtls
			return
		}

	}
}

func (s Submission) submitJobs(jobChan chan *Job) {
	vlog("submitJobs")

	dtls := s.SniffDetails()
	taskId := 0
	for lineId, vals := range s.Tasks {
		vlog("submitJobs:[%d,%v]", lineId, vals)
		for i := 0; i < vals.Count; i++ {
			select {
			case jobChan <- &Job{SubId: dtls.JobId, LineId: lineId, JobId: taskId, Args: vals.Args}:
				taskId++
			case <-s.stopChan:
				return //TODO: add indication that we stopped
			}
		}
	}
	log("[%d] tasks submitted for [%v]", taskId, dtls.JobId)
}

func (s Submission) writeIo() {
	dtls := s.SniffDetails()
	vlog("Submission.writeIo(%v)", dtls.JobId)

	outf, err := os.Create(fmt.Sprintf("%v.out.txt", dtls.JobId))
	if err != nil {
		log("writeIo: %v", err)
	}
	defer outf.Close()

	errf, err := os.Create(fmt.Sprintf("%v.err.txt", dtls.JobId))
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
