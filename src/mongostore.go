package main

import "fmt"
import "launchpad.net/mgo"
import "time"
import "launchpad.net/gobson/bson"
import "os"
import "strconv"
import "rand"

const mongoServer = "remus.systemsbiology.net"
const golemDb = "golemstore"
const jobCollection = "golemjobs" //golemjobs contain {1 .. n} tasks
const taskCollection = "taskjobs" 
const NSamples = 1

const(
	STATUS_RUNNING = "RUNNING"
	STATUS_UNSCHEDULED = "UNSCHEDULED"
	STATUS_ERROR = "ERROR"
	STATUS_KILLED = "KILLED"
	STATUS_COMPLETED = "COMPLETED"
)

type JobStore interface {
        Create(item JobPackage) (err os.Error)
        All() (items []JobHandle, err os.Error)
        Active() (items []JobHandle, err os.Error)
        Unscheduled() (items []JobHandle, err os.Error)
        Get(jobId string) (item JobHandle, err os.Error)
        Update(jobId string, status JobStatus) (err os.Error)
}
type JobHandle struct {
        JobId        string
        Owner        string
	Type	     string	
        FirstCreated string
        LastModified string
        Status       JobStatus
}
type JobStatus struct {
        TotalTasks    int
        FinishedTaskCount string
        ErroredTaskCount  string
	ErroredTasks	string	
        Running bool
	Status	string
}
type JobPackage struct {
        Handle JobHandle
        Tasks  []Task
}
type Task struct {
        Count     int
        Commands []string
}

type GolemJobC struct {
        Type string
        Id string
	Owner string
	TimeCreated string
	LastModified string
	TaskCount int
	TasksFinished string
	TaskErroredCount string
	TasksErrored string
	Running bool
	Status string	
}
type GolemTask struct {
	Name string
	JobId string
	Id string
}
type GolemType struct {
	Name string
}
type MongoJobStore struct {
        jobsById map[string]JobPackage
	session *mgo.Session
}
func NewMongoStore() *MongoJobStore{
	ms := new(MongoJobStore)
	initMongoSession()
	//ms.session = session
	return ms
}

func (s MongoJobStore) Create(item JobPackage) (err os.Error) {
        s.jobsById[item.Handle.JobId] = item
	fmt.Println("MongoJobStore Create() item", item)
	goljob_c := s.session.DB(golemDb).C(jobCollection)
	handle := item.Handle
        golemJob := GolemJobC{handle.Type, handle.JobId, handle.Owner, handle.FirstCreated, handle.LastModified, handle.Status.TotalTasks, handle.Status.FinishedTaskCount, handle.Status.ErroredTaskCount, handle.Status.ErroredTasks, handle.Status.Running, STATUS_UNSCHEDULED}
        insertGolemJob(golemJob, goljob_c)
        return
}

func (s MongoJobStore) All() (items []JobHandle, err os.Error) {
        /*for _, item := range s.jobsById {
                items = append(items, item.Handle)
        }*/
	goljob_c := s.session.DB(golemDb).C(jobCollection)
	results, err := goljob_c.Find(bson.M{}).Iter()
	fjobs := 0
        for {
                job := GolemJobC{}
                fjobs = fjobs + 1
                err = results.Next(&job)
		handle := JobHandle{job.Id, job.Owner, job.Type, job.TimeCreated, job.LastModified, 
			JobStatus{job.TaskCount,job.TasksFinished,job.TaskErroredCount, job.TasksErrored, job.Running, STATUS_UNSCHEDULED} }
		items = append(items, handle)
                if err != nil {
                        break
                }
        }
	fmt.Println("Found %v total jobs in golem collection:", strconv.Itoa(fjobs))
        return
}

