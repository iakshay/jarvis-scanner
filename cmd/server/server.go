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
	"sync"
	"time"
)

type TaskState int

const (
	Queued     TaskState = 0
	InProgress TaskState = 1
	Complete   TaskState = 2
)

type Task struct {
	Id          int
	JobId       int
	State       TaskState
	Params      string
	WorkerId    int
	Worker      Worker `gorm:"foreignkey:TaskId; association_foreignkey:Id"`
	Result      string
	CreatedAt   *time.Time
	CompletedAt *time.Time
}

type Job struct {
	Id     int
	Params string
	Tasks  []Task `gorm:"foreignkey:JobId;association_foreignkey:Id"`
}

type Worker struct {
	Id        int
	TaskId    int
	Name      string
	Address   string
	UpdatedAt *time.Time
	CreatedAt *time.Time
}

type Server struct {
	db *gorm.DB

	connections map[int]*rpc.Client
}

func (server *Server) RegisterWorker(args *common.RegisterWorkerArgs, reply *common.RegisterWorkerReply) error {
	fmt.Println("Register worker", args)

	// init connection to worker
	client, err := rpc.DialHTTP("tcp", args.Address)
	if err != nil {
		log.Println("dialing:", err)
	}

	// insert entry in workers table
	worker := &Worker{Name: args.Name, Address: args.Address}
	if err := server.db.Create(worker).Error; err != nil {
		log.Printf("Error creating worker %s\n", args.Name)
	}

	log.Println("created worker %d\n", worker.Id)
	server.connections[worker.Id] = client

	reply.WorkerId = worker.Id
	return nil
}

func (server *Server) CompleteTask(args *common.CompleteTaskArgs, reply *common.CompleteTaskReply) error {
	fmt.Println("Complete task", args)

	// insert entry in reports table
	if err := server.db.Table("tasks").Where("Id = ?", args.TaskId).Update("result", args.Result).Error; err != nil {
		log.Printf("Error adding resort for TaskId %d\n", args.TaskId)
	}
	return nil
}

func (server *Server) Heartbeat(args *common.HeartbeatArgs, reply *common.HeartbeatReply) error {
	fmt.Println("Send heartbeat", args)

	// update worker hearbeat
	if err := server.db.Table("workers").Where("Id = ?", args.WorkerId).Update("updated_at", time.Now()).Error; err != nil {
		log.Printf("Error updating hearbeat for worker %d\n", args.WorkerId)
	}
	return nil
}

func (server *Server) startTask() {

	for id, client := range server.connections {
		fmt.Printf("sending task to worker id: %d \n", id)
		args := &common.SendTaskArgs{Params: "Simple Task", TaskId: 1}
		var reply common.SendTaskReply
		client.Call("Worker.SendTask", args, reply)
		break
	}
}
func main() {
	fmt.Println("starting server")
	var wg sync.WaitGroup
	wg.Add(1)
	// setup database
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

	server := new(Server)
	server.db = db
	server.connections = make(map[int]*rpc.Client)

	// start rpc server
	rpc.Register(server)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", "localhost:8080")
	if e != nil {
		log.Fatal("listen error:", e)
	}

	go http.Serve(l, nil)

	time.Sleep(3 * time.Second)
	// start dummy task on one worker
	server.startTask()
	wg.Wait()
}

/*func dbMain() {

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
}*/
