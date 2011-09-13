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
	logger.Debug("NewSubmission(%v)", jd)
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

func (this *Submission) Stop() bool {
	logger.Debug("Stop()")
	dtls := this.SniffDetails()
	logger.Debug("Stop(): %v", dtls.JobId)

	if dtls.State == RUNNING {
		select {
		case this.stopChan <- 1:
			this.SetState(COMPLETE, STOPPED)
			logger.Debug("Stop():%v", this.SniffDetails())
		case <-time.After(250000000):
			logger.Printf("Stop(): timeout stopping: %v", dtls.JobId)
		}
	}

	return this.SniffDetails().IsRunning()
}

func (this *Submission) SniffDetails() JobDetails {
	dtls := <-this.Details
	this.Details <- dtls
	logger.Debug("SniffDetails(): dtls=%v", dtls)
	return dtls
}

func (this *Submission) MonitorWorkTasks() {
	logger.Debug("MonitorWorkTasks()")
	for {
		select {
		case <-this.ErrorChan:
			dtls := <-this.Details
			dtls.Progress.Errored = 1 + dtls.Progress.Errored
			dtls.LastModified = time.LocalTime().String()
			this.Details <- dtls

			logger.Debug("ERROR [%v,%v]", dtls.JobId, dtls.Progress.Errored)

		case <-this.FinishedChan:
			dtls := <-this.Details
			dtls.Progress.Finished = 1 + dtls.Progress.Finished
			dtls.LastModified = time.LocalTime().String()
			this.Details <- dtls

			logger.Debug("FINISHED [%v,%v]", dtls.JobId, dtls.Progress.Finished)
		}

		dtls := this.SniffDetails()
		if dtls.Progress.isComplete() {
			logger.Debug("COMPLETED [%v]", dtls)
			this.SetState(COMPLETE, SUCCESS)
			this.stopChan <- 1
			this.stopChan <- 1
			this.stopChan <- 1
			logger.Debug("COMPLETED [%v]: DONE", dtls.JobId)
			return
		}
	}
}

func (this *Submission) SubmitJobs(jobChan chan *WorkerJob) {
	logger.Debug("SubmitJobs()")

	this.SetState(RUNNING, READY)

	dtls := this.SniffDetails()
	taskId := 0
	for lineId, vals := range this.Tasks {
		logger.Debug("[%d,%v]", lineId, vals)
		for i := 0; i < vals.Count; i++ {
			select {
			case jobChan <- &WorkerJob{SubId: dtls.JobId, LineId: lineId, JobId: taskId, Args: vals.Args}:
				taskId++
			case <-this.stopChan:
				return //TODO: add indication that we stopped
			}
		}
	}
	logger.Printf("tasks submitted [%d, %v]", taskId, dtls.JobId)
}

func (this *Submission) WriteIo() {
	dtls := this.SniffDetails()
	logger.Debug("WriteIo(%v)", dtls.JobId)

	var stdOutFile io.WriteCloser = nil
	var stdErrFile io.WriteCloser = nil
	var err os.Error

	for {
		select {
		case msg := <-this.CoutFileChan:
			if stdOutFile == nil {
				if stdOutFile, err = os.Create(fmt.Sprintf("%v.out.txt", dtls.JobId)); err != nil {
					logger.Warn(err)
				}
				if stdOutFile != nil {
					defer stdOutFile.Close()
				}
			}

			fmt.Fprint(stdOutFile, msg)
		case errmsg := <-this.CerrFileChan:
			if stdErrFile == nil {
				if stdErrFile, err = os.Create(fmt.Sprintf("%v.err.txt", dtls.JobId)); err != nil {
					logger.Warn(err)
				}
				if stdErrFile != nil {
					defer stdErrFile.Close()
				}
			}

			fmt.Fprint(stdErrFile, errmsg)
		case <-time.After(1 * second):
			logger.Debug("checking for done: %v", dtls.JobId)
			select {
			case <-this.stopChan:
				logger.Debug("stop chan: %v", dtls.JobId)
				return
			default:
			}

		}
	}
}

func (this *Submission) SetState(state string, status string) {
	logger.Debug("SetState(%v,%v):before=%v", state, status, this.SniffDetails())
	x := <-this.Details
	x.State = state
	x.Status = status
	x.LastModified = time.LocalTime().String()
	this.Details <- x
	logger.Debug("SetState(%v,%v):after=%v", state, status, this.SniffDetails())
}

func (this *Submission) UpdateProgress() {
	logger.Debug("UpdateProgress():before=%v", this.SniffDetails())
	x := <-this.Details
	x.Progress.Errored = 1 + x.Progress.Errored
	x.LastModified = time.LocalTime().String()
	this.Details <- x
	logger.Debug("UpdateProgress():after=%v", this.SniffDetails())
}
