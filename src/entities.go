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
	"json"
	"time"
	"fmt"
	"os"
)

type ItemsHandle struct {
	Items         []interface{}
	NumberOfItems int
}

// jobs
type JobDetailsList struct {
	Items         []JobDetails
	NumberOfItems int
}

type TaskHolder struct {
	JobId string
	Tasks []Task
}

type Task struct {
	Count int
	Args  []string
}

type JobDetails struct {
	JobId string
	Uri   string

	Owner string
	Label string
	Type  string

	FirstCreated string
	LastModified string

	Progress TaskProgress

	State  string // job state
	Status string // job status
}

func (this JobDetails) IsRunning() bool {
	return this.State == RUNNING
}

// job state
const (
	NEW       = "NEW"       // job received and stored
	SCHEDULED = "SCHEDULED" // job placed in queue
	RUNNING   = "RUNNING"   // job assigned to worker
	COMPLETE  = "COMPLETE"  // job is finished
)

// job status
const (
	READY   = "READY"   // NEW, SCHEDULED, RUNNING job
	SUCCESS = "SUCCESS" // COMPLETE job
	FAIL    = "FAIL"    // COMPLETE job
	ERROR   = "ERROR"   // COMPLETE job
	STOPPED = "STOPPED" // COMPLETE job
)

type TaskProgress struct {
	Total    int
	Finished int
	Errored  int
}

func (this *TaskProgress) isComplete() bool {
	return this.Total <= (this.Finished + this.Errored)
}

func NewJobDetails(jobId string, owner string, label string, jobtype string, totalTasks int, state string, status string) JobDetails {
	return JobDetails{
		JobId: jobId, Uri: "/jobs/" + jobId,
		Owner: owner, Label: label, Type: jobtype,
		FirstCreated: time.LocalTime().String(),
		Progress:     TaskProgress{Total: totalTasks, Finished: 0, Errored: 0},
		State:        state, Status: status}
}

func TotalTasks(tasks []Task) (totalTasks int) {
	for _, task := range tasks {
		totalTasks += task.Count
	}
	return
}

// workers
const (
	HELLO   = iota //sent from worker to master on connect, body is number of processes available
	CHECKIN        //sent from worker every minute to keep connection alive

	START //sent from master to start job, body is json job
	KILL  //sent from master to stop jobs, SubId indicates what jobs to stop.

	COUT   //standard out line from worker
	CERROR //standard err line from worker

	JOBFINISHED //sent from worker on job finish, body is json job SubId set
	JOBERROR    //sent from worker on job error, body is json job, SubId set

	RESTART //Sent by master to nodes telling them to restart and reconnect themselves.
	DIE     //tell nodes to shutdown.
)

type HelloMsgBody struct {
	JobCapacity int
	RunningJobs int
	UniqueId    string
}

func NewHelloMsgBody(data string) (rv HelloMsgBody, err os.Error) {
	err = json.Unmarshal([]byte(data), rv)
	return
}

type WorkerNodeList struct {
	Items         []WorkerNode
	NumberOfItems int
}

type WorkerNode struct {
	NodeId      string
	Uri         string
	Hostname    string
	MaxJobs     int
	RunningJobs int
	Running     bool
}

func NewWorkerNode(nh *NodeHandle) WorkerNode {
	logger.Debug("NewWorkerNode()")
	maxJobs, running := nh.Stats()
	logger.Debug("creating new worker: %d,%d", maxJobs, running)
	return WorkerNode{NodeId: nh.NodeId, Uri: nh.Uri, Hostname: nh.Hostname,
		MaxJobs: maxJobs, RunningJobs: running, Running: (running > 0)}
}

type WorkerMessage struct {
	Type   int
	SubId  string
	Body   string
	ErrMsg string
}

func (wm *WorkerMessage) BodyFromInterface(Body interface{}) os.Error {
	b, err := json.Marshal(Body)
	if err != nil {
		logger.Warn(err)
		return err
	}
	wm.Body = fmt.Sprintf("%v", b)
	return nil

}

//Internal Job Representation used primarily as the body of job related messages
type WorkerJob struct {
	SubId  string
	LineId int
	JobId  int
	Args   []string
}

// NewJob creates a job from a json string (usually a message body)
func NewWorkerJob(jsonjob string) (job *WorkerJob) {
	logger.Debug("NewWorkerJob(%v)", jsonjob)
	if err := json.Unmarshal([]byte(jsonjob), &job); err != nil {
		logger.Warn(err)
	}
	return
}

// cluster stats
type ClusterStatList struct {
	Items         []ClusterStat
	NumberOfItems int
}

type ClusterStat struct {
	SnapshotAt       int64
	JobsRunning      int
	JobsPending      int
	WorkersRunning   int
	WorkersAvailable int
}

func NewClusterStat(running int, pending int, workers int, available int) ClusterStat {
	return ClusterStat{SnapshotAt: time.Seconds(),
		JobsRunning: running, JobsPending: pending,
		WorkersRunning: workers, WorkersAvailable: available}
}
