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
}


func NewSubmission(reqjson string) *Submission {
	rJobs := make([]RequestedJob, 0, 100)
	if err := json.Unmarshal([]byte(reqjson), &rJobs); err != nil {
		log("%v", err)
	}
	subId := <-subidChan
	subidChan <- subId + 1
	s := Submission{SubId: subId, CoutFileChan: make(chan string, 10), CerrFileChan: make(chan string, 10), Jobs: rJobs}
	go s.writeIo()
	jobId := 0
	for lineId, vals := range rJobs {
		for i := 0; i < vals.Count; i++ {
			jobChan <- &Job{SubId: subId, LineId: lineId, JobId: jobId, Args: vals.Args}
			jobId++
		}
	}
	return &s

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
