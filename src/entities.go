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

type JobDetailsList struct {
    Items []JobDetails
    NumberOfItems int
}

type WorkerNode struct {
	NodeId   string
	Uri      string
	Hostname string
	MaxJobs  int
	Running  int
}

func NewWorkerNode(nh *NodeHandle) WorkerNode {
	maxJobs, running := nh.Stats()
	return WorkerNode{NodeId: nh.NodeId, Uri: nh.Uri, Hostname: nh.Hostname,
		MaxJobs: maxJobs, Running: running}
}

// reformed entities
type Identity struct {
    JobId string
    Uri string
}

type Description struct {
    Owner string
    Label string
    Type string
}

type Timing struct {
	FirstCreated *time.Time
	LastModified *time.Time
}

type Progress struct {
	Total  int
	Finished  int
	Errored  int
}

type Status struct {
    // TODO : State
    Running       bool
    Scheduled   bool
}

type Task struct {
	Count int
	Args  []string
}

type JobDetails struct {
    Identity Identity
    Description Description
    Timing Timing
    Progress Progress
    Status Status
    Tasks []Task
}

func InitialStatus() Status {
    return Status{false, false}
}
func InitialProgress(tasks []Task) Progress {
	totalTasks := 0
	for _, task := range tasks {
		totalTasks += task.Count
	}
    return Progress{Total:totalTasks,Finished:0,Errored:0}
}
func InitialTiming() Timing {
    now := time.Time{}
    return Timing{&now, &now}
}

func (this Progress) isComplete() bool {
    return this.Total == (this.Finished + this.Errored)
}