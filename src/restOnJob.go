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
	"strconv"
)


/////////////////////////////////////////////////
// restOnJob


// REST interface for Job and Node Controllers
type RestOnJob struct {
	jobController  JobController
	nodeController NodeController
	hostname       string
	password       string
}

type JobController interface {
	RetrieveAll(r *http.Request) (json string, numberOfItems int, err os.Error)
	Retrieve(jobId string) (json string, err os.Error)
	NewJob(r *http.Request) (jobId string, err os.Error)
	Stop(jobId string) (err os.Error)
	Kill(jobId string) (err os.Error)
}

type NodeController interface {
	RetrieveAll(r *http.Request) (json string, numberOfItems int, err os.Error)
	Retrieve(nodeId string) (json string, err os.Error)
	Restart() os.Error
	Kill() os.Error
	Resize(nodeId string, numberOfThreads int) os.Error
}

// initializes the REST control node
func (j *RestOnJob) MakeReady() {
	log("running at %v", j.hostname)

	if j.password != "" {
		usepw = true
		hashedpw = hashPw(j.password)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { j.rootHandler(w, r) })
	http.Handle("/html/", http.FileServer("html", "/html"))
	http.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) { j.jobHandler(w, r) })
	http.HandleFunc("/admin/", func(w http.ResponseWriter, r *http.Request) { j.nodeHandler(w, r) })
	http.HandleFunc("/nodes/", func(w http.ResponseWriter, r *http.Request) { j.nodeHandler(w, r) })

	//relys on global useTls being set
	if err := ListenAndServeTLSorNot(j.hostname, nil); err != nil {
		log("ListenAndServeTLSorNot Error : %v", err)
		return
	}
}

// web handlers
func (j *RestOnJob) rootHandler(w http.ResponseWriter, r *http.Request) {
	log("%v /", r.Method)
	w.Header().Set("Content-Type", "text/plain")
	// w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "{ jobs: '/jobs', nodes: '/nodes' }")
}

func (j *RestOnJob) jobHandler(w http.ResponseWriter, r *http.Request) {
	log("%v /jobs", r.Method)

	w.Header().Set("Content-Type", "text/plain")
	// w.Header().Set("Content-Type", "application/json")

	// TODO : Add logic to retrieve outputs from job
	// TODO : Manage errors

	switch r.Method {
	case "GET":
		log("Method = GET.")
		jobId, verb := parseJobUri(r.URL.Path)
		switch {
		case jobId != "":
			json, err := j.jobController.Retrieve(jobId)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			fmt.Fprint(w, json)
		case jobId == "" && verb == "":
			json, _, err := j.jobController.RetrieveAll(r)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			fmt.Fprint(w, json)
		default:
			w.WriteHeader(http.StatusNotFound)
		}

	case "POST":
		log("Method = POST.")
		if usepw {
			pw := hashPw(r.Header.Get("Password"))
			log("Verifying password.")
			if hashedpw != pw {
				fmt.Fprint(w, "Passwords do not match.")
				return
			}
			log("Password verified")
		}

		jobId, verb := parseJobUri(r.URL.Path)
		switch {
		case jobId != "" && verb == "stop":
			err := j.jobController.Stop(jobId)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
		case jobId == "" && verb == "":
			jobId, err := j.jobController.NewJob(r)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			fmt.Fprint(w, "{ uri: '/jobs/%v' id:'%v' )", jobId)
		default:
			w.WriteHeader(http.StatusNotFound)
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (j *RestOnJob) nodeHandler(w http.ResponseWriter, r *http.Request) {
	log("nodeHandler")

	w.Header().Set("Content-Type", "text/plain")
	// w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		pathParts := splitRestUrl(r.URL.Path)
		nparts := len(pathParts)
		switch {
		case nparts == 2:
			nodeId := pathParts[1]
			json, err := j.nodeController.Retrieve(nodeId)
			if err != nil {
				w.WriteHeader(404)
			} else {
				fmt.Fprint(w, json)
			}
			return
		default:
			json, _, err := j.nodeController.RetrieveAll(r)
			if err != nil {
				w.WriteHeader(500)
			} else {
				fmt.Fprintf(w, "{ items:[%v] }", json)
			}
			return
		}
	case "POST":
		if usepw {
			if hashPw(r.Header.Get("Password")) != hashedpw {
				fmt.Fprintf(w, "Bad password.")
				return
			}
		}

		err := j.postNodeHandler(r)
		if err != nil {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(200)
		}
	}
}

func (j *RestOnJob) postNodeHandler(r *http.Request) os.Error {
	spliturl := splitRestUrl(r.URL.Path)
	nsplit := len(spliturl)
	switch {
	case nsplit == 2 && spliturl[1] == "restart":
		return j.nodeController.Restart()

	case nsplit == 2 && spliturl[1] == "die":
		return j.nodeController.Kill()

	case nsplit == 4 && spliturl[2] == "resize":
		nodeId := spliturl[1]
		numberOfThreads, err := strconv.Atoi(spliturl[3])
		if err != nil {
			return err
		}
		return j.nodeController.Resize(nodeId, numberOfThreads)
	}
	return nil
}
