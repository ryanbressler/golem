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
	"strconv"
	"strings"
)


/////////////////////////////////////////////////
//master


//THe master struct contains things that the master node needs but other nodes don't
type Master struct {
	subMap        map[string]*Submission //buffered channel for creating jobs TODO: verify thread safety... should be okay since we only set once
	jobChan       chan *Job              //buffered channel for creating jobs
	subidChan     chan int               //buffered channel for use as an incrementer to keep track of submissions
	brodcastChans []chan *clientMsg
}

//create a master node and initalize its channels
func NewMaster() *Master {
	m := Master{
		subMap:        map[string]*Submission{},
		jobChan:       make(chan *Job, 0),
		brodcastChans: make([]chan *clientMsg, 0, 0)}
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
		spliturl := strings.Split(r.URL.Path, "/", -1)
		nsplit := len(spliturl)
		if nsplit > 2 {
			switch spliturl[2] {
			case "restart":
				m.Broadcast(&clientMsg{Type: RESTART})
				fmt.Fprintf(w, "Restarting in 5 seconds.") //5 seconds is node
				go RestartIn(3000000000)
			case "die":
				m.Broadcast(&clientMsg{Type: DIE})
				fmt.Fprintf(w, "Dieing in 5 seconds.")
				go DieIn(3000000000)
			}
		}

	}
}

//restfull api for managing jobs handled on /jobs/
//TODO: refactor the whole jobs rest interface

//parser for rest jobs request
func (m *Master) parseJobRestUrl(path string) (jobid string, verb string) {
	spliturl := strings.Split(path, "/", -1)
	pathParts := make([]string, 0, 2)
	for _, part := range spliturl {
		if part != "" {
			pathParts = append(pathParts, part)
		}
	}
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
		fmt.Fprintf(w, "%v", m.subMap[subid].DescribeSelfJson())
	} else {
		fmt.Fprintf(w, "null")
		log("Request for non submission: %v", subid)
	}
}

//handler for a GET of /jobs/
func (m *Master) jobGetHandler(w http.ResponseWriter, r *http.Request) {
	jobdescs := make([]string, 0)
	for _, s := range m.subMap {
		jobdescs = append(jobdescs, s.DescribeSelfJson())
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
		if worked {
			log("Broadcasting stop message for SubId: %v", subid)
			m.Broadcast(&clientMsg{Type: STOP, SubId: subid})
		}
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
	fmt.Fprintf(w, "%v", s.DescribeSelfJson())
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
		case subid == "" && verb == "":
			m.jobPostHandler(w, r)
		default:
			fmt.Fprintf(w, "Unsupported Request")
		}

	}
}

//Broadcast sends a message to ever connected client
func (m *Master) Broadcast(msg *clientMsg) {
	log("Broadcasting message %v to %v nodes.", *msg, len(m.brodcastChans))
	for _, chn := range m.brodcastChans {
		chn <- msg
	}
	vlog("Broadcasting done.")

}

//websocket handler for connecting nodes on /master/
// creates the nodes broadcast chan and starts monitoring it
func (m *Master) nodeHandler(ws *websocket.Conn) {
	log("Node connectiong from %v.", ws.LocalAddr().String())
	bcChan := make(chan *clientMsg, 0)
	m.brodcastChans = append(m.brodcastChans, bcChan)
	m.monitorNode(NewConnection(ws), bcChan)

}


//takes a connection and job, turns job into a json messags and send it into the connections out box.
//This seems to sleep or deadlock if left alone to long so the client checks in every 
//60 seconds.
func (m *Master) sendJob(n *Connection, j *Job) {

	con := *n
	job := *j
	log("Sending job %v to %v", job, con.Socket.LocalAddr().String())
	jobjson, err := json.Marshal(job)
	if err != nil {
		log("error json.Marshaling job: %v", err)

	}
	msg := clientMsg{Type: START, Body: string(jobjson)}
	con.OutChan <- msg

}

//This waits for a handshake from a node then
//monitors messages and starts jobs as needed
//takes the connection and a chan to listen for broadcast messages on.
func (m *Master) monitorNode(n *Connection, bcChan chan *clientMsg) {
	con := *n
	nodename := con.Socket.LocalAddr().String()
	//number to run at once
	atOnce := 0
	//number running
	running := 0
	var msg clientMsg

	//wait for client handshake
	msg = <-con.InChan

	if msg.Type == HELLO {
		val, err := strconv.Atoi(msg.Body)
		if err != nil {
			log("error parsing client hello: %v", err)
			return
		}
		atOnce = val
	} else {
		log("%v didn't say hello as first message.", nodename)
		return
	}
	log("%v says hello and asks for %v jobs.", nodename, msg.Body)

	//control loop
	for {
		switch {
		case running < atOnce:
			vlog("%v has %v running. Waiting for job or message.", nodename, running)
			select {
			case bcMsg := <-bcChan:
				log("%v sending broadcast message %v", nodename, *bcMsg)
				con.OutChan <- *bcMsg
			case job := <-m.jobChan:
				m.sendJob(&con, job)
				running++
				vlog("%v got job, %v running.", nodename, running)
			case msg = <-con.InChan:
				vlog("%v Got msg", nodename)
				running = m.clientMsgSwitch(nodename, &msg, running)
				vlog("%v msg handled", nodename)
			}
		default:
			vlog("%v has %v running. Waiting for message.", nodename, running)
			select {
			case bcMsg := <-bcChan:
				log("%v sending broadcast message %v", nodename, *bcMsg)
				con.OutChan <- *bcMsg
			case msg = <-con.InChan:
				//log("Got msg from %v", nodename)
				running = m.clientMsgSwitch(nodename, &msg, running)
			}
		}

	}

}

//handle the diffrent messages a client can send and return the updated number of jobs
//that client is running
func (m *Master) clientMsgSwitch(nodename string, msg *clientMsg, running int) int {
	switch msg.Type {
	default:
		//cout <- msg.Body
	case CHECKIN:
		vlog("%v checks in", nodename)
	case COUT:
		vlog("%v got cout", nodename)
		m.subMap[msg.SubId].CoutFileChan <- msg.Body
	case CERROR:
		vlog("%v got cerror", nodename)
		m.subMap[msg.SubId].CerrFileChan <- msg.Body
	case JOBFINISHED:

		log("%v says job finished: %v running: %v", nodename, msg.Body, running)
		running--
		m.subMap[msg.SubId].FinishedChan <- NewJob(msg.Body)
		vlog("%v finished sent to Sub: %v running: %v", nodename, msg.Body, running)

	case JOBERROR:
		log("%v says job error: %v running: %v", nodename, msg.Body, running)
		running--
		m.subMap[msg.SubId].ErrorChan <- NewJob(msg.Body)
		vlog("%v finished sent to Sub: %v running: %v", nodename, msg.Body, running)

	}
	return running
}
