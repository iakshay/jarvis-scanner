package main

import (
	"fmt"
	common "github.com/iakshay/jarvis-scanner"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

type TaskState int

const (
	Queued     TaskState = 0
	InProgress TaskState = 1
	Complete   TaskState = 2
)

type Task struct {
	gorm.Model
	JobId    int
	State    TaskState
	Params   string
	WorkerId int
	Worker   Worker `gorm:"foreignkey:TaskId; association_foreignkey:Id"`
	//CreationTime *time.Time
	//CompletionTime *time.Time
}

type Job struct {
	gorm.Model
	//JobId int `gorm:"primary_key;auto_increment"`
	//State TaskState
	Params       string
	CreationTime *time.Time
	Tasks        []Task `gorm:"foreignkey:JobId;association_foreignkey:Id"`
}

type Worker struct {
	gorm.Model
	TaskId           int
	Name             string
	Ip               string
	Port             int
	HeartbeatTime    *time.Time
	RegistrationTime *time.Time
}

type Report struct {
	gorm.Model
	TaskId int
	Result string
}

type Server struct {
}

func (server *Server) RegisterWorker(args *common.RegisterWorkerArgs, reply *common.RegisterWorkerReply) error {
	fmt.Println("Register worker", args)

	return nil
}

func (server *Server) CompleteTask(args *common.CompleteTaskArgs, reply *common.CompleteTaskReply) error {
	fmt.Println("Complete task", args)

	return nil
}

func (server *Server) Heartbeat(args *common.HeartbeatArgs, reply *common.HeartbeatReply) error {
	fmt.Println("Send heartbeat", args)

	return nil
}

func main() {
	fmt.Println("starting server")
	server := new(Server)
	rpc.Register(server)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", "localhost:8080")
	if e != nil {
		log.Fatal("listen error:", e)
	}

	http.Serve(l, nil)
}

func dbMain() {
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()
	db.LogMode(true)
	// Migrate the schema
	db.AutoMigrate(&Job{})
	db.AutoMigrate(&Task{})
	db.AutoMigrate(&Worker{})

	db.Create(&Worker{Name: "Worker1", Ip: "100.0.2.4", Port: 80})
	// Create
	var worker Worker
	db.First(&worker, 0)
	for i := 0; i < 10; i++ {
		db.Create(&Job{
			//State: Queued,
			Params: fmt.Sprintf("FooBar %d", i),
			Tasks: []Task{
				{Params: "Task1", Worker: worker},
				// {Params: "Task2", Worker: 1},
			},
		})
	}

	// List jobs
	var jobs []Job
	db.Find(&jobs)
	fmt.Println(jobs)

	// Read
	var job1 Job
	db.First(&job1, 1) // find product with id 1
	fmt.Println(job1)

	// Get Tasks
	var job2 Job
	db.Preload("Tasks").First(&job2, 1)
	fmt.Println(job2)

	// check job completed
	//var tasks []Task
	//var job3 Job
	var count int
	db.Table("tasks").Where("state != ? AND job_id = ?", Complete, 1).Count(&count)
	//db.First(&job3, 1).Related(&tasks).Where("state != ? AND job_id = ?", Complete, 1).Count(&count);
	fmt.Println(count)

	// Delete - delete product
	//db.Delete(&product)

	db.Table("workers").Where("name = ?", "Worker1").Update("heartbeat_time", time.Now())
}
