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
	"fmt"
	"os"
)


/////////////////////////////////////////////////
//scribe
type Scribe struct {
	jobController  JobController
	nodeController NodeController
}

func NewScribe() *Scribe {
	s := Scribe{jobController: ScribeJobController{}, nodeController: ScribeNodeController{}}
	return &s
}

// Job and Node Controllers
type ScribeJobController struct {

}

func (mc ScribeJobController) RetrieveAll(r *http.Request) (json string, numberOfItems int, err os.Error) {
	log("RetrieveAll")
	json = "{ items:[], numberOfItems: 0, uri:'/jobs' }"
	numberOfItems = 0
	err = nil
	return
}
func (mc ScribeJobController) Retrieve(jobId string) (json string, err os.Error) {
	log("Retrieve:%v", jobId)
	json = fmt.Sprintf("{ items:[], numberOfItems: 0, uri:'/jobs/%v' }", jobId)
	err = nil
	return
}
func (mc ScribeJobController) NewJob(r *http.Request) (jobId string, err os.Error) {
	reqjson := r.FormValue("data")
	log("NewJob:%v", reqjson)
	jobId = UniqueId()
	err = nil
	return
}
func (mc ScribeJobController) Stop(jobId string) os.Error {
	log("Stop:%v", jobId)
	return os.NewError("unable to stop")
}
func (mc ScribeJobController) Kill(jobId string) os.Error {
	log("Kill:%v", jobId)
	return os.NewError("unable to kill")
}

// Do Nothing Node Controller implementation
type ScribeNodeController struct {

}

func (c ScribeNodeController) RetrieveAll(r *http.Request) (json string, numberOfItems int, err os.Error) {
	log("RetrieveAll")
	json = "{ items:[], numberOfItems: 0, uri:'/nodes' }"
	numberOfItems = 0
	err = nil
	return
}
func (c ScribeNodeController) Retrieve(nodeId string) (json string, err os.Error) {
	log("Retrieve:%v", nodeId)
	json = fmt.Sprintf("{ items:[], numberOfItems: 0, uri:'/nodes/%v' }", nodeId)
	err = nil
	return
}
func (c ScribeNodeController) Restart(nodeId string) os.Error {
	log("Restart:%v", nodeId)
	return os.NewError("unable to restart")
}
func (c ScribeNodeController) Resize(nodeId string, numberOfThreads int) os.Error {
	log("Resize:%v,%i", nodeId, numberOfThreads)
	return os.NewError("unable to resize")
}
func (c ScribeNodeController) Kill(nodeId string) os.Error {
	log("Kill:%v", nodeId)
	return os.NewError("unable to kill")
}
