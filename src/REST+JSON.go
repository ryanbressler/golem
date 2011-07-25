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
	"io"
	"os"
	"strconv"
)

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
		hpw = GetHashKey(password)
	}

	http.Handle("/", &RootHandler{})
	http.Handle("/html/", http.FileServer("html", "/html"))
	http.Handle("/jobs/", &JobsRestJson{jc, hpw})
	http.Handle("/nodes/", &NodesRestJson{nc, hpw})

	log("running at %v", hostname)

	if err := ListenAndServeTLSorNot(hostname, nil); err != nil {
		panic(err)
	}
	return
}

type RootHandler struct {

}

func (this *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log("%v /", r.Method)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{ jobs: '/jobs', nodes: '/nodes' }"))
}

type JobsRestJson struct {
	jobController JobController
	hashedpw      string
}

func (this *JobsRestJson) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log("%v /jobs", r.Method)

	w.Header().Set("Content-Type", "application/json")

	// TODO : Add logic to retrieve outputs from job

	switch r.Method {
	case "GET":
		log("Method = GET.")
		jobId, verb := parseJobUri(r.URL.Path)
		switch {
		case jobId != "":
			WriteItemAsJson("/jobs", jobId, this.jobController, w)
		case jobId == "" && verb == "":
			WriteItemsAsJson("/jobs", this.jobController, w)
		default:
			w.WriteHeader(http.StatusNotFound)
		}

	case "POST":
		log("Method = POST.")
		if CheckApiKey(this.hashedpw, r) == false {
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

			WriteItemAsJson("/jobs", jobId, this.jobController, w)
		default:
			w.WriteHeader(http.StatusNotFound)
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

type NodesRestJson struct {
	nodeController NodeController
	hashedpw       string
}

func (this *NodesRestJson) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log("NodesRestJson.ServeHTTP(%v %v)", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "application/json")

	spliturl := splitRestUrl(r.URL.Path)
	nparts := len(spliturl)

	switch r.Method {
	case "GET":
		switch {
		case nparts == 2:
			WriteItemAsJson("/nodes", spliturl[1], this.nodeController, w)
		default:
			WriteItemsAsJson("/nodes", this.nodeController, w)
		}
	case "POST":
		if CheckApiKey(this.hashedpw, r) == false {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		switch {
		case nparts == 2 && spliturl[1] == "restart":
			if err := this.nodeController.RestartAll(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		case nparts == 2 && spliturl[1] == "die":
			if err := this.nodeController.KillAll(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		case nparts == 4 && spliturl[2] == "resize":
			numberOfThreads, err := strconv.Atoi(spliturl[3])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			nodeId := spliturl[1]
			if err = this.nodeController.Resize(nodeId, numberOfThreads); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}

func CheckApiKey(hashedpw string, r *http.Request) bool {
	if hashedpw != "" {
		apikey := r.Header.Get("x-golem-apikey")
		if apikey == "" {
			return false
		}
		pw := GetHashKey(apikey)
		return hashedpw == pw
	}
	return true
}

func GetHashKey(apikey string) string {
	hash.Reset()
	io.WriteString(hash, apikey) //TODO: plus salt, or whatever
	return string(hash.Sum())
}
