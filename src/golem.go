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
	"os"
	"bufio"
)

//TODO: move these three global variables into a master object so they aren't decalred
//when running as a client
//buffered channel for use as an incrementer to keep track of submissions
var subidChan = make(chan int, 1)

//buffered channel for creating jobs
var jobChan = make(chan *Job, 1000)


//map of submissions by id
var subMap = map[int]*Submission{}


const (
	//Message type constants
	HELLO   = 1
	DONE    = 2
	START   = 3
	CHECKIN = 4

	COUT   = 5
	CERROR = 6
)

//structs
//messages sent between server and client
type clientMsg struct {
	Type  int
	SubId int
	Body  string
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
//Connection represents one end of a websocket and has facilities 
//For sending and recieving connections
//TODO:Needs code to clean up after it is done...

type Connection struct {
	Socket  *websocket.Conn
	OutChan chan clientMsg
	InChan  chan clientMsg
}

func NewConnection(Socket *websocket.Conn) *Connection {
	n := Connection{Socket: Socket, OutChan: make(chan clientMsg, 10), InChan: make(chan clientMsg, 10)}
	go n.GetMsgs()
	go n.SendMsgs()
	return &n
}

func (con Connection) SendMsgs() {

	for {
		msg := <-con.OutChan

		msgjson, err := json.Marshal(msg)
		if err != nil {
			fmt.Printf("error json.Marshaling msg: %v\n", err)
			return
		}

		//fmt.Printf("sending:%v\n", string(msgjson))
		if _, err := con.Socket.Write(msgjson); err != nil {
			fmt.Printf("Error sending msg: %v\n", err)
		}
		//fmt.Printf("msg sent\n")
	}

}

func (con Connection) GetMsgs() {
	decoder := json.NewDecoder(con.Socket)
	for {
		var msg clientMsg
		err := decoder.Decode(&msg)
		switch {
		case err == os.EOF:
			fmt.Printf("EOF recieved on websocket.")
			con.Socket.Close()
			return //TODO: recover
		case err != nil:
			fmt.Printf("error parseing client msg json: %v\n", err)
			continue
		}
		con.InChan <- msg

	}

}


////////////////////////////////////////////////////
/// Submission
/// TODO: Kill submissions once they finish.

type Submission struct {
	SubId        int
	CoutFileChan chan string
	CerrFileChan chan string
	Jobs         []RequestedJob
}


func NewSubmission(reqjson string) *Submission {
	rJobs := make([]RequestedJob, 0, 100)
	if err := json.Unmarshal([]byte(reqjson), &rJobs); err != nil {
		fmt.Printf("%v", err)
	}
	subId := <-subidChan
	subidChan <- subId + 1
	s := Submission{SubId: subId, CoutFileChan: make(chan string, 10), CerrFileChan: make(chan string, 10), Jobs: rJobs}
	go s.writeIo()
	jobId := 0
	for lineId, vals := range rJobs {
		for i := 0; i < vals.Count; i++ {
			jobChan <- &Job{SubId: subId, LineId: lineId, JobId: jobId, Args: vals.Args}
			jobId++
		}
	}
	return &s

}

func (s Submission) writeIo() {
	outf, err := os.Create(fmt.Sprintf("%v.out.txt", s.SubId))
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
	}
	defer outf.Close()
	errf, err := os.Create(fmt.Sprintf("%v.err.txt", s.SubId))
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
	}
	defer errf.Close()

	for {
		select {
		case msg := <-s.CoutFileChan:
			fmt.Fprint(outf, msg)
		case errmsg := <-s.CerrFileChan:
			fmt.Fprint(errf, errmsg)

		}
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


func RunMaster(hostname string) {
	//start a server
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


func pipeToChan(p *os.File, msgType int, Id int, ch chan clientMsg) {
	bp := bufio.NewReader(p)

	for {
		line, err := bp.ReadString('\n')
		if err != nil {
			return
		} else {
			ch <- clientMsg{Type: msgType, SubId: Id, Body: line} //string(buffer[0:n])
		}
	}

}

func startJob(cn *Connection, replyc chan int, jsonjob string) {
	var job Job
	con := *cn
	err := json.Unmarshal([]byte(jsonjob), &job)
	if err != nil {
		fmt.Printf("error parseing job json: %v\n", err)
	}
	jobcmd := job.Args[0]
	//make sure the path to the exec is fully qualified
	cmd, err := exec.LookPath(jobcmd)
	if err != nil {
		con.OutChan <- clientMsg{Type: CERROR, SubId: job.SubId, Body: fmt.Sprintf("Error finding %s: %s\n", jobcmd, err)}
		fmt.Printf("exec %s: %s\n", jobcmd, err)
		replyc <- -1 * job.JobId
		return
	}

	args := job.Args[:]
	args = append(args, fmt.Sprintf("%v", job.SubId))
	args = append(args, fmt.Sprintf("%v", job.LineId))
	args = append(args, fmt.Sprintf("%v", job.JobId))

	//start the job in test dir pass all stdio back to main.
	//note that cmd has to be the first thing in the args array
	c, err := exec.Run(cmd, args, nil, "./", exec.DevNull, exec.Pipe, exec.Pipe)
	if err != nil {
		fmt.Printf("%v", err)
	}
	go pipeToChan(c.Stdout, COUT, job.SubId, con.OutChan)
	go pipeToChan(c.Stderr, CERROR, job.SubId, con.OutChan)
	//wait for the job to finish
	w, err := c.Wait(0)
	if err != nil {
		fmt.Printf("joberror:%v", err)
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
	mcon := *NewConnection(ws)
	//go sendCio(&mcon)
	mcon.OutChan <- clientMsg{Type: HELLO, Body: fmt.Sprintf("%v", atOnce)}

	replyc := make(chan int)

	//control loop
	for {

		select {
		case rv := <-replyc:
			fmt.Printf("Got 'done' signal: %v\n", rv)
			mcon.OutChan <- clientMsg{Type: DONE, Body: fmt.Sprintf("%v", rv)}
			running--

		case msg := <-mcon.InChan:
			switch msg.Type {
			case START:
				go startJob(&mcon, replyc, msg.Body)
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
