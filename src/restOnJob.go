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
// restOnJob


// REST interface for JobDispatchers
type RestOnJob struct {
	dispatcher JobDispatcher
	hostname   string
	password   string
}

type JobDispatcher interface {
	RetrieveAll(r *http.Request) (json string, numberOfItems int, err os.Error)
	Retrieve(jobId string) (json string, err os.Error)
	NewJob(r *http.Request) (jobId string, err os.Error)
	Stop(jobId string) (err os.Error)
	Kill(jobId string) (err os.Error)
}

// initializes the REST control node
func (j *RestOnJob) MakeReady() {
	log("running at %v", j.hostname)

	if j.password != "" {
		usepw = true
		hashedpw = hashPw(j.password)
	}

	//relys on global useTls being set
	if err := ListenAndServeTLSorNot(j.hostname, nil); err != nil {
		log("ListenAndServeTLSorNot Error : %v", err)
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { j.rootHandler(w, r) })
	http.Handle("/html/", http.FileServer("html", "/html"))
	http.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) { j.jobHandler(w, r) })
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
			json, err := j.dispatcher.Retrieve(jobId)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			fmt.Fprint(w, json)
		case jobId == "" && verb == "":
			json, _, err := j.dispatcher.RetrieveAll(r)
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
			fmt.Fprint(w, j.dispatcher.Stop(jobId))
		case jobId == "" && verb == "":
			jobId, err := j.dispatcher.NewJob(r)
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

// Do Nothing JobDispatcher implementation
type IKnowNothingJobDispatcher struct {

}

func (sjd IKnowNothingJobDispatcher) RetrieveAll(r *http.Request) (json string, numberOfItems int, err os.Error) {
	log("RetrieveAll")
	json = "{ items:[], numberOfItems: 0, uri:'/jobs' }"
	numberOfItems = 0
	err = nil
	return
}
func (sjd IKnowNothingJobDispatcher) Retrieve(jobId string) (json string, err os.Error) {
	log("Retrieve:%v", jobId)
	json = fmt.Sprintf("{ items:[], numberOfItems: 0, uri:'/jobs/%v' }", jobId)
	err = nil
	return
}
func (sjd IKnowNothingJobDispatcher) NewJob(r *http.Request) (jobId string, err os.Error) {
	log("NewJob")
	jobId = UniqueId()
	err = nil
	return
}
func (sjd IKnowNothingJobDispatcher) Stop(jobId string) os.Error {
	log("Stop:%v", jobId)
	return os.NewError("unable to stop")
}
func (sjd IKnowNothingJobDispatcher) Kill(jobId string) os.Error {
	log("Kill:%v", jobId)
	return os.NewError("unable to kill")
}