func (s MongoJobStore) Unscheduled() (items []JobHandle, err os.Error) {
        /*for _, item := range s.jobsById {
                items = append(items, item.Handle)
        }*/
        goljob_c := s.session.DB(golemDb).C(jobCollection)
        results, err := goljob_c.Find(bson.M{"status":STATUS_UNSCHEDULED}).Iter()
        fjobs := 0
        for {
                job := GolemJobC{}
                fjobs = fjobs + 1
                err = results.Next(&job)
                handle := JobHandle{job.Id, job.Owner, job.Type, job.TimeCreated, job.LastModified,
                        JobStatus{job.TaskCount,job.TasksFinished,job.TaskErroredCount, job.TasksErrored, job.Running, STATUS_UNSCHEDULED} }
                items = append(items, handle)
                if err != nil {
                        break
                }
        }
        fmt.Println("Found %v unscheduled jobs:", strconv.Itoa(fjobs))
        return
}

func (s MongoJobStore) Active() (items []JobHandle, err os.Error) {
        /*for _, item := range s.jobsById {
                items = append(items, item.Handle)
        }*/
	return FindJobsByStatus(s, STATUS_RUNNING)
}

func FindJobsByStatus(store MongoJobStore, mystatus string) (items []JobHandle, err os.Error) {
        /*for _, item := range s.jobsById {
                items = append(items, item.Handle)
        }*/
        goljob_c := store.session.DB(golemDb).C(jobCollection)
        results, err := goljob_c.Find(bson.M{"status":mystatus}).Iter()
        fjobs := 0
        for {
                job := GolemJobC{}
                fjobs = fjobs + 1
                err = results.Next(&job)
                handle := JobHandle{job.Id, job.Owner, job.Type, job.TimeCreated, job.LastModified,
                        JobStatus{job.TaskCount,job.TasksFinished,job.TaskErroredCount, job.TasksErrored, job.Running, STATUS_UNSCHEDULED} }
                items = append(items, handle)
                if err != nil {
                        break
                }
        }
        fmt.Println("Found %v unscheduled jobs:", strconv.Itoa(fjobs))
        return
}



func (s MongoJobStore) Get(jobId string) (item JobHandle, err os.Error) {
	goljob_c := s.session.DB(golemDb).C(jobCollection)
	result := GolemJobC{}
	err = goljob_c.Find(bson.M{"id": jobId}).One(&result)
	fmt.Println("My result ", result)
	if (err != nil){
		fmt.Println("Item not found")
		return
	}
	status := JobStatus{result.TaskCount, result.TasksFinished, result.TaskErroredCount, result.TasksErrored, result.Running, result.Status}
	fmt.Println("My result's status ", status)
	return JobHandle{result.Id, result.Owner, result.Type, result.TimeCreated, result.LastModified, status}, err
	//return item, err		
}

/*
func (s MongoJobStore) Active() (items []JobHandle, err os.Error) {
        for _, item := range s.jobsById {
                if item.Handle.Status.Running {
                        items = append(items, item.Handle)
                }
        }
        return
}
*/
/*func (s MongoJobStore) Unscheduled() (items []JobHandle, err os.Error) {
        //    items = make([]JobHandle, 100)
        for _, item := range s.jobsById {
                if item.Handle.Status.Running == false {
                        items = append(items, item.Handle)
                }
        }
        return
}
*/
/*
func (s MongoJobStore) Get(jobId string) (item JobPackage, err os.Error) {
        item, isin := s.jobsById[jobId]
        if isin == false {
                err = os.NewError("item not found")
        }
        return
}
*/

func (s MongoJobStore) Update(jobId string, status JobStatus) (err os.Error) {
        //item, err := s.Get(jobId)
	goljob_c := s.session.DB(golemDb).C(jobCollection)
	modifierMap := map[string]interface{}{"$set": map[string]string{"status":status.Status, "running":strconv.Btoa(status.Running), "taskerrored":status.ErroredTaskCount, "taskfinished":status.FinishedTaskCount, "lastmodified":strconv.Itoa64(time.Seconds())}}
        err = goljob_c.Update(bson.M{"id":jobId},modifierMap)
        //item.Status = status
        return err
}

/*func findAllJobs(gjob_col mgo.Collection) (results *Iter, err os.Error) {
	results, err = gjob_col.Find(bson.M{""}).Iter()
	return results, err
}
*/
func insertGolemJob(gjob GolemJobC, gjob_col mgo.Collection){
	gjob_col.Insert(gjob)
	fmt.Println("Inserted job:", gjob)
}

