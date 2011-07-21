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

type ItemsHandle struct {
	Items         []interface{}
	NumberOfItems int
}

func NewItemsHandle(items []interface{}) ItemsHandle {
	return ItemsHandle{Items: items, NumberOfItems: len(items)}
}

type JobSubmission struct {
	Uri          string
	SubId        string
	CreatedAt    string
	TotalJobs    int
	FinishedJobs int
	ErroredJobs  int
	Running      bool
}

func NewJobSubmission(s *Submission) JobSubmission {
	total, finished, errored, running := s.Stats()

	return JobSubmission{Uri: s.Uri, SubId: s.SubId, CreatedAt: s.SubLocalTime,
		TotalJobs: total, FinishedJobs: finished, ErroredJobs: errored, Running: running}
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
