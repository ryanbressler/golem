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
	session       *mgo.Session
	storename     string
	jobCollection string
// jobsCollection  mgo.Collection
}

func NewMongoJobStore() *MongoJobStore {
	golemstore, err := ConfigFile.GetString("mgodb", "store")
	if err != nil {
		panic(err)
	}

	jobCollection, err := ConfigFile.GetString("mgodb", "jobcollection")
	if err != nil {
		panic(err)
	}

	session := NewMongoSession()

	return &MongoJobStore{
		session:       session,
		storename:     golemstore,
		jobCollection: jobCollection}
}

func (this *MongoJobStore) Create(item JobDetails, tasks []Task) (err os.Error) {
	vlog("MongoJobStore.Create(%v)", item)
	// TODO : Persist tasks
	this.JobsCollection().Insert(item)
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
	vlog("MongoJobStore.Get(%v)", jobId)

    m := make(map[string]interface{})

 	err = this.JobsCollection().Find(bson.M{"jobid": jobId}).One(m)
	if err != nil {
	    vlog("MongoJobStore.Get(%v):err=%v", jobId, err)
		return
	}

	vlog("MongoJobStore.Get(%v):item=%v", jobId, m)
	vlog("MongoJobStore.Get(%v):item=%v", jobId, m["handle"])

	// TODO : Populate job details

	return
}

func (this *MongoJobStore) Tasks(jobId string) (tasks []Task, err os.Error) {
	vlog("MongoJobStore.Tasks(%v)", jobId)

    m := make(map[string]interface{})

 	err = this.JobsCollection().Find(bson.M{"jobid": jobId}).One(m)
	if err != nil {
	    vlog("MongoJobStore.Tasks(%v):err=%v", jobId, err)
		return
	}

	vlog("MongoJobStore.Tasks(%v):item=%v", jobId, m)
	vlog("MongoJobStore.Tasks(%v):item=%v", jobId, m["tasks"])

	// TODO : Populate job details

	return
}

func (this *MongoJobStore) Update(item JobDetails) (err os.Error) {
	if item.JobId == "" {
		err = os.NewError("No Job Id Found")
		return
	}

	now := time.Time{}

	modifierMap := make(map[string]interface{})
	modifierMap["scheduled"] = item.Scheduled
	modifierMap["running"] = item.Running
	modifierMap["taskerrored"] = item.Errored
	modifierMap["taskfinished"] = item.Finished
	modifierMap["lastmodified"] = now.String()

    // TODO: Proper update
	err = this.JobsCollection().Update(bson.M{"jobid": item.JobId}, modifierMap)
	return
}

func (this *MongoJobStore) FindJobs(m map[string]interface{}) (items []JobDetails, err os.Error) {
	vlog("MongoJobStore.FindJobs(%v)", m)

	iter, err := this.JobsCollection().Find(m).Iter()
	if err != nil {
	    vlog("MongoJobStore.FindJobs(%v): %v", m, err)
		return
	}

	for {
		jd := JobDetails{}
		if nexterr := iter.Next(jd); nexterr != nil {
		    break
		}
		items = append(items, jd)
	}

	vlog("MongoJobStore.FindJobs: %v jobs matching %v", len(items), m)
	return
}

func (this *MongoJobStore) JobsCollection() mgo.Collection {
	db := this.session.DB(this.storename)
	return db.C(this.jobCollection)
}

func NewMongoSession() *mgo.Session {
	dbhost, err := ConfigFile.GetString("mgodb", "server")
	if err != nil {
		panic(err)
	}

	session, err := mgo.Mongo(dbhost)
	if err != nil {
		panic(err)
	}

	// Modes are Safe, Monotonic, and Strong, Strong tells the system to sync on inserts/updates
	session.SetMode(mgo.Strong, true)

	return session
}

func ParseTime(stringTime string) (tm *time.Time) {
	if stringTime == "" {
		tm = &time.Time{}
		return
	}

	tm, err := time.Parse("UnixDate", stringTime)
	if err != nil {
		log("ParseTime(%v):%v", stringTime, err)
	}
	return
}
