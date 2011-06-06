/*
   Copyright (C) 2003-2010 Institute for Systems Biology
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
)



/////////////////////////////////////////////////
//master

//web handlers
//Handler for /. Nothing on root so say hello.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "Hello. This is a golem master node:\n http://code.google.com/p/golem/")
}

//restfull api for managing jobs handled on /jobs/
func jobHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Jobs request.\n")
	w.Header().Set("Content-Type", "text/plain")
	switch r.Method {
	case "GET":
		fmt.Printf("Method = GET.\n")
		fmt.Fprint(w, "Listing of Jobs Not Yet implemented.")
	case "POST":
		fmt.Printf("Method = POST.\n")
		s := NewSubmission(r.FormValue("data"))
		subMap[s.SubId] = s
		fmt.Printf("Created submission: %v\n", s.SubId)
		fmt.Fprintf(w, "{\"SubId\":%v}", s.SubId)
	case "DEL":
		fmt.Fprint(w, "Deleting of jobs not yet implemented.")
	}
}


//start routinges to manage nodes as they conect
func nodeHandler(ws *websocket.Conn) {
	fmt.Printf("Node connectiong.\n")
	monitorNode(NewConnection(ws))
}


//wait for a job from jobChan, turn it into a json messags
//wait for the Connection socket to not be in use then send it to 
//the client. This may deadlock if the client is waiting for messages
//so the client checks in. TODO: test if the InUse lock is needed.
func sendJob(n *Connection, j *Job) {
	con := *n
	job := *j
	jobjson, err := json.Marshal(job)
	if err != nil {
		fmt.Printf("error json.Marshaling job: %v\n", err)

	}
	msg := clientMsg{Type: START, Body: string(jobjson)}
	con.OutChan <- msg

}

//This waits for a handshake from a node then
//monitors messages and starts jobs as needed
func monitorNode(n *Connection) {
	con := *n

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
			fmt.Printf("error parsing client hello: %v\n", err)
			return
		}
		atOnce = val
	} else {
		fmt.Printf("Client didn't say hello as first message.\n")
		return
	}
	fmt.Printf("Client says hello and asks for %v jobs.\n", msg.Body)

	//control loop
	for {
		switch {
		case running < atOnce:
			select {
			case job := <-jobChan:
				sendJob(&con, job)
				running++
			case msg = <-con.InChan:
				clientMsgSwitch(&msg, &running)
			}
		default:
			msg = <-con.InChan
			clientMsgSwitch(&msg, &running)
		}

	}

}

func clientMsgSwitch(msg *clientMsg, running *int) {
	switch msg.Type {
	default:
		//cout <- msg.Body
	case CHECKIN:
	case COUT:
		subMap[msg.SubId].CoutFileChan <- msg.Body
	case CERROR:
		subMap[msg.SubId].CerrFileChan <- msg.Body
	case DONE:
		*running--
	}
}


func RunMaster(hostname string, useTls bool) {
	//start a server
	subidChan <- 0
	fmt.Printf("Running as master at %v\n", hostname)
	
	if useTls {
		//tls.Listen("tcp", hostname, getTlsConfig()) 
	}
	
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/jobs/", jobHandler)
	http.Handle("/master/", websocket.Handler(nodeHandler))
	if useTls {
		certf,keyf:=getCertFiles()
		if err := http.ListenAndServeTLS(hostname,certf,keyf, nil); err != nil {
			fmt.Printf("ListenAndServe Error : %v\n", err)
		}
	}else{
		if err := http.ListenAndServe(hostname, nil); err != nil {
			fmt.Printf("ListenAndServe Error : %v\n", err)
		}
	
	}
}