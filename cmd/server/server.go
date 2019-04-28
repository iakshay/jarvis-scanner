package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	common "github.com/iakshay/jarvis-scanner"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"regexp"
	"strconv"
	"sync"
	"time"
)

const NotAllocatedWorkerId int = -1

type Task struct {
	Id          int
	JobId       int
	State       common.TaskState
	Params      []byte // Allows for UnMarshalling to struct objects, as needed
	WorkerId    int
	Worker      Worker `gorm:"foreignkey:TaskId; association_foreignkey:Id"`
	Result      string
	CreatedAt   *time.Time
	CompletedAt *time.Time
}

type Job struct {
	Id     int
	Params []byte
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

	Routes []Route
}

type RpcService struct {
	db          *gorm.DB
	connections map[int]*rpc.Client
}

func (service *RpcService) RegisterWorker(args *common.RegisterWorkerArgs, reply *common.RegisterWorkerReply) error {
	fmt.Println("Register worker", args)

	// init connection to worker
	client, err := rpc.DialHTTP("tcp", args.Address)
	if err != nil {
		log.Println("dialing:", err)
	}

	// insert entry in workers table
	worker := &Worker{Name: args.Name, Address: args.Address}
	if err := service.db.Create(worker).Error; err != nil {
		log.Printf("Error creating worker %s\n", args.Name)
	}

	log.Println("created worker %d\n", worker.Id)
	service.connections[worker.Id] = client

	reply.WorkerId = worker.Id
	return nil
}

func (service *RpcService) CompleteTask(args *common.CompleteTaskArgs, reply *common.CompleteTaskReply) error {
	fmt.Println("Complete task", args)

	// insert entry in reports table
	if err := service.db.Table("tasks").Where("Id = ?", args.TaskId).Update("result", args.Result).Error; err != nil {
		log.Printf("Error adding resort for TaskId %d\n", args.TaskId)
	}
	return nil
}

func (service *RpcService) Heartbeat(args *common.HeartbeatArgs, reply *common.HeartbeatReply) error {
	fmt.Println("Send heartbeat", args)

	// update worker hearbeat
	if err := service.db.Table("workers").Where("Id = ?", args.WorkerId).Update("updated_at", time.Now()).Error; err != nil {
		log.Printf("Error updating hearbeat for worker %d\n", args.WorkerId)
	}
	return nil
}

/*func (server *Server) startTask() {

	for id, client := range server.connections {
		fmt.Printf("sending task to worker id: %d \n", id)
		args := &common.SendTaskArgs{TaskData: common.IsAliveParam{}, TaskType: common.IsAliveTask, TaskId: 1}
		var reply common.SendTaskReply
		client.Call("Worker.SendTask", args, reply)
		break
	}
}*/

/*
func (server *Server) Schedule() {
	db := server.db
	connections := server.connections

	for {
		workerAvails := make(map[int]common.WorkerState)
		var workers []Worker
		db.Find(&workers)
		numWorkers := 0
		for worker := range workers {
			workerAvails[worker.Id] = common.Undetermined
			numWorkers += 1
		}
		freeWorkers := make([Worker], numWorkers)
		freeWorkerIndex := 0

		queuedTasks, inProgressTasks, completeTasks := make([Task], 0, numWorkers), make([Task], 0, numWorkers), make([Task], 0, numWorkers)
		queuedCount, inProgressCount, completeCount := 0, 0, 0
		var tasks []Task
		db.Find(&tasks)
		for task := range tasks {
			if task.State == common.Queued {
				queuedTasks[queuedCount] = task
				queuedCount += 1
			} else if task.State == common.InProgress {
				inProgressTasks[inProgressCount] = task
				inProgressCount += 1
			} else {
				completeTasks[completeCount] = task
				completeCount += 1
			}
		}

		for task := range completeTasks {
			workerAvails[task.Worker.Id] = common.Available
			freeWorkers[freeWorkerIndex] = task.Worker
			freeWorkerIndex += 1
		}

		for task := range inProgressTasks {
			workerAvails[task.Worker.Id] = common.Unavailable
		}

		for worker := range workerAvails {
			if workerAvails[worker] == common.Undetermined {
				freeWorkers[freeWorkerIndex]] = worker
				freeWorkerIndex += 1
			}
		}

		index := 0

		for i := 0; i < 2; i++ {
			var arr []Task
			if i == 0 {
				arr = queuedTasks
			} else {
				arr = inProgressTasks
			}

			for (index < numWorkers) && (arr[index] != nil) && (freeWorkers[index] != nil) {
				if (time.Now().Sub(freeWorkers[index].UpdatedAt).Seconds()) <= common.LifeCycle {

				}

				index += 1
			}
		}

		db.Where("updated_at <", Time.
		time.Sleep(1.5 * time.Second)
	}

}
*/

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
	Response http.ResponseWriter
	Request  *http.Request
	Server   *Server
	Params   []string
}

func (c *Context) Text(code int, body string) {
	c.Response.Header().Set("Content-Type", "application/json")
	c.Response.WriteHeader(code)

	payload := struct {
		Message string
	}{body}
	json.NewEncoder(c.Response).Encode(payload)
}

func (c *Context) Json(code int, body interface{}) {
	c.Response.Header().Set("Content-Type", "application/json")
	c.Response.WriteHeader(code)

	json.NewEncoder(c.Response).Encode(body)
}

func (c *Context) Error(code int, err error) {
	log.Println(code, err)
	c.Text(code, err.Error())
}

