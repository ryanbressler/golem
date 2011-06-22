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
)


/////////////////////////////////////////////////
// restOnJob


// REST interface for JobDispatchers
type RestOnJob struct {
	dispatcher JobDispatcher
}

type JobDispatcher interface {
	RetrieveAll(params map[string]string) string // json
	Retrieve(jobId string) string                // json
	NewJob(params map[string]string) string      //json
	Stop(jobId string) string                    //json
	Archive(jobId string) string                 //json
	Log(jobId string, w http.ResponseWriter)
}

// create a REST control node wrapping the given jobDispatcher
func NewRestOnJob(jd JobDispatcher) *RestOnJob {
	roj := RestOnJob{
		dispatcher: jd,
	}
	return &roj
}

// initializes the REST control node
func (j *RestOnJob) RunRestOnJob(hostname string, password string) {
	log("running at %v", hostname)

	if password != "" {
		usepw = true
		hashedpw = hashPw(password)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { j.rootHandler(w, r) })
	http.Handle("/html/", http.FileServer("html", "/html"))
	http.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) { j.jobHandler(w, r) })

	//relys on global useTls being set
	if err := ListenAndServeTLSorNot(hostname, nil); err != nil {
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

	switch r.Method {
	case "GET":
		log("Method = GET.")
		jobId, verb := parseJobRestUrl(r.URL.Path)
		switch {
		case jobId != "":
			fmt.Fprint(w, j.dispatcher.Retrieve(jobId))
		case jobId == "" && verb == "":
			fmt.Fprint(w, j.dispatcher.RetrieveAll(map[string]string{}))
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

		jobId, verb := parseJobRestUrl(r.URL.Path)
		switch {
		case jobId != "" && verb == "stop":
			fmt.Fprint(w, j.dispatcher.Stop(jobId))
		case jobId == "" && verb == "":
			fmt.Fprint(w, j.dispatcher.NewJob(map[string]string{}))
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

//parser for rest jobs request
func parseJobRestUrl(path string) (jobid string, verb string) {
	pathParts := splitRestUrl(path)
	nparts := len(pathParts)
	jobid = ""
	verb = ""
	switch {
	case nparts == 2:
		jobid = pathParts[1]
	case nparts == 3:
		jobid = pathParts[1]
		verb = pathParts[2]
	}
	vlog("Parsed job request id:\"%v\" verb:\"%v\"", jobid, verb)
	return
}

type SimpleJobDispatcher struct {

}

func (sjd SimpleJobDispatcher) RetrieveAll(params map[string]string) string {
	log("RetrieveAll")
	return "{ items:[], numberOfItems: 0, uri:'/jobs' }"
}
func (sjd SimpleJobDispatcher) Retrieve(jobId string) string {
	log("Retrieve:%v", jobId)
	return fmt.Sprintf("{ items:[], numberOfItems: 0, uri:'/jobs/%v' }", jobId)
}
func (sjd SimpleJobDispatcher) NewJob(params map[string]string) string {
	log("NewJob")
	jobId := UniqueId()
	return sjd.Retrieve(jobId)
}
func (sjd SimpleJobDispatcher) Stop(jobId string) string {
	log("Stop:%v", jobId)
	return sjd.Retrieve(jobId)
}
func (sjd SimpleJobDispatcher) Archive(jobId string) string {
	log("Archive:%v", jobId)
	return sjd.Retrieve(jobId)
}
func (sjd SimpleJobDispatcher) Log(jobId string, w http.ResponseWriter) {
	log("Log:%v", jobId)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Some output content\nYippe!!!"))
}
func NewSimpleJobDispatcher() JobDispatcher {
	return SimpleJobDispatcher{}
}
