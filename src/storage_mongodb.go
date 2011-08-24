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
	database mgo.Database
	JOBS     string
	TASKS    string
}

func NewMongoJobStore() *MongoJobStore {
	dbhost := ConfigFile.GetRequiredString("mgodb", "server")
	storename := ConfigFile.GetRequiredString("mgodb", "store")
	jobCollection := ConfigFile.GetRequiredString("mgodb", "jobcollection")
	taskCollection := ConfigFile.GetRequiredString("mgodb", "taskcollection")
	log("NewMongoJobStore():%v:%v:%v:%v", dbhost, storename, jobCollection, taskCollection)

	session, err := mgo.Mongo(dbhost)
	if err != nil {
		log("NewMongoJobStore():%v", err.String())
	}

	session.SetMode(mgo.Strong, true) // [Safe, Monotonic, Strong] Strong syncs on inserts/updates

	db := session.DB(storename)
	return &MongoJobStore{database: db, JOBS: jobCollection, TASKS: taskCollection}
}

func (this *MongoJobStore) DbJobs() mgo.Collection {
	return this.database.C(this.JOBS)
}
func (this *MongoJobStore) DbTasks() mgo.Collection {
	return this.database.C(this.TASKS)
}
func (this *MongoJobStore) Create(item JobDetails, tasks []Task) os.Error {
	vlog("MongoJobStore.Create(%v)", item)
	err := this.DbJobs().Insert(item)
	if err != nil {
		return err
	}
	return this.DbTasks().Insert(TaskHolder{item.JobId, tasks})
}

func (this *MongoJobStore) All() ([]JobDetails, os.Error) {
	return this.FindJobs(bson.M{})
}

func (this *MongoJobStore) Unscheduled() ([]JobDetails, os.Error) {
	return this.FindJobs(bson.M{"scheduled": false})
}

func (this *MongoJobStore) Active() ([]JobDetails, os.Error) {
	return this.FindJobs(bson.M{"running": true})
}

func (this *MongoJobStore) Get(jobId string) (item JobDetails, err os.Error) {
	err = this.DbJobs().Find(bson.M{"jobid": jobId}).One(&item)
	return
}

func (this *MongoJobStore) Tasks(jobId string) (tasks []Task, err os.Error) {
	item := TaskHolder{}
	err = this.DbTasks().Find(bson.M{"jobid": jobId}).One(&item)
	if err == nil {
		tasks = item.Tasks
	}
	return
}

func (this *MongoJobStore) Update(item JobDetails) os.Error {
	if item.JobId == "" {
		return os.NewError("No Job Id Found")
	}

	existing, err := this.Get(item.JobId)
	if err != nil {
		return err
	}

	existing.LastModified = time.LocalTime().String()
	existing.Progress.Finished = item.Progress.Finished
	existing.Progress.Errored = item.Progress.Errored
	existing.Running = item.Running
	existing.Scheduled = item.Scheduled

	return this.DbJobs().Update(bson.M{"jobid": item.JobId}, existing)
}

func (this *MongoJobStore) FindJobs(m map[string]interface{}) (items []JobDetails, err os.Error) {
	iter, err := this.DbJobs().Find(m).Iter()
	if err != nil {
		log("MongoJobStore.FindJobs(%v):%v", m, err)
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
