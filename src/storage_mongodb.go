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

type GolemJobC struct {
	Type          string
	Id            string
	Owner         string
	Label         string
	TimeCreated   string
	LastModified  string
	TaskCount     int
	TasksFinished int
	TasksErrored  int
	Running       bool
	Scheduled     bool
}

func (this GolemJobC) ToJobHandle() (handle JobHandle) {
	status := JobStatus{
		TotalTasks:    this.TaskCount,
		FinishedTasks: this.TasksFinished,
		ErroredTasks:  this.TasksErrored,
		Running:       this.Running}

	handle = JobHandle{
		JobId:        this.Id,
		Owner:        this.Owner,
		Label:        this.Label,
		Type:         this.Type,
		FirstCreated: ParseTime(this.TimeCreated),
		LastModified: ParseTime(this.LastModified),
		Status:       status}
	return
}

type MongoJobStore struct {
	session       *mgo.Session
	storename     string
	jobCollection string
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

func (this *MongoJobStore) Create(item JobPackage) (err os.Error) {
	log("Create(%v)", item)

	handle := item.Handle
	status := handle.Status
	golemJob := GolemJobC{
		Type:          handle.Type,
		Id:            handle.JobId,
		Owner:         handle.Owner,
		Label:         handle.Label,
		TimeCreated:   handle.FirstCreated.String(),
		LastModified:  handle.LastModified.String(),
		TaskCount:     status.TotalTasks,
		TasksFinished: status.FinishedTasks,
		TasksErrored:  status.ErroredTasks,
		Running:       status.Running,
		Scheduled:     false}

	this.JobsCollection().Insert(golemJob)

	return
}

func (this *MongoJobStore) All() (items []JobHandle, err os.Error) {
	items, err = this.FindJobs(bson.M{})
	return
}

func (this *MongoJobStore) Unscheduled() (items []JobHandle, err os.Error) {
	items, err = this.FindJobs(bson.M{"scheduled": false})
	return
}

func (this *MongoJobStore) Active() (items []JobHandle, err os.Error) {
	items, err = this.FindJobs(bson.M{"running": true})
	return
}

func (this *MongoJobStore) Get(jobId string) (item JobPackage, err os.Error) {
	job := GolemJobC{}

	err = this.JobsCollection().Find(bson.M{"id": jobId}).One(&job)
	if err != nil {
		return
	}

	log("Get(%v):%v", jobId, job)

	item = JobPackage{Handle: job.ToJobHandle()}
	return
}

func (this *MongoJobStore) Update(jobId string, status JobStatus) (err os.Error) {
	now := time.Time{}

	modifierMap := make(map[string]interface{})
	modifierMap["scheduled"] = true // TODO : More granular
	modifierMap["running"] = status.Running
	modifierMap["taskerrored"] = status.ErroredTasks
	modifierMap["taskfinished"] = status.FinishedTasks
	modifierMap["lastmodified"] = now.String()

	err = this.JobsCollection().Update(bson.M{"id": jobId}, modifierMap)
	return
}

func (this *MongoJobStore) FindJobs(m map[string]interface{}) (items []JobHandle, err os.Error) {
	job := GolemJobC{}
	err = this.JobsCollection().Find(m).For(&job, func() os.Error {
		items = append(items, job.ToJobHandle())
		return nil
	})

	log("Found %v %v jobs:", len(items), m)
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
	tm, err := time.Parse("UnixDate", stringTime)
	if err != nil {
		log("ParseTime(%v):%v", stringTime, err)
	}
	return
}
