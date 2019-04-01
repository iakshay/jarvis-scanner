package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"strconv"
	common "github.com/iakshay/jarvis-scanner"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
        "regexp"
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
	Routes		[]Route
}

type JobType int

const (
isAlive	JobType = 1
portScan JobType = 2
)

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

// https://gist.github.com/reagent/043da4661d2984e9ecb1ccb5343bf438
// From the example under, "Custom Regular Expression-Based Router"

type Handler func(*Context)

type Route struct {
	Pattern *regexp.Regexp
	Handler Handler
}

func (s *Server) Handle(pattern string, handler Handler) {
	re := regexp.MustCompile(pattern)
	route := Route{Pattern: re, Handler: handler}

	s.Routes = append(s.Routes, route)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{Response: w, Request: r, Server: s}

	for _, rt := range s.Routes {
		if matches := rt.Pattern.FindStringSubmatch(r.URL.Path); len(matches) > 0 {
			if len(matches) > 1 {
				ctx.Params = matches[1:]
			}

			rt.Handler(ctx)
			return
		}
	}
}

type Context struct {
	Response	http.ResponseWriter
	Request		*http.Request
	Server		*Server
	Params		[]string
}

func (s *Server) handleJobs(ctx *Context) {
	r := ctx.Request
	w := ctx.Response
	db := s.db

	if r.URL.Path != "/jobs/" {
		w := ctx.Response
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	switch r.Method {
	//TODO: Just return each job's ID, params; for specific job's details, use its ID returned form this function
	case "GET":
		rows, err := db.Table("tasks").Joins("inner join jobs on jobs.id = tasks.job_id").Rows()
		if err != nil {
			log.Fatal(err)
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		jobId := -1
		completed := true
		for rows.Next() {
			task := new(Task)
			if err := rows.Scan(&(task.Id), &(task.JobId), &(task.State), &(task.Params), &(task.WorkerId), &(task.Result), &(task.CreatedAt), &(task.CompletedAt)); err != nil {
				log.Fatal(err)
			}

			if task.JobId != jobId {
				jobId = task.JobId
				if jobId > 0 {
					if completed == true {
						io.WriteString(w, "Job " + strconv.Itoa(task.JobId) + ": complete")
					} else {
						io.WriteString(w, "Job " + strconv.Itoa(task.JobId) + ": incomplete")
					}
					completed = true
					io.WriteString(w, "\n\n")
				}
			}

			if task.State != Complete {
				completed = false
			}

			io.WriteString(w,  task.Params + "\n")
		}
		return
	case "POST":
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		s := string(b)
		spl := strings.Split(s, ", ")
		fullType := strings.Split(spl[1], ":")

		// The code representing the type of scan the client tells us to perform
		// "1" for isAlive, "2" for scanning of ports
		typeVal := strconv.Atoi(fullType[1][1])

		workerCount := db.

		if typVal == isAlive {
			fullIPRange := strings.Split(spl[0], ":")
			length := len(fullIPRange[1])
			ipRangeVal := fullIPRange[1][1:(length - 1)]
			
		} else {

		}
		io.WriteString(w, s)
		return
	}

}

func (s *Server) handleJobID(ctx *Context) {
	r  := ctx.Request

	switch r.Method {
	case "GET":
		return
	case "DELETE":
		return
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

/*	for i:= 0; i < 2; i++ {
		var tasks []Task
		for j:= 0; j < 3; j++ {
			worker := new(Worker)
			task := new(Task)
			params := "Task" + strconv.Itoa(j)
			task.Params = params
			task.Worker = *worker
			task.State = Queued
			db.Create(task)
			db.Create(worker)
			tasks = append(tasks, *task)
		}
                db.Create(&Job{
                        Params: fmt.Sprintf("FooBar %d", i),
                        Tasks: tasks,
                })
        }*/

	server := new(Server)
	server.db = db
	server.connections = make(map[int]*rpc.Client)

	//Star custom Mux, for dynamic routing from client-server interactions
	server.Handle("/jobs/([0-9]+)$", func(ctx *Context) {
		server.handleJobID(ctx)
	})

	server.Handle("/jobs/", func(ctx *Context) {
		server.handleJobs(ctx)
	})

	err_ := http.ListenAndServe("localhost:8080", server)
	if err_ != nil {
		log.Fatalf("Could not start server: %s\n", err_.Error())
	}

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