func (s *Server) handleJobs(ctx *Context) {
	r := ctx.Request
	w := ctx.Response
	db := s.db

	log.Printf("%s /jobs", r.Method)

	if r.URL.Path != "/jobs/" {
		w := ctx.Response
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	var tasks []Task
	switch r.Method {
	//TODO: Just return each job's ID, params; for specific job's details, use its ID returned form this function
	case "GET":
		rows, err := db.Table("jobs").Rows()
		if err != nil {
			log.Fatal(err)
		}
		for rows.Next() {
			var job Job
			rows.Scan(&job.Id, &job.Params)
			io.WriteString(w, "JobId: "+strconv.Itoa(job.Id)+" param:"+string(job.Params)+"\n")
		}
	case "POST":
		contentType := r.Header.Get("Content-type")

		if contentType != "application/json" {
			ctx.Error(http.StatusBadRequest, errors.New("invalid content type"))
			return
		}

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			ctx.Error(http.StatusBadRequest, err)
			return
		}

		var jobParams common.JobSubmitParam
		err = json.Unmarshal(b, &jobParams)
		if err != nil {
			ctx.Error(http.StatusBadRequest, err)
			return
		}

		jobType := jobParams.Type
		var workerCount int
		db.Table("workers").Count(&workerCount)
		log.Printf("JobType: %v WorkerCount:%d", jobType, workerCount)
		if jobType == common.IsAliveJob {
			// unmarshalling interface
			var isAliveParam common.JobIsAliveParam
			err = json.Unmarshal([]byte(jobParams.Data), &isAliveParam)
			if err != nil {
				ctx.Error(http.StatusBadRequest, err)
				return
			}
			// validates ip or subnet
			ipRanges, err := common.SubnetSplit(isAliveParam.IpBlock, workerCount)
			if err != nil {
				ctx.Error(http.StatusBadRequest, err)
				return
			}
			for _, ipRange := range ipRanges {
				taskParamData := common.IsAliveParam{ipRange}
				buf, err := json.Marshal(taskParamData)
				if err != nil {
					log.Fatal(err)
					return
				}

				tasks = append(tasks, Task{WorkerId: NotAllocatedWorkerId, State: common.Queued, Params: buf})
			}

		} else if jobType == common.PortScanJob {
			var portScanParam common.JobPortScanParam
			err = json.Unmarshal([]byte(jobParams.Data), &portScanParam)
			if err != nil {
				ctx.Error(http.StatusBadRequest, err)
				return
			}
			log.Printf("Request: %v", portScanParam)
			err := portScanParam.Validate()
			if err != nil {
				ctx.Error(http.StatusBadRequest, err)
				return
			}

			ip := net.ParseIP(portScanParam.Ip).To4()
			log.Println(workerCount, common.PortRangeSplit(portScanParam.PortRange, workerCount))
			for _, portRange := range common.PortRangeSplit(portScanParam.PortRange, workerCount) {
				taskParamData := common.PortScanParam{portScanParam.Type, ip, portRange}

				buf, e := json.Marshal(taskParamData)
				if e != nil {
					log.Fatal(e)
				}

				tasks = append(tasks, Task{WorkerId: NotAllocatedWorkerId, State: common.Queued, Params: buf})
			}
		}
		job := Job{Params: b, Tasks: tasks}
		db.Create(&job)
		ctx.Text(http.StatusOK, "Successfully submitted job")
		return
	}

}

func (s *Server) handleJobID(ctx *Context) {
	r := ctx.Request
	params := ctx.Params
	db := s.db

	var id int
	id, err := strconv.Atoi(params[0])
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("%s /jobs/%s JobId: %d", r.Method, params[0], id)

	switch r.Method {
	case "GET":
		rows, err := db.Table("jobs").Where("Id = ?", id).Rows()
		if err != nil {
			log.Fatal(err)
		}
		for rows.Next() {
			var jobs Job
			rows.Scan(&jobs.Id, &jobs.Params)
			io.WriteString(ctx.Response, "JobId: "+strconv.Itoa(jobs.Id)+" param:"+string(jobs.Params)+"\n")
		}
		return
	case "DELETE":
		if db.Delete(&Job{}, "id = ?", id).Error != nil {
			ctx.Text(http.StatusBadRequest, "Failed to delete job")
		} else {
			ctx.Text(http.StatusOK, "Failed to delete job")
		}

		return
	}
}

func main() {
	var serverAddr string
	var rpcAddr string
	var dbPath string
	flag.StringVar(&serverAddr, "serverAddr", "localhost:8080", "address of http service")
	flag.StringVar(&rpcAddr, "rpcAddr", "localhost:8081", "address of rpc service")
	flag.StringVar(&dbPath, "db", "test.db", "database path")
	flag.Parse()
	fmt.Println("starting server")
	var wg sync.WaitGroup
	// setup database
	db, err := gorm.Open("sqlite3", dbPath)
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

	//Star custom Mux, for dynamic routing from client-server interactions
	server.Handle("/jobs/([0-9]+)$", func(ctx *Context) {
		server.handleJobID(ctx)
	})

	server.Handle("/jobs/", func(ctx *Context) {
		server.handleJobs(ctx)
	})

	wg.Add(1)
	go http.ListenAndServe(serverAddr, server)

	// start rpc server
	service := new(RpcService)
	service.db = db
	service.connections = make(map[int]*rpc.Client)

	rpc.RegisterName("Server", service)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", rpcAddr)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	wg.Add(1)
	go http.Serve(l, nil)

	// start thread for Scheduler aspect of Server
	//	go server.Schedule()

	wg.Wait()
}
