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
	"json"
	"fmt"
	"os"
	"strconv"
)

// REST interface for Job and Node Controllers
type RestJsonAPI struct {
	jobController  JobController
	nodeController NodeController
	hashedpw       string
}

type Retriever interface {
	RetrieveAll() (items []interface{}, err os.Error)
	Retrieve(itemId string) (item interface{}, err os.Error)
}
type JobController interface {
	Retriever
	NewJob(r *http.Request) (jobId string, err os.Error)
	Stop(jobId string) (err os.Error)
	Kill(jobId string) (err os.Error)
}

type NodeController interface {
	Retriever
	RestartAll() os.Error
	KillAll() os.Error
	Resize(nodeId string, numberOfThreads int) os.Error
}

func HandleRestJson(jc JobController, nc NodeController) {
	hostname, err := ConfigFile.GetString("default", "hostname")
	if err != nil {
		panic(err)
	}

	hpw := ""
	password, err := ConfigFile.GetString("default", "password")
	if err != nil {
		log("no password specified")
	} else {
		hpw = hashPw(password)
	}

	api := RestJsonAPI{jobController: jc, nodeController: nc, hashedpw: hpw}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { api.rootHandler(w, r) })
	http.Handle("/html/", http.FileServer("html", "/html"))
	http.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) { api.jobHandler(w, r) })
	http.HandleFunc("/nodes/", func(w http.ResponseWriter, r *http.Request) { api.nodeHandler(w, r) })

	log("running at %v", hostname)

	if err := ListenAndServeTLSorNot(hostname, nil); err != nil {
		panic(err)
	}
	return
}

// web handlers
func (this *RestJsonAPI) rootHandler(w http.ResponseWriter, r *http.Request) {
	log("%v /", r.Method)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "{ jobs: '/jobs', nodes: '/nodes' }")
}

func (this *RestJsonAPI) jobHandler(w http.ResponseWriter, r *http.Request) {
	log("%v /jobs", r.Method)

	w.Header().Set("Content-Type", "application/json")

	// TODO : Add logic to retrieve outputs from job
	// TODO : Manage errors

	switch r.Method {
	case "GET":
		log("Method = GET.")
		jobId, verb := parseJobUri(r.URL.Path)
		switch {
		case jobId != "":
			this.retrieve("/jobs", jobId, this.jobController, w)
		case jobId == "" && verb == "":
			this.retrieveAll("/jobs", this.jobController, w)
		default:
			w.WriteHeader(http.StatusNotFound)
		}

	case "POST":
		log("Method = POST.")
		if this.checkPassword(r) == false {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		jobId, verb := parseJobUri(r.URL.Path)
		switch {
		case jobId != "" && verb == "stop":
			err := this.jobController.Stop(jobId)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
		case jobId == "" && verb == "":
			jobId, err := this.jobController.NewJob(r)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			fmt.Fprintf(w, "{ uri: '/jobs/%v' id:'%v' }", jobId, jobId)
		default:
			w.WriteHeader(http.StatusNotFound)
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (this *RestJsonAPI) nodeHandler(w http.ResponseWriter, r *http.Request) {
	log("nodeHandler")

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		pathParts := splitRestUrl(r.URL.Path)
		nparts := len(pathParts)
		switch {
		case nparts == 2:
			this.retrieve("/nodes", pathParts[1], this.nodeController, w)
		default:
			this.retrieveAll("/nodes", this.nodeController, w)
		}
	case "POST":
		if this.checkPassword(r) == false {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		err := this.postNodeHandler(r)
		if err != nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(200)
		}
	}
}

func (this *RestJsonAPI) postNodeHandler(r *http.Request) os.Error {
	spliturl := splitRestUrl(r.URL.Path)
	nsplit := len(spliturl)
	switch {
	case nsplit == 2 && spliturl[1] == "restart":
		return this.nodeController.RestartAll()

	case nsplit == 2 && spliturl[1] == "die":
		return this.nodeController.KillAll()

	case nsplit == 4 && spliturl[2] == "resize":
		nodeId := spliturl[1]
		numberOfThreads, err := strconv.Atoi(spliturl[3])
		if err != nil {
			return err
		}
		return this.nodeController.Resize(nodeId, numberOfThreads)
	}
	return nil
}

func (this *RestJsonAPI) checkPassword(r *http.Request) bool {
	if this.hashedpw != "" {
		pw := hashPw(r.Header.Get("Password"))
		log("Verifying password.")
		return this.hashedpw == pw
	}
	return true
}

// TODO : Deal with URI
func (this *RestJsonAPI) retrieve(baseUri string, itemId string, r Retriever, w http.ResponseWriter) {
	item, err := r.Retrieve(itemId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	val, err := json.Marshal(item)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Write(val)
}

// TODO : Deal with URI
func (this *RestJsonAPI) retrieveAll(baseUri string, r Retriever, w http.ResponseWriter) {
	items, err := r.RetrieveAll()
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	vlog("RestJsonAPI.retrieveAll(%v):%v", baseUri, items)

	itemsHandle := NewItemsHandle(items)

	val, err := json.Marshal(itemsHandle)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Write(val)
}
