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
	"crypto/tls"
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
	http.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) { s.jobsHandler(w, r) })
	http.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) { s.jobsHandler(w, r) })
//	http.HandleFunc("/jobs/*", func(w http.ResponseWriter, r *http.Request) { s.jobHandler(w, r) })
//	http.HandleFunc("/jobs/*/stop", func(w http.ResponseWriter, r *http.Request) { s.stopJobHandler(w, r) })
//	http.HandleFunc("/jobs/*/delete", func(w http.ResponseWriter, r *http.Request) { s.deleteJobHandler(w, r) })
	
	// TODO : Move to tls.go
	if useTls {
		cfg := getTlsConfig()
		listener, err := tls.Listen("tcp", hostname, cfg)
		if err != nil {
			log("Listen Error : %v", err)
			return
		}

		if err := http.Serve(listener, nil); err != nil {
			log("Serve Error : %v", err)
			return
		}
	} else {
		if err := http.ListenAndServe(hostname, nil); err != nil {
			log("ListenAndServe Error : %v", err)
			return
		}
	}
}

//web handlers
//Root Handler.  Show JSON or redirect to HTML
func (m *Scribe) rootHandler(w http.ResponseWriter, r *http.Request) {
	log("root request")
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "Hello. This is a golem scribe node:\n http://code.google.com/p/golem/")
}

// REST API [GET/POST] for all jobs and individual jobs 
func (m *Scribe) jobsHandler(w http.ResponseWriter, r *http.Request) {
	log("Jobs request")

	w.Header().Set("Content-Type", "text/plain")
	switch r.Method {
		case "GET":
			log("Method = GET.")
			spliturl := strings.Split(r.URL.Path, "/", -1)
			nsplit := len(spliturl)
			vlog("path: %v, nsplit: %v %v %v %v", r.URL.Path, nsplit, spliturl[0], spliturl[1], spliturl[2])
			switch {
				case nsplit == 3 && spliturl[2] == "":
					vlog("jobs")					
				case nsplit == 3:
					subid := spliturl[2]
					vlog("job: %v", subid)
				case nsplit == 4:
				    subid := spliturl[3]
					vlog("job verbs: %v", subid)
			}
	
		case "POST":
			log("Method = POST.")
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
	
			vlog("getting json from form")
			reqjson := r.FormValue("data")
			vlog("json is : %v", reqjson)
			// TODO : Persist JSON to database
	}
}
