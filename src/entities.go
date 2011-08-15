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
	"time"
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

	Running   bool
	Scheduled bool
}

type TaskProgress struct {
	Total    int
	Finished int
	Errored  int
}

func (this *TaskProgress) isComplete() bool {
	return this.Total <= (this.Finished + this.Errored)
}

func NewJobDetails(jobId string, owner string, label string, jobtype string, totalTasks int) JobDetails {
	return JobDetails{
		JobId: jobId, Uri: "/jobs/" + jobId,
		Owner: owner, Label: label, Type: jobtype,
		FirstCreated: time.LocalTime().String(),
		Progress:     TaskProgress{Total: totalTasks, Finished: 0, Errored: 0},
		Running:      false, Scheduled: false}
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

	COUT   //cout from worker, body is line of cout
	CERROR //cout from worker, body is line of cerror

	JOBFINISHED //sent from worker on job finish, body is json job SubId set
	JOBERROR    //sent from worker on job error, body is json job, SubId set

	RESTART //Sent by master to nodes telling them to resart and reconnec themselves.
	DIE     //tell nodes to shutdown.
)

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
	maxJobs, running := nh.Stats()
	return WorkerNode{NodeId: nh.NodeId, Uri: nh.Uri, Hostname: nh.Hostname,
		MaxJobs: maxJobs, RunningJobs: running, Running: (running > 0)}
}

type WorkerMessage struct {
	Type   int
	SubId  string
	Body   string
	ErrMsg string
}
