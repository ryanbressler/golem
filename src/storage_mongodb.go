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
	"launchpad.net/mgo"
	"launchpad.net/gobson/bson"
)

type MongoJobStore struct {
	JOBS mgo.Collection
	TASKS mgo.Collection
}

type TaskHolder struct {
    JobId string
    Tasks []Task
}

func NewMongoJobStore() *MongoJobStore {
	dbhost := ConfigFile.GetRequiredString("mgodb", "server")
	storename := ConfigFile.GetRequiredString("mgodb", "store")
	jobCollection := ConfigFile.GetRequiredString("mgodb", "jobcollection")
	taskCollection := ConfigFile.GetRequiredString("mgodb", "taskcollection")

	session, err := mgo.Mongo(dbhost)
	if err != nil {
		panic(err)
	}

	// Modes are Safe, Monotonic, and Strong, Strong tells the system to sync on inserts/updates
	session.SetMode(mgo.Strong, true)

	db := session.DB(storename)

	return &MongoJobStore{JOBS: db.C(jobCollection), TASKS: db.C(taskCollection)}
}

func (this *MongoJobStore) Create(item JobDetails, tasks []Task) (err os.Error) {
	vlog("MongoJobStore.Create(%v)", item)
	item.FirstCreated = time.LocalTime().String()
	this.JOBS.Insert(item)
	this.TASKS.Insert(TaskHolder{item.JobId, tasks})
	return
}

func (this *MongoJobStore) All() (items []JobDetails, err os.Error) {
	items, err = this.FindJobs(bson.M{})
	return
}

func (this *MongoJobStore) Unscheduled() (items []JobDetails, err os.Error) {
	items, err = this.FindJobs(bson.M{"scheduled": false})
	return
}

func (this *MongoJobStore) Active() (items []JobDetails, err os.Error) {
	items, err = this.FindJobs(bson.M{"running": true})
	return
}

func (this *MongoJobStore) Get(jobId string) (item JobDetails, err os.Error) {
	err = this.JOBS.Find(bson.M{"jobid": jobId}).One(&item)
	return
}

func (this *MongoJobStore) Tasks(jobId string) (tasks []Task, err os.Error) {
	vlog("MongoJobStore.Tasks(%v)", jobId)

    item := TaskHolder{}
	err = this.TASKS.Find(bson.M{"jobid": jobId}).One(&item)
	if err != nil {
		return
	}
	tasks = item.Tasks
	return
}

func (this *MongoJobStore) Update(item JobDetails) os.Error {
	if item.JobId == "" {
		return os.NewError("No Job Id Found")
	}

	return this.JOBS.Update(bson.M{"jobid": item.JobId}, item)
}

func (this *MongoJobStore) FindJobs(m map[string]interface{}) (items []JobDetails, err os.Error) {
	vlog("MongoJobStore.FindJobs(%v)", m)

	iter, err := this.JOBS.Find(m).Iter()
	if err != nil {
		return
	}

	for {
		jd := JobDetails{}
		if nexterr := iter.Next(&jd); nexterr != nil {
			break
		}
		items = append(items, jd)
	}

	return
}
