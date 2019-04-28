package main

import (
	"encoding/json"
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

func (s *Server) handleJobs(ctx *Context) {
	r := ctx.Request
	w := ctx.Response
	db := s.db

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
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		var jobParams common.JobSubmitParam
		err = json.Unmarshal(b, &jobParams)
		if err != nil {
			log.Fatal(err)
		}

		typVal := jobParams.Type
		var workerCount int
		db.Table("workers").Count(&workerCount)
		if typVal == common.IsAliveJob {
			// unmarshalling interface
			var isAliveParam common.JobIsAliveParam
			err = json.Unmarshal([]byte(jobParams.Data), &isAliveParam)
			if err != nil {
				log.Fatal(err)
			}
			ipRanges := common.SubnetSplit(isAliveParam.IpBlock, workerCount)
			for _, ipRange := range ipRanges {
				taskParamData := common.IsAliveParam{ipRange}
				buf, e := json.Marshal(taskParamData)
				if e != nil {
					log.Fatal(e)
				}

				tasks = append(tasks, Task{State: common.Queued, Params: buf})
				//db.Create(&task)
			}

		} else if typVal == common.PortScanJob {
			var portScanParam common.JobPortScanParam
			err = json.Unmarshal([]byte(jobParams.Data), &portScanParam)
			if err != nil {
				log.Fatal(err)
			}
			rangeLength := (portScanParam.EndPort - portScanParam.StartPort) + 1
			quotientWork := (rangeLength / uint16(workerCount)) - 1
			remainderWork := rangeLength % uint16(workerCount)
			currStart := portScanParam.StartPort
			var currEnd uint16
			for i := 0; i < workerCount; i++ {
				currEnd = currStart + quotientWork
				if remainderWork > 0 {
					currEnd += 1
					remainderWork = remainderWork - 1
				}
				taskRange := common.PortRange{currStart, currEnd}
				taskParamData := common.PortScanParam{portScanParam.Type, portScanParam.Ip, taskRange}

				buf, e := json.Marshal(taskParamData)
				if e != nil {
					log.Fatal(e)
				}

				currStart = currEnd + 1

				tasks = append(tasks, Task{State: common.Queued, Params: buf})
				//db.Create(&task)
			}
		}
		job := Job{Params: b, Tasks: tasks}
		db.Create(&job)
		return
	}

}

func (s *Server) handleJobID(ctx *Context) {
	r := ctx.Request
	param := ctx.Params
	w := ctx.Response
	db := s.db

	var id int
	id, err := strconv.Atoi(param[0])
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("id is %d\n", id)

	switch r.Method {
	case "GET":
		rows, err := db.Table("jobs").Where("Id = ?", id).Rows()
		if err != nil {
			log.Fatal(err)
		}
		for rows.Next() {
			var jobs Job
			rows.Scan(&jobs.Id, &jobs.Params)
			io.WriteString(w, "JobId: "+strconv.Itoa(jobs.Id)+" param:"+string(jobs.Params)+"\n")
		}
		return
	case "DELETE":
		fmt.Println("Delete\n")
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
