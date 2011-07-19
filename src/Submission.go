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
	Uri              string
	SubId            string
	CoutFileChan     chan string
	CerrFileChan     chan string
	Tasks            []Task
	ErrorChan        chan *Job
	FinishedChan     chan *Job
	TotalJobsChan    chan int
	FinishedJobsChan chan int
	ErroredJobsChan  chan int
	stopChan         chan int
	runningChan      chan bool
	SubLocalTime     string
}


func NewSubmission(js []Task, jobChan chan *Job) *Submission {
    iobuffersize, err := ConfigFile.GetInt("master", "buffersize")
     if err != nil {
        vlog("defaulting buffer to 1000:%v", err)
        iobuffersize = 1000
     }

	subId := UniqueId()
	localTime := time.SecondsToLocalTime(time.Seconds())
	formattedTime := localTime.Format(time.ANSIC)
	s := Submission{
		Uri:              fmt.Sprintf("/jobs/%v", subId),
		SubId:            subId,
		CoutFileChan:     make(chan string, iobuffersize),
		CerrFileChan:     make(chan string, iobuffersize),
		Tasks:            js,
		ErrorChan:        make(chan *Job, 1),
		FinishedChan:     make(chan *Job, 1),
		FinishedJobsChan: make(chan int, 1),
		ErroredJobsChan:  make(chan int, 1),
		TotalJobsChan:    make(chan int, 1),
		stopChan:         make(chan int, 0),
		runningChan:      make(chan bool, 1),
		SubLocalTime:     formattedTime}
	totalJobs := 0
	for _, vals := range s.Tasks {
		totalJobs += vals.Count
	}
	s.TotalJobsChan <- totalJobs
	s.FinishedJobsChan <- 0
	s.ErroredJobsChan <- 0
	s.runningChan <- true
	go s.monitorJobs()
	go s.writeIo()
	go s.submitJobs(jobChan)

	return &s
}

func (s *Submission) MarshalJSON() ([]byte, os.Error) {
	vlog("sniffing total jobs")
	TotalJobs := <-s.TotalJobsChan
	s.TotalJobsChan <- TotalJobs
	vlog("sniffing finished jobs")
	FinishedJobs := <-s.FinishedJobsChan
	s.FinishedJobsChan <- FinishedJobs
	vlog("sniffing errored jobs")
	ErroredJobs := <-s.ErroredJobsChan
	s.ErroredJobsChan <- ErroredJobs
	vlog("sniffing running jobs")
	running := <-s.runningChan
	s.runningChan <- running

	log("Describing SubId: %v, %v finished, %v errored, %v total, %v localtime", s.SubId, FinishedJobs, ErroredJobs, TotalJobs, s.SubLocalTime)
	rv := fmt.Sprintf("{ uri:\"%v\", SubId:\"%v\", TotalJobs:%v, FinishedJobs:%v, ErroredJobs:%v, Running:%v, CreatedAt:\"%v\" }", s.Uri, s.SubId, TotalJobs, FinishedJobs, ErroredJobs, running, s.SubLocalTime)

	vlog("Returning description")
	return []byte(rv), nil
}

func (s *Submission) Stop() bool {
	rv := false
	running := <-s.runningChan
	if running {
		select {
		case s.stopChan <- 1:
			rv = true
			log("Stoped submission for SubId: %v", s.SubId)
			running = false
		case <-time.After(250000000):
			rv = false
			log("Time out stoping SubId: %v", s.SubId)
		}
	}
	s.runningChan <- running
	return rv
}

func (s Submission) monitorJobs() {
	vlog("monitorJobs")

	for {
		select {
		case <-s.ErrorChan:
			ErroredJobs := <-s.ErroredJobsChan
			ErroredJobs++
			s.ErroredJobsChan <- ErroredJobs
		case <-s.FinishedChan:
			FinishedJobs := <-s.FinishedJobsChan
			FinishedJobs++
			s.FinishedJobsChan <- FinishedJobs
		}

		TotalJobs := <-s.TotalJobsChan
		s.TotalJobsChan <- TotalJobs

		FinishedJobs := <-s.FinishedJobsChan
		s.FinishedJobsChan <- FinishedJobs

		ErroredJobs := <-s.ErroredJobsChan
		s.ErroredJobsChan <- ErroredJobs

		running := <-s.runningChan

		vlog("Job update SubId: %v, %v finished, %v errored, %v total", s.SubId, FinishedJobs, ErroredJobs, TotalJobs)
		if TotalJobs == (FinishedJobs + ErroredJobs) {
			log("All Jobs done for SubId: %v, %v finished, %v errored", s.SubId, FinishedJobs, ErroredJobs)
			//TODO: clean up submission object here
			s.runningChan <- false
			return
		}
		s.runningChan <- running
	}
}

func (s Submission) submitJobs(jobChan chan *Job) {
	vlog("submitJobs")

	taskId := 0
	for lineId, vals := range s.Tasks {
		vlog("submitJobs:[%d,%v]", lineId, vals)
		for i := 0; i < vals.Count; i++ {
			select {
			case jobChan <- &Job{SubId: s.SubId, LineId: lineId, JobId: taskId, Args: vals.Args}:
				taskId++
			case <-s.stopChan:
				return //TODO: add indication that we stopped
			}
		}
	}
	log("[%d] tasks submitted for [%v]", taskId, s.SubId)
}

func (s Submission) writeIo() {
	vlog("writeIo")

	outf, err := os.Create(fmt.Sprintf("%v.out.txt", s.SubId))
	if err != nil {
		log("writeIo: %v", err)
	}
	defer outf.Close()

	errf, err := os.Create(fmt.Sprintf("%v.err.txt", s.SubId))
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
