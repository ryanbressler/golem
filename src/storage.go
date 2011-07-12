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
	"os"
	"time"
)

type JobStore interface {
	Create(item JobPackage) (err os.Error)

	RetrieveAll() (items []JobHandle, err os.Error)

	Retrieve(jobId string) (item JobPackage, err os.Error)

	Update(jobId string, status JobStatus) (err os.Error)
}

type JobHandle struct {
	JobId        string
	Owner        string
	FirstCreated time.Time
	LastModified time.Time
	Status       JobStatus
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

type Task struct {
	Count     int
	Arguments []string
}
