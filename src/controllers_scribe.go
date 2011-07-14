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
	"strings"
	"os"
)

type ScribeJobController struct {
	scribe *Scribe
}

func (c ScribeJobController) RetrieveAll() (json string, numberOfItems int, err os.Error) {
	log("RetrieveAll")
	items, err := c.scribe.store.All()
	if err != nil {
		return
	}

	jsonArray := make([]string, 0)
	for _, s := range items {
		val, _ := s.MarshalJSON()
		jsonArray = append(jsonArray, string(val))
	}
	numberOfItems = len(jsonArray)
	json = strings.Join(jsonArray, ",")
	err = nil
	return
}
func (c ScribeJobController) Retrieve(jobId string) (json string, err os.Error) {
	log("Retrieve:%v", jobId)

	item, err := c.scribe.store.Get(jobId)
	if err != nil {
	    return
	}

	val, err := item.MarshalJSON()
	if err != nil {
	    return
	}

	json = string(val)
	return
}
func (c ScribeJobController) NewJob(r *http.Request) (jobId string, err os.Error) {
	reqjson := r.FormValue("data")
	log("NewJob:%v", reqjson)
	jobId = UniqueId()
	err = nil
	return
}
func (c ScribeJobController) Stop(jobId string) os.Error {
	log("Stop:%v", jobId)
	return os.NewError("unable to stop")
}
func (c ScribeJobController) Kill(jobId string) os.Error {
	log("Kill:%v", jobId)
	return os.NewError("unable to kill")
}

type ScribeNodeController struct {
	scribe *Scribe
}

func (c ScribeNodeController) RetrieveAll() (json string, numberOfItems int, err os.Error) {
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
func (c ScribeNodeController) RestartAll() os.Error {
	log("Restart:")
	return os.NewError("unable to restart")
}
func (c ScribeNodeController) Resize(nodeId string, numberOfThreads int) os.Error {
	log("Resize:%v,%i", nodeId, numberOfThreads)
	return os.NewError("unable to resize")
}
func (c ScribeNodeController) KillAll() os.Error {
	log("Kill")
	return os.NewError("unable to kill")
}
