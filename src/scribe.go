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
)


/////////////////////////////////////////////////
//scribe

type Scribe struct {
	subMap        map[string]*Submission //buffered channel for creating jobs
	jobChan       chan *Job              //buffered channel for creating jobs
	subidChan     chan int               //buffered channel for use as an incrementer to keep track of submissions
	brodcastChans []chan *clientMsg
}

func NewScribe() *Scribe {
	s := Scribe{
		subMap:        map[string]*Submission{},
		jobChan:       make(chan *Job, 0),
		brodcastChans: make([]chan *clientMsg, 0, 0)}
	return &s
}

func (s *Scribe) RunScribe(hostname string, password string) {
	log("Running as scribe at %v", hostname)

	if password != "" {
		usepw = true
		hashedpw = hashPw(password)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { s.rootHandler(w, r) })
    http.Handle("/html/", http.FileServer("html","/html"))
	http.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) { s.jobsHandler(w, r) })
	http.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) { s.jobsHandler(w, r) })

	//relys on global useTls being set
	if err := ListenAndServeTLSorNot(hostname, nil); err != nil {
		log("ListenAndServeTLSorNot Error : %v", err)
		return
	}
}

//web handlers
//Root Handler.  Show JSON or redirect to HTML
func (s *Scribe) rootHandler(w http.ResponseWriter, r *http.Request) {
	log("ROOT [%v %v]", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "Hello. This is a golem scribe node:\n http://code.google.com/p/golem/")
}

// REST API [GET/POST] for all jobs and individual jobs 
func (s *Scribe) jobsHandler(w http.ResponseWriter, r *http.Request) {
	spliturl := strings.Split(r.URL.Path, "/", -1)
	nsplit := len(spliturl)
	jobid := spliturl[2]

	log("JOBS [%v %v %v %v]", r.Method, r.URL.Path, nsplit, jobid)

	w.Header().Set("Content-Type", "application/json")

	// TODO : 404 when no verbs found

	switch r.Method {
	case "GET":
		switch {
		case nsplit == 3 && jobid == "":
			s.retrieveAllJobs(w)
			return
		case nsplit == 3 && jobid != "":
			s.retrieveJob(w, jobid)
			return
		default:
			w.WriteHeader(404)
		}

	case "POST":
		// TODO : Move to password.go
		if usepw {
			pw := hashPw(r.Header.Get("Password"))
			log("Verifying password.")
			if hashedpw != pw {
				fmt.Fprint(w, "Passwords do not match.")
				return
			}
			log("Password verified")
		}

		switch {
		case nsplit == 3 && jobid == "":
			s.persistJob(w, r)
			return
		case nsplit == 3 && jobid != "":
			// operation not supported
			w.WriteHeader(501)
			return
		case nsplit == 4:
			verb := spliturl[3]
			switch verb {
			case "stop":
				s.stopJob(w, jobid)
				return
			case "delete":
				s.deleteJob(w, jobid)
				return
			default:
				w.WriteHeader(404)
			}
		}
	default:
		w.WriteHeader(404)
	}
}


func (s *Scribe) retrieveAllJobs(w http.ResponseWriter) {
	log("retrieveAllJobs")
}

func (s *Scribe) retrieveJob(w http.ResponseWriter, jobid string) {
	log("retrieveJob:%v", jobid)
}

func (s *Scribe) persistJob(w http.ResponseWriter, r *http.Request) {
	reqjson := r.FormValue("data")
	log("persistJob:%v", reqjson)
}

func (s *Scribe) stopJob(w http.ResponseWriter, jobid string) {
	log("stopJob:%v", jobid)
}

func (s *Scribe) deleteJob(w http.ResponseWriter, jobid string) {
	log("deleteJob:%v", jobid)
}
