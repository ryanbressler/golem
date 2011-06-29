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
	"websocket"
	"json"
	"strings"
	"strconv"
	"os"
)


/////////////////////////////////////////////////
//master


//THe master struct contains things that the master node needs but other nodes don't
type Master struct {
	subMap         map[string]*Submission //buffered channel for creating jobs TODO: verify thread safety... should be okay since we only set once
	jobChan        chan *Job              //buffered channel for creating jobs
	subidChan      chan int               //buffered channel for use as an incrementer to keep track of submissions
	NodeHandles    map[string]*NodeHandle
	jobController  JobController
	nodeController NodeController
}

//create a master node and initalize its channels
func NewMaster() *Master {
	m := Master{
		subMap:         map[string]*Submission{},
		jobChan:        make(chan *Job, 0),
		NodeHandles:    map[string]*NodeHandle{},
		jobController:  MasterJobController{},
		nodeController: MasterNodeController{}}
	return &m

}

//start the masters http server (load balancing starts as needed)
func (m *Master) RunMaster(hostname string, password string) {
	log("Running as master at %v", hostname)

	if password != "" {
		usepw = true
		hashedpw = hashPw(password)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { m.rootHandler(w, r) })
	http.Handle("/html/", http.FileServer("html", "/html"))
	http.HandleFunc("/jobs/", func(w http.ResponseWriter, r *http.Request) { m.jobHandler(w, r) })
	http.HandleFunc("/admin/", func(w http.ResponseWriter, r *http.Request) { m.adminHandler(w, r) })
	http.Handle("/master/", websocket.Handler(func(ws *websocket.Conn) { m.nodeHandler(ws) }))

	//relys on global useTls being set
	if err := ListenAndServeTLSorNot(hostname, nil); err != nil {
		log("ListenAndServeTLSorNot Error : %v", err)
		return
	}

}

//web handlers
//Handler for /. Nothing on root so say hello.
func (m *Master) rootHandler(w http.ResponseWriter, r *http.Request) {
	log("Root request.")
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "Hello. This is a golem master node:\n http://code.google.com/p/golem/")
}

//handler for /admin. ... controlls restarting and killing the cluster
func (m *Master) adminHandler(w http.ResponseWriter, r *http.Request) {
	log("admin request.")

	w.Header().Set("Content-Type", "text/plain")
	switch r.Method {
	case "POST":
		if usepw {
			if hashPw(r.Header.Get("Password")) != hashedpw {
				fmt.Fprintf(w, "Bad password.")
				return
			}
		}
		spliturl := splitRestUrl(r.URL.Path)
		nsplit := len(spliturl)
		switch {
		case nsplit == 2 && spliturl[1] == "restart":
			m.Broadcast(&clientMsg{Type: RESTART})
			fmt.Fprintf(w, "Restarting in 10 seconds.")
			go RestartIn(3000000000)
		case nsplit == 2 && spliturl[1] == "die":
			m.Broadcast(&clientMsg{Type: DIE})
			fmt.Fprintf(w, "Dieing in 10 seconds.")
			go DieIn(3000000000)
		case nsplit == 4 && spliturl[2] == "resize":
			nodeid := spliturl[1]
			newmax, err := strconv.Atoi(spliturl[3])
			if err != nil {
				log("error parsing new max size: %v", err)
				return
			}
			_, isin := m.NodeHandles[nodeid]
			if isin {
				m.NodeHandles[nodeid].ReSize(newmax)
				val, _ := m.NodeHandles[nodeid].MarshalJSON()
				fmt.Fprintf(w, "%v", string(val))
			} else {
				fmt.Fprintf(w, "null")
				log("Request for non node: %v", nodeid)
			}

		}
	case "GET":
		pathParts := splitRestUrl(r.URL.Path)
		nparts := len(pathParts)
		switch {
		default:
			nodedescs := make([]string, 0, len(m.NodeHandles))
			for _, n := range m.NodeHandles {
				val, _ := n.MarshalJSON()
				nodedescs = append(nodedescs, string(val))
			}
			fmt.Fprintf(w, "[%v]", strings.Join(nodedescs, ",\n"))
		case nparts == 2:
			nodeid := pathParts[1]
			_, isin := m.NodeHandles[nodeid]
			if isin {
				val, _ := m.NodeHandles[nodeid].MarshalJSON()
				fmt.Fprintf(w, "%v", string(val))
			} else {
				fmt.Fprintf(w, "null")
				log("Request for non node: %v", nodeid)
			}
		}

	}
}


