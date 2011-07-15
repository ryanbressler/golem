package main

import "fmt"
import "launchpad.net/mgo"
import "time"
import "launchpad.net/gobson/bson"
import "os"
import "strconv"

const mongoServer = "remus.systemsbiology.net"
const golemDb = "golemstore"
const jobCollection = "golemjobs" //golemjobs contain {1 .. n} tasks
const taskCollection = "taskjobs"
const NSamples = 10

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
	Type         string
	FirstCreated int64
	LastModified int64
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
	Count    int
	Commands []string
}

type GolemJob struct {
	Type         string
	Id           string
	Owner        string
	TimeCreated  int64
	LastModified int64
	TaskCount    int
	TaskFinished int
	TaskErrored  int
	Running      bool
}

type GolemTask struct {
	Name  string
	JobId string
	Id    string
}

type GolemType struct {
	Name string
}

type MongoJobStore struct {
	jobsById map[string]JobPackage
	session  *mgo.Session
}

func NewMongoStore() *MongoJobStore {
	ms := new(MongoJobStore)
	initMongoSession()
	//ms.session = session
	return ms
}

func (s MongoJobStore) Create(item JobPackage) (err os.Error) {
	s.jobsById[item.Handle.JobId] = item
	fmt.Println("MongoJobStore Create() item", item)
	/*
		a package has a jobhandle with a list of tasks
	*/
	goljob_c := s.session.DB(golemDb).C(jobCollection)
	handle := item.Handle
	golemJob := GolemJob{handle.Type, handle.JobId, handle.Owner, handle.FirstCreated, handle.LastModified, handle.Status.TotalTasks, handle.Status.FinishedTasks, handle.Status.ErroredTasks, handle.Status.Running}
	insertGolemJob(golemJob, goljob_c)
	return
}

func (s MongoJobStore) All() (items []JobHandle, err os.Error) {
	for _, item := range s.jobsById {
		items = append(items, item.Handle)
	}
	return
}

func (s MongoJobStore) Active() (items []JobHandle, err os.Error) {
	for _, item := range s.jobsById {
		if item.Handle.Status.Running {
			items = append(items, item.Handle)
		}
	}
	return
}

func (s MongoJobStore) Unscheduled() (items []JobHandle, err os.Error) {
	//    items = make([]JobHandle, 100)
	for _, item := range s.jobsById {
		if item.Handle.Status.Running == false {
			items = append(items, item.Handle)
		}
	}
	return
}

func (s MongoJobStore) Get(jobId string) (item JobPackage, err os.Error) {
	item, isin := s.jobsById[jobId]
	if isin == false {
		err = os.NewError("item not found")
	}
	return
}

func (s MongoJobStore) Update(jobId string, status JobStatus) (err os.Error) {
	item, err := s.Get(jobId)
	item.Handle.Status = status
	return
}

func insertGolemJob(gjob GolemJob, gjob_col mgo.Collection) {
	gjob_col.Insert(gjob)
	fmt.Println("Inserted job:", gjob)
}

func insertGolemTask(gtask GolemTask, tjob_col mgo.Collection) {
	fmt.Println("Inserting Task:", gtask)
	tjob_col.Insert(gtask)
}

func initMongoSession() (*mgo.Session, os.Error) {
	session, err := mgo.Mongo(mongoServer)
	return session, err
}

func main() {
	fmt.Println("Begin", time.LocalTime(), "Seconds", time.Seconds())
	session, err := initMongoSession() //mgo.Mongo(mongoServer)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	//fmt.Println("Got session", session)
	// Optional. Switch the session to a monotonic behavior, place this inside init
	session.SetMode(mgo.Strong, true)
	goljob_c := session.DB(golemDb).C(jobCollection)
	//goljob_c.RemoveAll(&GolemType{"rf_ace"})
	/*err = goljob_c.Insert(&GolemJob{"rf_ace", "id001"},
	               &GolemJob{"rf_ace", "id002"},&GolemJob{"rf_ace", "id003"},&GolemJob{"rf_ace", "id004"},
		       &GolemJob{"rf_ace", "id005"})
	*/
	jobStatus := JobStatus{100, 0, 0, true}
	jobHandle := JobHandle{"myjobid", "jlin", "rf-ace", time.Seconds(), time.Seconds(), jobStatus}
	commands := []string{"ls", "ls", "ls"}
	task := Task{100, commands}
	mytasks := []Task{task}
	myjobSubmission := JobPackage{jobHandle, mytasks}
	var mymap = map[string]JobPackage{
		"pkgstring_id": myjobSubmission}
	mystore := MongoJobStore{mymap, session}
	fmt.Println("Testing mystore.Create")
	mystore.Create(myjobSubmission)

	for i := 0; i < NSamples; i++ {
		golemJob := GolemJob{"rf_ace", "id_" + strconv.Itoa(i) + "_" + strconv.Itoa64(time.Nanoseconds()), "jlin", 999, 999, 100, 0, 0, true}
		insertGolemJob(golemJob, goljob_c)
	}
	fmt.Println("Begin Querying, done with ", strconv.Itoa(NSamples)+" inserts ", time.LocalTime(), "Seconds", time.Seconds())
	result := GolemJob{}
	err = goljob_c.Find(bson.M{"name": "rf_ace"}).One(&result)
	if err != nil {
		panic(err)
	}
	fmt.Println("Begin Testing Find/Select", time.LocalTime())
	results, err := goljob_c.Find(bson.M{"name": "rf_ace"}).Iter()
	fjobs := 0
	for {
		job := GolemJob{}
		fjobs = fjobs + 1
		err = results.Next(&job)
		if err != nil {
			break
		}
		//fmt.Println("Found rf_ace typed jobs:", job.Id, job.Name)
	}
	fmt.Println("Found X jobs:", strconv.Itoa(fjobs))
	if err != mgo.NotFound {
		panic(err)
	}

	fmt.Println("End", time.LocalTime(), "Seconds", time.Seconds())
}
