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

const (
	JOBS          = "jobs"
	TASKS         = "tasks"
	CLUSTER_STATS = "cluster_stats"
)

func NewMongoJobStore(dbhost string, dbstore string) *MongoJobStore {
	session, err := mgo.Mongo(dbhost)
	if err != nil {
		logger.Fatal(err)
	}

	session.SetMode(mgo.Strong, true) // [Safe, Monotonic, Strong] Strong syncs on inserts/updates
	db := session.DB(dbstore)
	return &MongoJobStore{db}
}

type MongoJobStore struct {
	Database mgo.Database
}

func (this *MongoJobStore) Create(item JobDetails, tasks []Task) (err os.Error) {
	logger.Debug("Create(%v)", item)
	jobsCollection := this.Database.C(JOBS)

	if err = jobsCollection.Insert(item); err != nil {
		logger.Warn(err)
		return
	}

	tasksCollection := this.Database.C(TASKS)
	return tasksCollection.Insert(TaskHolder{item.JobId, tasks})
}

func (this *MongoJobStore) All() ([]JobDetails, os.Error) {
	return this.FindJobs(bson.M{})
}

func (this *MongoJobStore) Unscheduled() ([]JobDetails, os.Error) {
	return this.FindJobs(bson.M{"state": NEW})
}

func (this *MongoJobStore) CountPending() (pending int, err os.Error) {
	var newOnes int
	var scheduledOnes int
	if newOnes, err = this.CountJobs(bson.M{"state": NEW}); err != nil {
		return
	}
	if scheduledOnes, err = this.CountJobs(bson.M{"state": SCHEDULED}); err != nil {
		return
	}
	pending = newOnes + scheduledOnes
	return
}

func (this *MongoJobStore) CountActive() (int, os.Error) {
	return this.CountJobs(bson.M{"state": RUNNING})
}

func (this *MongoJobStore) Get(jobId string) (item JobDetails, err os.Error) {
	logger.Debug("Get(%v)", jobId)

	jobsCollection := this.Database.C(JOBS)
	err = jobsCollection.Find(bson.M{"jobid": jobId}).One(&item)
	return
}

func (this *MongoJobStore) Tasks(jobId string) (tasks []Task, err os.Error) {
	tasksCollection := this.Database.C(TASKS)

	item := TaskHolder{}
	if err = tasksCollection.Find(bson.M{"jobid": jobId}).One(&item); err != nil {
		logger.Warn(err)
		return
	}

	tasks = item.Tasks
	return
}

func (this *MongoJobStore) Update(item JobDetails) (err os.Error) {
	logger.Debug("Update(%v)", item)
	if item.JobId == "" {
		return os.NewError("No Job Id Found")
	}

	jobsCollection := this.Database.C(JOBS)

	existing := JobDetails{}
	if err = jobsCollection.Find(bson.M{"jobid": item.JobId}).One(&existing); err != nil {
		logger.Warn(err)
		return
	}

	existing.LastModified = time.LocalTime().String()
	existing.Progress.Finished = item.Progress.Finished
	existing.Progress.Errored = item.Progress.Errored
	existing.State = item.State
	existing.Status = item.Status

	return jobsCollection.Update(bson.M{"jobid": item.JobId}, existing)
}

func (this *MongoJobStore) FindJobs(m map[string]interface{}) (items []JobDetails, err os.Error) {
	logger.Debug("FindJobs(%v)", m)

	jobsCollection := this.Database.C(JOBS)
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

func (this *MongoJobStore) CountJobs(m map[string]interface{}) (int, os.Error) {
	logger.Debug("CountJobs(%v)", m)

	jobsCollection := this.Database.C(JOBS)
	return jobsCollection.Find(m).Count()
}

func (this *MongoJobStore) SnapshotCluster(snapshot ClusterStat) os.Error {
	collection := this.Database.C(CLUSTER_STATS)
	return collection.Insert(snapshot)
}

func (this *MongoJobStore) ClusterStats(numberOfSecondsSince int64) (items []ClusterStat, err os.Error) {
	logger.Debug("ClusterStats(%d)", numberOfSecondsSince)

	collection := this.Database.C(CLUSTER_STATS)

	m := bson.M{}
	if numberOfSecondsSince > 0 {
		timeSince := time.Seconds() - numberOfSecondsSince
		m = bson.M{"snapshotat": bson.M{"$gt": timeSince}}
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

	logger.Debug("ClusterStats(%d):%d", numberOfSecondsSince, len(items))
	return
}
