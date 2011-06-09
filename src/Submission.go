/*
   Copyright (C) 2003-2010 Institute for Systems Biology
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
	"json"
	"fmt"
	"os"
)


////////////////////////////////////////////////////
/// Submission
/// TODO: Kill submissions once they finish.

type Submission struct {
	SubId        int
	CoutFileChan chan string
	CerrFileChan chan string
	Jobs         []RequestedJob
	ErrorChan    chan *Job
	FinishedChan chan *Job
	TotalJobs    int
	FinishedJobs int
	ErroredJobs  int
	killChan     chan int
}


func NewSubmission(reqjson string) *Submission {
	rJobs := make([]RequestedJob, 0, 100)
	if err := json.Unmarshal([]byte(reqjson), &rJobs); err != nil {
		log("%v", err)
	}
	subId := <-subidChan
	subidChan <- subId + 1
	s := Submission{SubId: subId,
		CoutFileChan: make(chan string, 10),
		CerrFileChan: make(chan string, 10),
		Jobs:         rJobs,
		ErrorChan:    make(chan *Job, 1),
		FinishedChan: make(chan *Job, 1),
		FinishedJobs: 0,
		ErroredJobs:  0,
		TotalJobs:    0,
		killChan:     make(chan int, 0)}
	for _, vals := range s.Jobs {
		s.TotalJobs += vals.Count
	}
	go s.monitorJobs()
	go s.writeIo()
	go s.submitJobs()

	return &s

}

func (s Submission) monitorJobs() {
	for {
		select {
		case <-s.ErrorChan:
			s.ErroredJobs++
		case <-s.FinishedChan:
			s.FinishedJobs++
		}
		log("Job update SubId: %v, %v finished, %v errored, %v total", s.SubId, s.FinishedJobs, s.ErroredJobs, s.TotalJobs)
		if s.TotalJobs == (s.FinishedJobs + s.ErroredJobs) {
			log("All Jobs done for SubId: %v, %v finished, %v errored", s.SubId, s.FinishedJobs, s.ErroredJobs)
			//s.killChan <- 1 //TODO: clean up submission object here
			return
		}
	}

}

func (s Submission) submitJobs() {
	jobId := 0
	for lineId, vals := range s.Jobs {
		for i := 0; i < vals.Count; i++ {
			jobChan <- &Job{SubId: s.SubId, LineId: lineId, JobId: jobId, Args: vals.Args}
			jobId++
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
