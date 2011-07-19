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
	"fmt"
	"os"
	"time"
)

type JobStore interface {
	Create(item JobPackage) (err os.Error)

	All() (items []JobHandle, err os.Error)

	Active() (items []JobHandle, err os.Error)

	Unscheduled() (items []JobHandle, err os.Error)

	Get(jobId string) (item JobPackage, err os.Error)

	Update(jobId string, status JobStatus) (err os.Error)
}

type JobHandle struct {
	JobId        string
	Owner        string
	Label        string
	FirstCreated time.Time
	LastModified time.Time
	Status       JobStatus
}

func (h JobHandle) MarshalJSON() ([]byte, os.Error) {
	s := h.Status
	rv := fmt.Sprintf("{ uri:\"/jobs/%v\", id:\"%v\", createdAt:\"%v\", modifiedAt:\"%v\", totalTasks:%v, finishedTasks:%v, erroredTasks:%v, isRunning:%v }", h.JobId, h.JobId, s.TotalTasks, s.FinishedTasks, s.ErroredTasks, s.Running)
	return []byte(rv), nil
}

type JobStatus struct {
	TotalTasks    int
	FinishedTasks int
	ErroredTasks  int
	Running       bool
}

type JobPackage struct {
	Handle JobHandle
	Tasks  []Task
}

func (j JobPackage) MarshalJSON() ([]byte, os.Error) {
	h := j.Handle
	return h.MarshalJSON()
}

type Task struct {
	Count int
	Args  []string
}

type DoNothingJobStore struct {
	jobsById map[string]JobPackage
}

func (s DoNothingJobStore) Create(item JobPackage) (err os.Error) {
	log("Create(%v)", item)
	s.jobsById[item.Handle.JobId] = item
	return
}

func (s DoNothingJobStore) All() (items []JobHandle, err os.Error) {
	for _, item := range s.jobsById {
		items = append(items, item.Handle)
	}
	return
}

func (s DoNothingJobStore) Active() (items []JobHandle, err os.Error) {
	for _, item := range s.jobsById {
		if item.Handle.Status.Running {
			items = append(items, item.Handle)
		}
	}
	return
}

func (s DoNothingJobStore) Unscheduled() (items []JobHandle, err os.Error) {
	for _, item := range s.jobsById {
		if item.Handle.Status.Running == false {
			items = append(items, item.Handle)
		}
	}
	return
}

func (s DoNothingJobStore) Get(jobId string) (item JobPackage, err os.Error) {
	if jobId == "" {
		err = os.NewError("No job id specified")
		return
	}

	item, isin := s.jobsById[jobId]
	if isin == false {
		err = os.NewError("item not found")
	}
	return
}

func (s DoNothingJobStore) Update(jobId string, status JobStatus) (err os.Error) {
	item, err := s.Get(jobId)
	if err != nil {
		return
	}

	item.Handle.Status = status
	return
}
