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
	"json"
	"websocket"
	//"io/ioutil"
)

var subidChan = make(chan int, 10)
var jobChan = make(chan Job, 100)
var nodeChan = make(chan Node, 100)

type RequestedJob struct {
	Count int
	Args  []string
}

type Job struct {
	SubId  int
	LineId int
	JobId  int
	Args   []string
}

type Node struct {
	Ws *websocket.Conn
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "This is a golem master node.")
}

func jobHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Fprint(w, "Listing of jobs and statuses goes here")
	case "POST":
		go parseJobSub(r.FormValue("data"))
		fmt.Fprint(w, "Job loaded.")
	case "DEL":
		fmt.Fprint(w, "Allow users to kill jobs here.")
	}
}

func nodeHandler(ws *websocket.Conn) {
	nodeChan <- Node{Ws: ws}
}

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

func monitorRemote(in chan string, idOnNode int) {

}

func runNode(node Node) {
	/*//msg:=ioutil.ReadAll(node.ws)
	//load ms from client
	max:=3
	inchans:=[max]chan string
	for i:=0;i<max;i++{
		go monitorRemote
	}
	running := false
	var msg []byte
	var job Job
	for {
		if running == false{
			job=<-jobChan
		}
		msg=ioutil.ReadAll(node.ws)

	}*/


}

func distribute() {
	for {
		node := <-nodeChan
		go runNode(node)
	}
}

func main() {
	go distribute()
	subidChan <- 0
	fmt.Printf("Starting http Server ... ")
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/jobs/", jobHandler)
	http.Handle("/distributor/", websocket.Handler(nodeHandler))
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		fmt.Printf("ListenAndServe Error :" + err.String())
	}

}
