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
	"crypto/rand"
)


////////////////////////////////////////////////////
/// Submission
/// TODO: Kill submissions once they finish.

type Submission struct {
	Uri              string
	SubId            string
	CoutFileChan     chan string
	CerrFileChan     chan string
	Jobs             []RequestedJob
	ErrorChan        chan *Job
	FinishedChan     chan *Job
	TotalJobsChan    chan int
	FinishedJobsChan chan int
	ErroredJobsChan  chan int
	stopChan         chan int
}


func NewSubmission(js *[]RequestedJob, jobChan chan *Job) *Submission {
	rJobs := *js
	//subId := <-subidChan
	//subidChan <- subId + 1
	subId := make([]byte, 16)
	_, err := rand.Read(subId)
	if err != nil {
		log("Error generating rand subId: %v", err)
	}
	subIds := fmt.Sprintf("%x", subId)
	s := Submission{
		Uri:              fmt.Sprintf("/jobs/%v", subIds),
		SubId:            subIds,
		CoutFileChan:     make(chan string, iobuffersize),
		CerrFileChan:     make(chan string, iobuffersize),
		Jobs:             rJobs,
		ErrorChan:        make(chan *Job, 1),
		FinishedChan:     make(chan *Job, 1),
		FinishedJobsChan: make(chan int, 1),
		ErroredJobsChan:  make(chan int, 1),
		TotalJobsChan:    make(chan int, 1),
		stopChan:         make(chan int, 1)}
	totalJobs := 0
	for _, vals := range s.Jobs {
		totalJobs += vals.Count
	}
	s.TotalJobsChan <- totalJobs
	s.FinishedJobsChan <- 0
	s.ErroredJobsChan <- 0
	go s.monitorJobs()
	go s.writeIo()
	go s.submitJobs(jobChan)

	return &s

}

func (s *Submission) DescribeSelfJson() string {
	TotalJobs := <-s.TotalJobsChan
	FinishedJobs := <-s.FinishedJobsChan
	ErroredJobs := <-s.ErroredJobsChan
	log("Describing SubId: %v, %v finished, %v errored, %v total", s.SubId, FinishedJobs, ErroredJobs, TotalJobs)
	rv := fmt.Sprintf("{\"uri\":\"%v\",\"SubId\":%v, \"TotalJobs\":%v,\"FinishedJobs\":%v,\"ErroredJobs\":%v}", s.Uri, s.SubId, TotalJobs, FinishedJobs, ErroredJobs)
	s.TotalJobsChan <- TotalJobs
	s.FinishedJobsChan <- FinishedJobs
	s.ErroredJobsChan <- ErroredJobs
	return rv
}

func (s *Submission) Stop() {

	s.stopChan <- 1
	s.stopChan <- 1 //second time to make sure submitJobs takes the 1 out of the chan
	log("Stoped submission for SubId: %v", s.SubId)
}

func (s Submission) monitorJobs() {
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
		FinishedJobs := <-s.FinishedJobsChan
		ErroredJobs := <-s.ErroredJobsChan

		log("Job update SubId: %v, %v finished, %v errored, %v total", s.SubId, FinishedJobs, ErroredJobs, TotalJobs)
		if TotalJobs == (FinishedJobs + ErroredJobs) {
			log("All Jobs done for SubId: %v, %v finished, %v errored", s.SubId, FinishedJobs, ErroredJobs)
			//s.killChan <- 1 //TODO: clean up submission object here
			s.TotalJobsChan <- TotalJobs
			s.FinishedJobsChan <- FinishedJobs
			s.ErroredJobsChan <- ErroredJobs
			return
		}
		s.TotalJobsChan <- TotalJobs
		s.FinishedJobsChan <- FinishedJobs
		s.ErroredJobsChan <- ErroredJobs
	}

}

func (s Submission) submitJobs(jobChan chan *Job) {
	jobId := 0
	for lineId, vals := range s.Jobs {
		for i := 0; i < vals.Count; i++ {
			select {
			case jobChan <- &Job{SubId: s.SubId, LineId: lineId, JobId: jobId, Args: vals.Args}:
				jobId++
			case <-s.stopChan:
				return //TODO: add indication that we stopped
			}
		}
	}
	log("All jobs submitted for SubId: %v", s.SubId)
}

func (s Submission) writeIo() {
	outf, err := os.Create(fmt.Sprintf("%v.out.txt", s.SubId))
	if err != nil {
		log("Error creating output file: %v\n", err)
	}
	defer outf.Close()
	errf, err := os.Create(fmt.Sprintf("%v.err.txt", s.SubId))
	if err != nil {
		log("Error creating output file: %v\n", err)
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
