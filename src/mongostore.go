package main

import "fmt"
import "launchpad.net/mgo"
import "time"
import "launchpad.net/gobson/bson"
import "os"
import "strconv"
import "rand"
import "goconf.googlecode.com/hg"

//const golemDb = "golemstore"
var golemstore = ""
var jobCollection = ""
var taskCollection = ""
//const jobCollection = "golemjobs" //golemjobs contain {1 .. n} tasks
//const taskCollection = "taskjobs" 
const TestNSamples = 1000

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

func initMongoSession(mgoserver string) (*mgo.Session, os.Error){
        session, err := mgo.Mongo(mgoserver)
        //defer session.Close()
        // Modes are Safe, Monotonic, and Strong, Strong tells the system to sync on inserts/updates
        session.SetMode(mgo.Strong, true)
        return session, err
}

func NewMongoStore(mapid map[string]JobPackage, configpath string) *MongoJobStore{
	config, err := conf.ReadConfigFile(configpath)
	if (err != nil){
		panic(err)
	}
	mstore := new(MongoJobStore)
	dbhost, _ := config.GetString("mgodb", "server")
	golstore, _ := config.GetString("mgodb", "store")
	golcollection, _ := config.GetString("mgodb", "jobcollection")
	goltask, _ := config.GetString("mgodb", "taskcollection")
	golemstore = golstore
	jobCollection = golcollection
	taskCollection = goltask
	session, err := initMongoSession(dbhost)
	if (err != nil){
		return nil
	}
	mstore.jobsById = mapid
	mstore.session = session
	return mstore
}

func (s MongoJobStore) Create(item JobPackage) (err os.Error) {
        s.jobsById[item.Handle.JobId] = item
	fmt.Println("MongoJobStore Create() item", item)
	goljob_c := s.session.DB(golemstore).C(jobCollection)
	handle := item.Handle
        golemJob := GolemJobC{handle.Type, handle.JobId, handle.Owner, handle.FirstCreated, handle.LastModified, handle.Status.TotalTasks, handle.Status.FinishedTaskCount, handle.Status.ErroredTaskCount, handle.Status.ErroredTasks, handle.Status.Running, STATUS_UNSCHEDULED}
        insertGolemJob(golemJob, goljob_c)
        return
}

func (s MongoJobStore) All() (items []JobHandle, err os.Error) {
        /*for _, item := range s.jobsById {
                items = append(items, item.Handle)
        }*/
	goljob_c := s.session.DB(golemstore).C(jobCollection)
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
        return FindJobsByStatus(s, STATUS_UNSCHEDULED)
}

func (s MongoJobStore) Active() (items []JobHandle, err os.Error) {
	return FindJobsByStatus(s, STATUS_RUNNING)
}

func FindJobsByStatus(store MongoJobStore, mystatus string) (items []JobHandle, err os.Error) {
        goljob_c := store.session.DB(golemstore).C(jobCollection)
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
        fmt.Println("Found %v %v jobs:", strconv.Itoa(fjobs), mystatus)
        return
}

func (s MongoJobStore) Get(jobId string) (item JobHandle, err os.Error) {
	goljob_c := s.session.DB(golemstore).C(jobCollection)
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
}

func (s MongoJobStore) Update(jobId string, status JobStatus) (err os.Error) {
        //item, err := s.Get(jobId)
	goljob_c := s.session.DB(golemstore).C(jobCollection)
	modifierMap := map[string]interface{}{"$set": map[string]string{"status":status.Status, "running":strconv.Btoa(status.Running), "taskerrored":status.ErroredTaskCount, "taskfinished":status.FinishedTaskCount, "lastmodified":strconv.Itoa64(time.Seconds())}}
        err = goljob_c.Update(bson.M{"id":jobId},modifierMap)
        return err
}

func insertGolemJob(gjob GolemJobC, gjob_col mgo.Collection){
	gjob_col.Insert(gjob)
	fmt.Println("Inserted job:", gjob)
}

func insertGolemTask(gtask GolemTask, tjob_col mgo.Collection){
	fmt.Println("Inserting Task:", gtask)
	tjob_col.Insert(gtask)
}

func main() {
        fmt.Println("Begin", time.LocalTime(), "Seconds", time.Seconds())
        
	//session, err := initMongoSession() //mgo.Mongo(mongoServer)
        //if err != nil {
        //        panic(err)
        //}
        //defer session.Close()
        // Optional. Modes are Switch the session to a monotonic behavior, place this inside init
        //session.SetMode(mgo.Strong, true)
        
	//goljob_c := session.DB(golemstore).C(jobCollection)
        //goljob_c.RemoveAll(&GolemType{"rf_ace"})
        var NotFound = os.NewError("Document not found")

        jobStatus := JobStatus{rand.Int(), "0", "0", "", true, STATUS_UNSCHEDULED}
        jobHandle := JobHandle{"myjobid_jlin", "jlin", "rf_ace", strconv.Itoa64(time.Seconds()), strconv.Itoa64(time.Seconds()), jobStatus}
        commands := []string{"ls","ls","ls"}
        task := Task{1, commands}
        mytasks := []Task{task}
        myjobSubmission := JobPackage{jobHandle,mytasks}
        var packageMap = map[string] JobPackage {"pkgstring_id": myjobSubmission}        
        
	mystore := NewMongoStore(packageMap, "./templates/golem.config")
	defer mystore.session.Close()

	fmt.Println("Begin testing store.Create", time.LocalTime(), "Seconds", time.Seconds())
        for i := 0; i< TestNSamples; i++{
                jobHandle := JobHandle{"myjobid_" + strconv.Itoa(i) + "_" + strconv.Itoa64(time.Seconds()), "jlin", "rf_ace", strconv.Itoa64(time.Seconds()), strconv.Itoa64(time.Seconds()), jobStatus}
                myjobSubmission := JobPackage{jobHandle,mytasks}
                mystore.Create(myjobSubmission)
        }
        fmt.Println("Done testing store.Create", time.LocalTime(), "Seconds", time.Seconds())
	fmt.Println("Begin store.All()")
        mystore.All()
        fmt.Println("\nBegin Querying, done with ", strconv.Itoa(TestNSamples) + " inserts ", time.LocalTime(), "Seconds", time.Seconds())
        
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
	mystore.Update("myjobid_999", JobStatus{-1, "900", "10", "1,2,3,4,5,6,7,8,9,19", true, STATUS_RUNNING})
	//err = goljob_c.Update(bson.M{"id":jid}, bson.M{"$set":{"running":false,"a":"a_value"}})
	if (err != nil && err != NotFound){
                panic(err)
        }
	fmt.Println("Begin testing of store.Unscheduled()")
	mystore.Unscheduled()
	fmt.Println("Begin testing of store.Active()")
	mystore.Active()
        fmt.Println("End of main", time.LocalTime(), "Seconds", time.Seconds())
}
