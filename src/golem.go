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
	"flag"
	"http"
	"fmt"
	"websocket"
	"json"
	"strconv"
	"exec"
)

//TODO: make these server only
//buffered chanel for use as an incrementer to keep track of submissions
var subidChan = make(chan int, 10)

//buffered chanel for creating jobs
var jobChan = make(chan Job, 1000)

//buffered chanel for writing to Cerror
var coutChan = make(chan string, 1000)



const (
	//Message type constants
	HELLO        = 1
	DONE         = 2
	START        = 3
	CHECKIN      = 4
	COUT         = 5
	
	MAXMSGLENGTH = 512
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

///////////////////////////////////////////////////////////
//A node as seen from another node
//ie master to a client, client to a master...used for 
//sending and reciving
type Node struct {
	Socket  *websocket.Conn
	OutChan chan clientMsg
	InChan  chan clientMsg
}

func NewNode(Socket *websocket.Conn) *Node {
	n := Node{Socket: Socket, OutChan: make(chan clientMsg, 10), InChan: make(chan clientMsg, 10)}
	go n.GetMsgs()
	go n.SendMsgs()
	return &n
}

func (node Node) SendMsgs() {

	for {
		msg := <-node.OutChan

		msgjson, err := json.Marshal(msg)
		if err != nil {
			fmt.Printf("error json.Marshaling msg: %v\n", err)
			return
		}

		fmt.Printf("msg:%v\n", string(msgjson))
		if _, err := node.Socket.Write(msgjson); err != nil {
			fmt.Printf("Error sending msg: %v\n", err)
		}
		fmt.Printf("msg sent\n")
	}

}

func (node Node) GetMsgs() {

	for {
		var msg clientMsg
		var msgjson = make([]byte, MAXMSGLENGTH)
		n, err := node.Socket.Read(msgjson)
		if err != nil {
			fmt.Printf("error recieving client msg: %v\n", err.String())
			continue
		}
		fmt.Printf("Got msg: %v\n", string(msgjson[0:n]))

		err = json.Unmarshal(msgjson[0:n], &msg)
		if err != nil {
			fmt.Printf("error parseing client msg json: %v\n", err)
			continue
		}
		node.InChan <- msg

	}

}


/////////////////////////////////////////////////
//master

//web handlers
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
	fmt.Printf("Node connectiong.\n")
	go monitorNode(NewNode(ws))
}


//wait for a job from jobChan, turn it into a json messags
//wait for the Node socket to not be in use then send it to 
//the client. This may deadlock if the client is waiting for messages
//so the client checks in. TODO: test if the InUse lock is needed.
func sendJob(n *Node, j *Job) {
	node := *n
	job := *j
	jobjson, err := json.Marshal(job)
	if err != nil {
		fmt.Printf("error json.Marshaling job: %v\n", err)

	}
	msg := clientMsg{Type: START, Body: string(jobjson)}
	node.OutChan <- msg

}

//This waits for a handshake from a node then
//monitors messages and starts jobs as needed
func monitorNode(n *Node) {
	node := *n

	//number to run at once
	atOnce := 0
	//number running
	running := 0
	var msg clientMsg

	//wait for client handshake
	msg = <-node.InChan

	if msg.Type == HELLO {
		val, err := strconv.Atoi(msg.Body)
		if err != nil {
			fmt.Printf("error parseing client hello: %v\n", err)
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
				sendJob(&node, &job)
				running++
			case msg = <-node.InChan:
				clientMsgSwitch(&msg, &running)
			}
		default:
			msg = <-node.InChan
			clientMsgSwitch(&msg, &running)
		}

	}

}

func clientMsgSwitch(msg *clientMsg, running *int) {
	switch msg.Type {
	default:
		coutChan <- msg.Body
	case CHECKIN:
	case DONE:
		*running--
	}
}

//this monitors the out end of the coutChan and sends it where you want it
func handleCout() {
	for {
		out := <-coutChan
		fmt.Printf("cout:%v\n", out)
	}
}

func RunMaster(hostname string) {
	//start a server
	go handleCout()
	subidChan <- 0
	fmt.Printf("Running as master at %v\n", hostname)
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/jobs/", jobHandler)
	http.Handle("/master/", websocket.Handler(nodeHandler))
	if err := http.ListenAndServe(hostname, nil); err != nil {
		fmt.Printf("ListenAndServe Error : %v\n", err)
	}
}

//////////////////////////////////////////////
//node 
func startJob(replyc chan int, jsonjob string) {
	var job Job

	err := json.Unmarshal([]byte(jsonjob), &job)
	if err != nil {
		fmt.Printf("error parseing job json: %v\n", err)
	}
	jobcmd := job.Args[0]
	//make sure the path to the exec is fully qualified
	cmd, err := exec.LookPath(jobcmd)
	if err != nil {
		fmt.Printf("exec %s: %s", jobcmd, err)
	}

	args := job.Args[:]
	args = append(args, fmt.Sprintf("%v", job.SubId))
	args = append(args, fmt.Sprintf("%v", job.LineId))
	args = append(args, fmt.Sprintf("%v", job.JobId))

	//start the job in test dir pass all stdio back to main.
	//note that cmd has to be the first thing in the args array
	c, err := exec.Run(cmd, args, nil, "", exec.PassThrough, exec.PassThrough, exec.PassThrough)
	if err != nil {
		fmt.Printf("%v", err)
	}

	//wait for the job to finish
	w, err := c.Wait(0)
	if err != nil {
		fmt.Printf("%v", err)
	}

	fmt.Printf("Finishing job %v\n", job.JobId)
	//send signal back to main
	if w.Exited() && w.ExitStatus() == 0 {
		replyc <- job.JobId
	} else {
		replyc <- -1 * job.JobId
	}

}

func RunNode(atOnce int, master string) {
	running := 0
	fmt.Printf("Running as %v process node owned by %v\n", atOnce, master)

	ws, err := websocket.Dial("ws://"+master+"/master/", "", "http://localhost/")
	if err != nil {
		fmt.Printf("Error connectiong to master:%v\n", err)
		return
	}
	//ws.Write([]byte("h"))
	mnode := *NewNode(ws)
	mnode.OutChan <- clientMsg{Type: HELLO, Body: fmt.Sprintf("%v", atOnce)}

	replyc := make(chan int)

	//control loop
	for {

		select {
		case rv := <-replyc:
			fmt.Printf("Got 'done' signal: %v\n", rv)
			mnode.OutChan <- clientMsg{Type: DONE, Body: fmt.Sprintf("%v", rv)}
			running--

		case msg := <-mnode.InChan:
			switch msg.Type {
			case START:
				go startJob(replyc, msg.Body)
				running++
			}
		}

	}

}

//////////////////////////////////////////////
//main method
func main() {
	var isMaster bool
	flag.BoolVar(&isMaster, "m", false, "Start as master node.")
	var atOnce int
	flag.IntVar(&atOnce, "n", 3, "For client nodes, the number of procceses to allow at once.")
	var hostname string
	flag.StringVar(&hostname, "hostname", "localhost:8083", "The address and port of/at wich to start the master.")
	flag.Parse()

	switch isMaster {
	case true:
		RunMaster(hostname)
	default:
		RunNode(atOnce, hostname)
	}

}
