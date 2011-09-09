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
	"launchpad.net/mgo"
	"launchpad.net/gobson/bson"
)

type MongoJobStore struct {
	Host                   string
	Store                  string
	JobsCollection         string
	TasksCollection        string
	ClusterStatsCollection string
}

func (this *MongoJobStore) GetCollection(collectionName string) (c mgo.Collection, err os.Error) {
	logger.Debug("GetCollection(%v)", collectionName)
	session, err := mgo.Mongo(this.Host)
	if err != nil {
		return
	}

	session.SetMode(mgo.Strong, true) // [Safe, Monotonic, Strong] Strong syncs on inserts/updates
	db := session.DB(this.Store)
	c = db.C(collectionName)
	return
}

func (this *MongoJobStore) Create(item JobDetails, tasks []Task) os.Error {
	logger.Debug("Create(%v)", item)
	jobsCollection, err := this.GetCollection(this.JobsCollection)
	if err != nil {
		logger.Warn(err)
		return err
	}

	err = jobsCollection.Insert(item)
	if err != nil {
		logger.Warn(err)
		return err
	}

	tasksCollection, err := this.GetCollection(this.TasksCollection)
	if err != nil {
		logger.Warn(err)
		return err
	}

	return tasksCollection.Insert(TaskHolder{item.JobId, tasks})
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
	logger.Debug("Get(%v)", jobId)
	jobsCollection, err := this.GetCollection(this.JobsCollection)
	if err != nil {
		logger.Warn(err)
		return
	}

	err = jobsCollection.Find(bson.M{"jobid": jobId}).One(&item)
	if err != nil {
		logger.Warn(err)
		return
	}
	return
}

func (this *MongoJobStore) Tasks(jobId string) (tasks []Task, err os.Error) {
	tasksCollection, err := this.GetCollection(this.TasksCollection)
	if err != nil {
		logger.Warn(err)
		return
	}

	item := TaskHolder{}
	err = tasksCollection.Find(bson.M{"jobid": jobId}).One(&item)
	if err != nil {
		logger.Warn(err)
		return
	}

	tasks = item.Tasks
	return
}

func (this *MongoJobStore) Update(item JobDetails) os.Error {
	logger.Debug("Update(%v)", item)
	if item.JobId == "" {
		return os.NewError("No Job Id Found")
	}

	jobsCollection, err := this.GetCollection(this.JobsCollection)
	if err != nil {
		logger.Warn(err)
		return err
	}

	existing := JobDetails{}
	err = jobsCollection.Find(bson.M{"jobid": item.JobId}).One(&existing)
	if err != nil {
		logger.Warn(err)
		return err
	}

	existing.LastModified = time.LocalTime().String()
	existing.Progress.Finished = item.Progress.Finished
	existing.Progress.Errored = item.Progress.Errored
	existing.Running = item.Running
	existing.Scheduled = item.Scheduled

	logger.Debug("Update(%v): %v", item, existing)
	return jobsCollection.Update(bson.M{"jobid": item.JobId}, existing)
}

func (this *MongoJobStore) FindJobs(m map[string]interface{}) (items []JobDetails, err os.Error) {
	logger.Debug("FindJobs(%v)", m)
	jobsCollection, err := this.GetCollection(this.JobsCollection)
	if err != nil {
		logger.Warn(err)
		return
	}

	iter, err := jobsCollection.Find(m).Iter()
	if err != nil {
		logger.Warn(err)
		return
	}

	for {
		jd := JobDetails{}
		if nexterr := iter.Next(&jd); nexterr != nil {
			logger.Warn(nexterr)
			break
		}
		items = append(items, jd)
	}
	return
}

func (this *MongoJobStore) ClusterSnapshot(snapshot ClusterStat) (err os.Error) {
	collection, err := this.GetCollection(this.ClusterStatsCollection)
	if err == nil {
		err = collection.Insert(snapshot)
	}

	if err != nil {
		logger.Warn(err)
	}

	return
}

func (this *MongoJobStore) ClusterStats(numberOfSecondsSince int64) (items []ClusterStat, err os.Error) {
	m := make(map[string]string)
	if numberOfSecondsSince > 0 {
		m["SnapshotAt"] = fmt.Sprintf("$gt:%d", time.Seconds()-numberOfSecondsSince)
	}

	collection, err := this.GetCollection(this.ClusterStatsCollection)
	if err != nil {
		logger.Warn(err)
		return
	}

	iter, err := collection.Find(m).Iter()
	if err != nil {
		logger.Warn(err)
		return
	}

	for {
		cs := ClusterStat{}
		if nexterr := iter.Next(&cs); nexterr != nil {
			logger.Warn(nexterr)
			break
		}
		items = append(items, cs)
	}
	return
}