//restfull api for managing jobs handled on /jobs/
//TODO: refactor the whole jobs rest interface

//parser for rest jobs request
func (m *Master) parseJobRestUrl(path string) (jobid string, verb string) {
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

//Rest full hanlders
//handler for a GET by job id /jobs/jobid
func (m *Master) jobIdGetHandler(w http.ResponseWriter, r *http.Request, subid string) {
	_, isin := m.subMap[subid]
	if isin {
		val, _ := m.subMap[subid].MarshalJSON()
		fmt.Fprintf(w, "%v", string(val))
	} else {
		fmt.Fprintf(w, "null")
		log("Request for non submission: %v", subid)
	}
}

//handler for a GET of /jobs/
func (m *Master) jobGetHandler(w http.ResponseWriter, r *http.Request) {
	jobdescs := make([]string, 0)
	for _, s := range m.subMap {
		val, _ := s.MarshalJSON()
		jobdescs = append(jobdescs, string(val))
	}
	fmt.Fprintf(w, "[%v]", strings.Join(jobdescs, ",\n"))
}

//handler for a job stop POST request
// /jobs/jobid/stop
func (m *Master) jobIdStopPostHandler(w http.ResponseWriter, r *http.Request, subid string) {
	worked := false
	_, isin := m.subMap[subid]
	if isin {
		worked = m.subMap[subid].Stop()
		fmt.Fprintf(w, "%v", worked)
	} else {

		fmt.Fprintf(w, "not found.")
		log("stop Request for non submission: %v", subid)
	}
}

func (m *Master) jobIdKillPostHandler(w http.ResponseWriter, r *http.Request, subid string) {
	worked := false
	_, isin := m.subMap[subid]
	if isin {
		log("Broadcasting kill message for SubId: %v", subid)
		m.Broadcast(&clientMsg{Type: KILL, SubId: subid})
		worked = m.subMap[subid].Stop()
		fmt.Fprintf(w, "%v", worked)
	} else {

		fmt.Fprintf(w, "not found.")
		log("stop Request for non submission: %v", subid)
	}
}

//handler for a POST of a submission to /jobs/
func (m *Master) jobPostHandler(w http.ResponseWriter, r *http.Request) {
	vlog("getting json from form")
	mpreader, err := r.MultipartReader()
	if err != nil {
		log("Error getting multipart reader: %v", err)
	}

	frm, err := mpreader.ReadForm(10000)
	if err != nil {
		log("Error reading multipart form: %v", err)
	}
	cmd := frm.Value["command"][0]
	log("command: %s", cmd)

	rJobs := make([]RequestedJob, 0, 100)
	jsonfile, err := frm.File["jsonfile"][0].Open()
	if err != nil {
		log("Error opening file from request: %v", err)
	}
	dec := json.NewDecoder(jsonfile)
	if err := dec.Decode(&rJobs); err != nil {
		fmt.Fprintf(w, "{\"Error\":\"%s\"}", err)
		log("json parse error: %v\n json: %v", err)
		return
	}
	jsonfile.Close()

	s := NewSubmission(&rJobs, m.jobChan)
	m.subMap[s.SubId] = s
	log("Created submission: %v", s.SubId)
	val, _ := s.MarshalJSON()
	fmt.Fprintf(w, "%v", string(val))
}

//actuall http handler for /jobs/ parses the url and method and sends to one of the above methods
func (m *Master) jobHandler(w http.ResponseWriter, r *http.Request) {
	log("Jobs request.")

	w.Header().Set("Content-Type", "text/plain")
	switch r.Method {
	case "GET":
		log("Method = GET.")
		subid, verb := m.parseJobRestUrl(r.URL.Path)
		switch {
		case subid != "":
			m.jobIdGetHandler(w, r, subid)
		case subid == "" && verb == "":
			m.jobGetHandler(w, r)
		default:
			fmt.Fprintf(w, "Unsupported Request")
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
		subid, verb := m.parseJobRestUrl(r.URL.Path)
		switch {
		case subid != "" && verb == "stop":
			m.jobIdStopPostHandler(w, r, subid)
		case subid != "" && verb == "kill":
			m.jobIdKillPostHandler(w, r, subid)
		case subid == "" && verb == "":
			m.jobPostHandler(w, r)
		default:
			fmt.Fprintf(w, "Unsupported Request")
		}

	}
}

//Broadcast sends a message to ever connected client
func (m *Master) Broadcast(msg *clientMsg) {
	log("Broadcasting message %v to %v nodes.", *msg, len(m.NodeHandles))
	for _, nh := range m.NodeHandles {
		nh.BroadcastChan <- msg
	}
	vlog("Broadcasting done.")

}


//goroutine to remove nodehandles from the map used to store them as they disconect
func (m *Master) RemoveNodeOnDeath(nh *NodeHandle) {
	<-nh.Con.DiedChan
	m.NodeHandles[nh.NodeId] = nh, false
}

//websocket handler for connecting nodes on /master/
// creates the nodes broadcast chan and starts monitoring it
func (m *Master) nodeHandler(ws *websocket.Conn) {
	log("Node connectiong from %v.", ws.LocalAddr().String())
	nh := NewNodeHandle(NewConnection(ws), m)
	m.NodeHandles[nh.NodeId] = nh
	go m.RemoveNodeOnDeath(nh)
	nh.Monitor()

}

// Job and Node Controllers
type MasterJobController struct {

}

func (mc MasterJobController) RetrieveAll(r *http.Request) (json string, numberOfItems int, err os.Error) {
	log("RetrieveAll")
	json = "{ items:[], numberOfItems: 0, uri:'/jobs' }"
	numberOfItems = 0
	err = nil
	return
}
func (mc MasterJobController) Retrieve(jobId string) (json string, err os.Error) {
	log("Retrieve:%v", jobId)
	json = fmt.Sprintf("{ items:[], numberOfItems: 0, uri:'/jobs/%v' }", jobId)
	err = nil
	return
}
func (mc MasterJobController) NewJob(r *http.Request) (jobId string, err os.Error) {
	log("NewJob")
	jobId = UniqueId()
	err = nil
	return
}
func (mc MasterJobController) Stop(jobId string) os.Error {
	log("Stop:%v", jobId)
	return os.NewError("unable to stop")
}
func (mc MasterJobController) Kill(jobId string) os.Error {
	log("Kill:%v", jobId)
	return os.NewError("unable to kill")
}

// Do Nothing Node Controller implementation
type MasterNodeController struct {

}

func (c MasterNodeController) RetrieveAll(r *http.Request) (json string, numberOfItems int, err os.Error) {
	log("RetrieveAll")
	json = "{ items:[], numberOfItems: 0, uri:'/nodes' }"
	numberOfItems = 0
	err = nil
	return
}
func (c MasterNodeController) Retrieve(nodeId string) (json string, err os.Error) {
	log("Retrieve:%v", nodeId)
	json = fmt.Sprintf("{ items:[], numberOfItems: 0, uri:'/nodes/%v' }", nodeId)
	err = nil
	return
}
func (c MasterNodeController) Restart(nodeId string) os.Error {
	log("Restart:%v", nodeId)
	return os.NewError("unable to restart")
}
func (c MasterNodeController) Resize(nodeId string, numberOfThreads int) os.Error {
	log("Resize:%v,%i", nodeId, numberOfThreads)
	return os.NewError("unable to resize")
}
func (c MasterNodeController) Kill(nodeId string) os.Error {
	log("Kill:%v", nodeId)
	return os.NewError("unable to kill")
}
