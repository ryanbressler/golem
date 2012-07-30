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
	"errors"
	"labix.org/v1/mgo"
	"labix.org/v1/mgo/bson"
	"time"
)

const (
	JOBS          = "jobs"
	TASKS         = "tasks"
	CLUSTER_STATS = "cluster_stats"
)

func NewMongoJobStore(dbhost string, dbstore string) *MongoJobStore {
	session, err := mgo.Dial(dbhost)
	if err != nil {
		logger.Fatal(err)
	}

	session.SetMode(mgo.Strong, true) // [Safe, Monotonic, Strong] Strong syncs on inserts/updates
	db := *session.DB(dbstore)
	return &MongoJobStore{db}
}

type MongoJobStore struct {
	Database mgo.Database
}

func (this *MongoJobStore) Create(item JobDetails, tasks []Task) (err error) {
	logger.Debug("Create(%v)", item)
	jobsCollection := this.Database.C(JOBS)

	if err = jobsCollection.Insert(item); err != nil {
		logger.Warn(err)
		return
	}

	tasksCollection := this.Database.C(TASKS)
	return tasksCollection.Insert(TaskHolder{item.JobId, tasks})
}

func (this *MongoJobStore) All() ([]JobDetails, error) {
	return this.FindJobs(bson.M{})
}

func (this *MongoJobStore) Unscheduled() ([]JobDetails, error) {
	return this.FindJobs(bson.M{"state": NEW})
}

func (this *MongoJobStore) CountPending() (pending int, err error) {
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

func (this *MongoJobStore) CountActive() (int, error) {
	return this.CountJobs(bson.M{"state": RUNNING})
}

func (this *MongoJobStore) Get(jobId string) (item JobDetails, err error) {
	logger.Debug("Get(%v)", jobId)

	jobsCollection := this.Database.C(JOBS)
	err = jobsCollection.Find(bson.M{"jobid": jobId}).One(&item)
	return
}

func (this *MongoJobStore) Tasks(jobId string) (tasks []Task, err error) {
	tasksCollection := this.Database.C(TASKS)

	item := TaskHolder{}
	if err = tasksCollection.Find(bson.M{"jobid": jobId}).One(&item); err != nil {
		logger.Warn(err)
		return
	}

	tasks = item.Tasks
	return
}

func (this *MongoJobStore) Update(item JobDetails) (err error) {
	logger.Debug("Update(%v)", item)
	if item.JobId == "" {
		return errors.New("No Job Id Found")
	}

	jobsCollection := this.Database.C(JOBS)

	existing := JobDetails{}
	if err = jobsCollection.Find(bson.M{"jobid": item.JobId}).One(&existing); err != nil {
		logger.Warn(err)
		return
	}

	existing.LastModified = time.Now().String()
	existing.Progress.Finished = item.Progress.Finished
	existing.Progress.Errored = item.Progress.Errored
	existing.State = item.State
	existing.Status = item.Status

	return jobsCollection.Update(bson.M{"jobid": item.JobId}, existing)
}

func (this *MongoJobStore) FindJobs(m map[string]interface{}) (items []JobDetails, err error) {
	logger.Debug("FindJobs(%v)", m)

	jobsCollection := this.Database.C(JOBS)
	iter := jobsCollection.Find(m).Iter()

	for {
		jd := JobDetails{}
		if !iter.Next(&jd) {
			logger.Warn(iter.Err())
			break
		}
		items = append(items, jd)
	}
	return
}

func (this *MongoJobStore) CountJobs(m map[string]interface{}) (int, error) {
	logger.Debug("CountJobs(%v)", m)

	jobsCollection := this.Database.C(JOBS)
	return jobsCollection.Find(m).Count()
}

func (this *MongoJobStore) SnapshotCluster(snapshot ClusterStat) error {
	collection := this.Database.C(CLUSTER_STATS)
	return collection.Insert(snapshot)
}

func (this *MongoJobStore) ClusterStats(numberOfSecondsSince int64) (items []ClusterStat, err error) {
	logger.Debug("ClusterStats(%d)", numberOfSecondsSince)

	collection := this.Database.C(CLUSTER_STATS)

	m := bson.M{}
	if numberOfSecondsSince > 0 {
		timeSince := time.Now().Sub(time.Unix(numberOfSecondsSince,0))
		m = bson.M{"snapshotat": bson.M{"$gt": timeSince}}
	}

	iter := collection.Find(m).Iter()

	for {
		cs := ClusterStat{}
		if !iter.Next(&cs) {
			logger.Warn(iter.Err())
			break
		}
		items = append(items, cs)
	}

	logger.Debug("ClusterStats(%d):%d", numberOfSecondsSince, len(items))
	return
}