func insertGolemTask(gtask GolemTask, tjob_col mgo.Collection){
	fmt.Println("Inserting Task:", gtask)
	tjob_col.Insert(gtask)
}

func initMongoSession() (*mgo.Session, os.Error){
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
        //goljob_c := session.DB(golemDb).C(jobCollection)
        //goljob_c.RemoveAll(&GolemType{"rf_ace"})
        var NotFound = os.NewError("Document not found")

        jobStatus := JobStatus{rand.Int(), "0", "0", "", true, STATUS_UNSCHEDULED}
        jobHandle := JobHandle{"myjobid_jlin", "jlin", "rf_ace", strconv.Itoa64(time.Seconds()), strconv.Itoa64(time.Seconds()), jobStatus}
        commands := []string{"ls","ls","ls"}
        task := Task{1, commands}
        mytasks := []Task{task}
        myjobSubmission := JobPackage{jobHandle,mytasks}
        var mymap = map[string] JobPackage {
                "pkgstring_id": myjobSubmission}
        mystore := MongoJobStore{mymap, session}
        //fmt.Println("Testing mystore.Create")
	fmt.Println("Begin testing store.Create", time.LocalTime(), "Seconds", time.Seconds())
        for i := 0; i< NSamples; i++{
                //golemJob := GolemJob{"rf_ace", "id_" + strconv.Itoa(i) + "_" + strconv.Itoa64(time.Nanoseconds()), "jlin", 999, 999, 100, 0, 0, true}
                //insertGolemJob(golemJob, goljob_c)
                jobHandle := JobHandle{"myjobid_" + strconv.Itoa(i), "jlin", "rf_ace", strconv.Itoa64(time.Seconds()), strconv.Itoa64(time.Seconds()), jobStatus}
                myjobSubmission := JobPackage{jobHandle,mytasks}
                mystore.Create(myjobSubmission)
        }
        fmt.Println("Done testing store.Create", time.LocalTime(), "Seconds", time.Seconds())
	fmt.Println("Begin store.All()")
        mystore.All()
	
        fmt.Println("\nBegin Querying, done with ", strconv.Itoa(NSamples) + " inserts ", time.LocalTime(), "Seconds", time.Seconds())
        
	/*result := GolemJob{}
        err = goljob_c.Find(bson.M{"type": "rf_ace"}).One(&result)
        if err != nil && err != NotFound {
                fmt.Println("Error at find rf_ace type")
                panic(err)
        }*/

        /*fmt.Println("Begin Testing Find/Select", time.LocalTime())
        results, err := goljob_c.Find(bson.M{"name": "rf_ace"}).Iter()
        fjobs := 0
        for {
                job := GolemJobC{}
                fjobs = fjobs + 1
                err = results.Next(&job)
                if err != nil {
                        break
                }
                //fmt.Println("Found rf_ace typed jobs:", job.Id, job)
        }
        fmt.Println("Found X jobs in All(...):", strconv.Itoa(fjobs))
        if err != mgo.NotFound {
           panic(err)
        }*/

	jid := "myjobid_jlin_1372544545"
	jid = "myjobid_jlin_1325201247"
	jid = "unique101"
	item, err := mystore.Get(jid)
	if (err != nil && err != NotFound){
		panic(err)
	}
	fmt.Println("Found job item handle:", item)
	fmt.Println("Begin testing of store.Update")
	//modifierMap := map[string]interface{}{"$set": map[string]string{"running":STATUS_RUNNING, "taskerrored":"99", "taskfinished":"1000", "lastmodified":strconv.Itoa64(time.Seconds())}} 	
	//err = goljob_c.Update(bson.M{"id":jid},modifierMap)
	mystore.Update(jid, JobStatus{1000, "0", "0", "", true, STATUS_RUNNING})
	
	//err = goljob_c.Update(bson.M{"id":jid}, bson.M{"$set":{"running":false,"a":"a_value"}})
	if (err != nil && err != NotFound){
                panic(err)
        }
        fmt.Println("End of main", time.LocalTime(), "Seconds", time.Seconds())
}
