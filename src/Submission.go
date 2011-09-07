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
	"io"
	"time"
)

// TODO: Kill submissions once they finish.
type Submission struct {
	Details chan JobDetails
	Tasks   []Task

	CoutFileChan chan string
	CerrFileChan chan string
	ErrorChan    chan *WorkerJob
	FinishedChan chan *WorkerJob
	stopChan     chan int
}

func NewSubmission(jd JobDetails, tasks []Task, jobChan chan *WorkerJob) *Submission {
	vlog("NewSubmission(%v)", jd)
	s := Submission{
		Details:      make(chan JobDetails, 1),
		Tasks:        tasks,
		CoutFileChan: make(chan string, iobuffersize),
		CerrFileChan: make(chan string, iobuffersize),
		ErrorChan:    make(chan *WorkerJob, 1),
		FinishedChan: make(chan *WorkerJob, 1),
		stopChan:     make(chan int, 3)}

	s.Details <- jd

	go s.MonitorWorkTasks()
	go s.WriteIo()
	go s.SubmitJobs(jobChan)

	return &s
}

func (s *Submission) Stop() bool {
	vlog("Submission.Stop()")
	dtls := s.SniffDetails()
	vlog("Submission.Stop(): %v", dtls.JobId)

	if dtls.Running {
		select {
		case s.stopChan <- 1:
			dtls := <-s.Details
			dtls.Running = false
			dtls.LastModified = time.LocalTime().String()
			s.Details <- dtls

			log("Submission.Stop(): %v", dtls.JobId)
		case <-time.After(250000000):
			log("Submission.Stop(): timeout stopping: %v", dtls.JobId)
		}
	}

	return s.SniffDetails().Running
}

func (this *Submission) SniffDetails() JobDetails {
	vlog("Submission.SniffDetails()")
	dtls := <-this.Details
	this.Details <- dtls
	return dtls
}

func (s Submission) MonitorWorkTasks() {
	vlog("Submission.MonitorWorkTasks()")
	for {
		select {
		case <-s.ErrorChan:
			dtls := <-s.Details
			dtls.Progress.Errored = 1 + dtls.Progress.Errored
			dtls.LastModified = time.LocalTime().String()
			s.Details <- dtls

			vlog("Submission.MonitorWorkTasks() [ERROR,%v,%v]", dtls.JobId, dtls.Progress.Errored)

		case <-s.FinishedChan:
			dtls := <-s.Details
			dtls.Progress.Finished = 1 + dtls.Progress.Finished
			dtls.LastModified = time.LocalTime().String()
			s.Details <- dtls

			vlog("Submission.MonitorWorkTasks() [FINISHED,%v,%v]", dtls.JobId, dtls.Progress.Finished)
		}

		dtls := s.SniffDetails()
		if dtls.Progress.isComplete() {
			vlog("Submission.MonitorWorkTasks() [COMPLETED,%v]", dtls.JobId)
			dtls = <-s.Details
			dtls.Running = false
			dtls.LastModified = time.LocalTime().String()
			s.Details <- dtls
			//send three messages to stop chan to stop all things plus one
			s.stopChan <- 1
			s.stopChan <- 1
			s.stopChan <- 1
			vlog("Submission.MonitorWorkTasks() [COMPLETED,%v]: DONE", dtls.JobId)
			return
		}
	}
}

func (s Submission) SubmitJobs(jobChan chan *WorkerJob) {
	vlog("Submission.SubmitJobs()")

	dtls := <-s.Details
	dtls.Scheduled = true
	dtls.Running = true
	dtls.LastModified = time.LocalTime().String()
	s.Details <- dtls

	taskId := 0
	for lineId, vals := range s.Tasks {
		vlog("Submission.SubmitJobs():[%d,%v]", lineId, vals)
		for i := 0; i < vals.Count; i++ {
			select {
			case jobChan <- &WorkerJob{SubId: dtls.JobId, LineId: lineId, JobId: taskId, Args: vals.Args}:
				taskId++
			case <-s.stopChan:
				return //TODO: add indication that we stopped
			}
		}
	}
	log("Submission.SubmitJobs(): [%d] tasks submitted for [%v]", taskId, dtls.JobId)
}

func (s Submission) WriteIo() {
	dtls := s.SniffDetails()
	vlog("Submission.WriteIo(%v)", dtls.JobId)

	var stdOutFile io.WriteCloser = nil
	var stdErrFile io.WriteCloser = nil
	var err os.Error

	for {
		select {
		case msg := <-s.CoutFileChan:
			if stdOutFile == nil {
				if stdOutFile, err = os.Create(fmt.Sprintf("%v.out.txt", dtls.JobId)); err != nil {
					warn("Submission.WriteIo(%v.out.txt): %v", dtls.JobId, err)
				}
				if stdOutFile != nil {
					defer stdOutFile.Close()
				}
			}

			fmt.Fprint(stdOutFile, msg)
		case errmsg := <-s.CerrFileChan:
			if stdErrFile == nil {
				if stdErrFile, err = os.Create(fmt.Sprintf("%v.err.txt", dtls.JobId)); err != nil {
					warn("Submission.WriteIo(%v.err.txt): %v", dtls.JobId, err)
				}
				if stdErrFile != nil {
					defer stdErrFile.Close()
				}
			}

			fmt.Fprint(stdErrFile, errmsg)
		case <-time.After(1 * second):
			vlog("Submission.WriteIO(%v): checking for done", dtls.JobId)
			select {
			case <-s.stopChan:
				vlog("Submission.WriteIO(%v): stopChan", dtls.JobId)
				return
			default:
			}

		}
	}
}
