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
	"io/ioutil"
	"strconv"
	"os"
)

//buffered chanel for use as an incrementer to keep track of submissions
var subidChan = make(chan int, 10)

//buffered chanel for creating jobs
var jobChan = make(chan Job, 1000)

//buffered chanel for writing to Cerror
var coutChan = make(chan string, 1000)


//Message type constants
const (
	HELLO = 1
	DONE = 2
	START = 3
	CHECKIN = 4
	COUT = 5
	
)
	
//structs
//messages sent between server and client
type clientMsg struct {
	Type int
	Body string
}

//job requested over rest api
type RequestedJob struct {
	Count int
	Args  []string
}

//Internal Job Representation
type Job struct {
	SubId  int
	LineId int
	JobId  int
	Args   []string
}

//A node as represented by a websocket connection
//and an unbuffered channel to represent weather the websocket is in use
type Node struct {
	Socket *websocket.Conn
	InUse chan bool
}


//web handelers
//Handler for /. Nothing on root so say hello.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello. This is a golem master node:\n http://code.google.com/p/golem/")
}

//restfull api for managing jobs handled on /jobs/
func jobHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Fprint(w, "Listing of Jobs Not Yet implemented.")
	case "POST":
		go parseJobSub(r.FormValue("data"))
		fmt.Fprint(w, "Job loaded.")
	case "DEL":
		fmt.Fprint(w, "Deleting of jobs not yet implemented.")
	}
}

//interprets a post request of jobs to run
func parseJobSub(reqjson string) {
	rJobs := make([]RequestedJob, 0, 100)
	if err := json.Unmarshal([]byte(reqjson), &rJobs); err != nil {
		fmt.Printf("%v", err)
	}
	subId := <-subidChan
	subidChan <- subId + 1
	jobId := 0
	for lineId, vals := range rJobs {
		for i := 0; i < vals.Count; i++ {
			jobChan <- Job{SubId: subId, LineId: lineId, JobId: jobId, Args: vals.Args}
		}
	}

}

//start routinges to manage nodes as they conect
func nodeHandler(ws *websocket.Conn) {
	go runNode(&Node{Socket: ws, InUse: make(chan bool) })
}


//wait for a job from jobChan, turn it into a json messags
//wait for the Node socket to not be in use then send it to 
//the client. This may deadlock if the client is waiting for messages
//so the client checks in. TODO: test if the InUse lock is needed.
func startJob(n *Node) {
	node := *n
	job :=<- jobChan
	jobjson, err := json.Marshal(job)
	if err != nil {
		fmt.Printf("error json.Marshaling job: %v", err)
		
	}
	msg := clientMsg{Type:START,Body:string(jobjson)}
	msgjson, err := json.Marshal(msg)
	if err != nil {
		fmt.Printf("error json.Marshaling msg: %v", err)
		
	}
	
	node.InUse<-true
	fmt.Fprint(node.Socket, string(msgjson))
	<-node.InUse
}

//This waits for a handshake from a node then
//monitors messages and starts jobs as needed
func runNode(n *Node) {
	node := *n
	
	//number to run at once
	atOnce := 0
	//number running
	running := 0
	var msg clientMsg
	
	//wait for client handshake
	msgjson, err := ioutil.ReadAll(node.Socket)
	if err != nil && err != os.EOF {
			fmt.Printf("error contacting client: %v", err)
			return
	}
	err = json.Unmarshal(msgjson, &msg)
	if  err != nil {
		fmt.Printf("error parseing json: %v", err)
		return
	}
	
	if msg.Type == HELLO {
		val, err := strconv.Atoi(msg.Body)
		if err != nil {
			fmt.Printf("error parseing client hello: %v", err)
			return
		}
		atOnce = val
	} else {
		fmt.Printf("Client didn't say hello as first message.")
		return
	}
	
	
	

	//control loop
	for {
		switch {
		case running < atOnce: 
			//start a job
			go startJob(&node)
			running++
		default:
			//this will cause a deadlock where no jobs are started
			//if client doesn't check in
			node.InUse<-true
			msgjson, err := ioutil.ReadAll(node.Socket); 
			if err != nil && err != os.EOF {
				fmt.Printf("error recieving client msg: %v", err)
				continue
			} 
			<-node.InUse
			err = json.Unmarshal(msgjson, &msg); 
			if err != nil {
				fmt.Printf("error parseing client msg json: %v", err)
				continue
			}
				
			switch msg.Type {
			default: 
				coutChan<-msg.Body
			case CHECKIN:
				continue
			case DONE: 
				running--			
			
			}
		}

	}

}

//this monitors the out end of the coutChan and sends it where you want it
func handleCout() {
	for {
		out :=<-coutChan
		fmt.Printf("cout:%v\n", out)
	}
}

//main methos
func main() {

	//start a server
	go handleCout()
	subidChan <- 0
	fmt.Printf("Starting http Server ... ")
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/jobs/", jobHandler)
	http.Handle("/distributor/", websocket.Handler(nodeHandler))
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		fmt.Printf("ListenAndServe Error :" + err.String())
	}

}
