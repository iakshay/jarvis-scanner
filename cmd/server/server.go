package main

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	common "github.com/iakshay/jarvis-scanner"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

const NotAllocatedWorkerId int = -1

const NotAllocatedTaskId int = -1

type Task struct {
	Id          int
	JobId       int
	State       common.TaskState
	Type        common.TaskType
	WorkerId    int
	Params      []byte // Allows for Unmarshalling to struct objects, as needed
	Worker      []byte `gorm:"foreignkey:TaskId; association_foreignkey:Id"`
	Result      []byte
	CreatedAt   *time.Time
	CompletedAt *time.Time
}

type Job struct {
	Id     int
	Type   common.JobType
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
	worker := &Worker{Name: args.Name, Address: args.Address, TaskId: NotAllocatedWorkerId}
	if err := service.db.Create(worker).Error; err != nil {
		log.Printf("Error creating worker %s\n", args.Name)
	}

	log.Println("created worker %d\n", worker.Id)
	service.connections[worker.Id] = client

	reply.WorkerId = worker.Id
	return nil
}

func (service *RpcService) CompleteTask(args *common.CompleteTaskArgs, reply *common.CompleteTaskReply) error {
	log.Println("Complete task received")
	buf, err := json.Marshal(args.Result)
	if err != nil {
		log.Fatal(err)
	}

	service.db.Table("workers").Where("task_id = ?", args.TaskId).Update("task_id", NotAllocatedTaskId)

	// insert entry in reports table
	if err := service.db.Table("tasks").Where("Id = ?", args.TaskId).Updates(Task{State: common.Complete, Result: buf}).Error; err != nil {
		log.Printf("Error adding resort for TaskId %s %d\n", err, args.TaskId)
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

func (server *Server) startTask(workerId int, task Task, service *RpcService) {

	fmt.Printf("sending task to worker id: %d \n", workerId)
	client := service.connections[workerId]
	var args common.SendTaskArgs
	if task.Type == common.IsAliveTask {
		var taskData common.IsAliveParam
		if err := json.Unmarshal(task.Params, &taskData); err != nil {
			log.Printf("Error while unmarshalling task data.\n")
		}
		args = common.SendTaskArgs{TaskData: taskData, TaskType: task.Type, TaskId: task.Id}
	} else {
		var taskData common.PortScanParam
		if err := json.Unmarshal(task.Params, &taskData); err != nil {
			log.Printf("Error while unmarshalling task data.\n")
		}
		args = common.SendTaskArgs{TaskData: taskData, TaskType: task.Type, TaskId: task.Id}
	}
	var reply common.SendTaskReply
	if err := client.Call("Worker.SendTask", &args, &reply); err != nil {
		log.Fatal(err)
	}
}

func (server *Server) Schedule(service *RpcService) {
	db := server.db

	for {
		var workerCount int
		db.Table("workers").Count(&workerCount)
		fmt.Printf("total workers: %d\n", workerCount)
		var freeWorkers []Worker
		if result := db.Table("workers").Where("task_id = ?", -1).Find(&freeWorkers); result.Error != nil {
			fmt.Printf("Error returning free workers.")
		}

		fmt.Printf("free workers: %d\n", len(freeWorkers))

		// get active workers
		availWorkers := make([]Worker, len(freeWorkers))
		numAvailWorkers := 0
		for _, worker := range freeWorkers {
			updatedAt := *(worker.UpdatedAt)
			if (time.Now().Sub(updatedAt)) <= common.HeartbeatInterval {
				availWorkers[numAvailWorkers] = worker
				numAvailWorkers += 1
			}
		}

		fmt.Printf("active workers: %d\n", len(availWorkers))

		var queuedTasks []Task
		db.Order("created_at asc").Where("state = ?", common.Queued).Limit(numAvailWorkers).Find(&queuedTasks)
		availIndex := 0
		fmt.Printf("queued tasks: %d\n", len(queuedTasks))
		var inProgressTasks []Task
		db.Order("created_at asc").Where("state = ?", common.InProgress).Find(&inProgressTasks)

		for i := 0; i < 2; i++ {
			var tasks []Task
			if i == 0 {
				tasks = queuedTasks
			} else if availIndex < numAvailWorkers {
				tasks = inProgressTasks
			}

			if tasks != nil {
				for _, task := range tasks {
					if availIndex == numAvailWorkers {
						break
					}

					currWorker := Worker{Id: -1}
					if i == 1 {
						workerBytes := task.Worker
						if err := json.Unmarshal(workerBytes, &currWorker); err != nil {
							log.Fatal(err)
						}
					}

					if currWorker.Id == -1 || (time.Now().Sub(*(currWorker.UpdatedAt)) > common.HeartbeatInterval){
						worker := availWorkers[availIndex]
						db.Table("workers").Where("id = ?", worker.Id).Update("task_id", task.Id)
						db.Table("tasks").Where("id = ?", task.Id).Update("worker_id", worker.Id)
						workerBytes, _ := json.Marshal(&worker)
						db.Table("tasks").Where("id = ?", task.Id).Update("worker", workerBytes)
						if currWorker.Id == -1 {
							db.Table("tasks").Where("id= ?", task.Id).Update("state", common.InProgress)
						}
						//db.Table("tasks").Where("id = ?", task.Id).Update("worker", worker)
						log.Println("starting task for worker")
						server.startTask(worker.Id, task, service)

						availIndex += 1
					}
				}
			}
		}

		time.Sleep(3 * time.Second)
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

	http.FileServer(http.Dir("ui/build")).ServeHTTP(w, r)

}

type Context struct {
	Response http.ResponseWriter
	Request  *http.Request
	Server   *Server
	Params   []string
}

func (c *Context) Text(code int, body string) {
	c.Response.Header().Set("Content-Type", "application/json")
	c.Response.Header().Set("Access-Control-Allow-Origin", "*")
	c.Response.WriteHeader(code)

	payload := struct {
		Message string
	}{body}
	json.NewEncoder(c.Response).Encode(payload)
}

func (c *Context) Json(code int, body interface{}) {
	c.Response.Header().Set("Content-Type", "application/json")
	c.Response.Header().Set("Access-Control-Allow-Origin", "*")
	c.Response.WriteHeader(code)

	json.NewEncoder(c.Response).Encode(body)
}

func (c *Context) Error(code int, err error) {
	log.Println(code, err)
	c.Text(code, err.Error())
}

func (s *Server) handleJobs(ctx *Context) {
	r := ctx.Request
	db := s.db

	log.Printf("%s /jobs", r.Method)

	if r.URL.Path != "/jobs/" {
		w := ctx.Response
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	var tasks []Task
	switch r.Method {
	case "GET":
		rows, err := db.Table("jobs").Rows()
		if err != nil {
			log.Fatal(err)
		}
		var reply common.JobListReply

		// create zero length slice, so we don't return null
		reply.Jobs = make([]common.JobInfo, 0)
		var replyDetail common.JobInfo

		var taskCount int
		var completedCount int
		var queuedCount int
		var taskState common.TaskState
		var taskCreate time.Time
		var taskComplete time.Time
		var tempJobComplete time.Time
		var taskDuration time.Duration

		for rows.Next() {
			var job Job
			rows.Scan(&job.Id, &job.Type, &job.Params)
			replyDetail.JobId = job.Id
			replyDetail.Type = job.Type
			replyDetail.JobCompletedAt = time.Time{}

			/*Finding Job status*/
			rows, err := db.Table("tasks").Select("state, created_at, completed_at").Where("job_id = ?", job.Id).Count(&taskCount).Rows()
			if err != nil {
				ctx.Error(http.StatusBadRequest, err)
				return
			}
			completedCount = 0
			queuedCount = 0
			taskDuration = 0
			for rows.Next() {
				rows.Scan(&taskState, &taskCreate, &taskComplete)
				replyDetail.JobCreatedAt = taskCreate

				if taskState == common.Complete {

					/*Finds the MAX completion time of Tasks*/
					if taskDuration < taskComplete.Sub(taskCreate) {
                                                taskDuration = taskComplete.Sub(taskCreate)
                                                tempJobComplete = taskComplete
                                        }
					completedCount++
				} else if taskState == common.Queued {
					queuedCount++
				} else if taskState == common.InProgress {
					replyDetail.JobState = common.JobInProgress
				}
			}

			/*Checks if all tasks are completed or all tasks are queued*/
			if completedCount == taskCount {
				//Assigns max task completion time as Job completion time
				replyDetail.JobCompletedAt = tempJobComplete
				replyDetail.JobState = common.Completed
			} else if queuedCount == taskCount {
				replyDetail.JobState = common.NotStarted
			}

			/*Unmarshalling Job params*/
			if job.Type == common.IsAliveJob {
				var isAliveParam common.JobIsAliveParam
				err = json.Unmarshal([]byte(job.Params), &isAliveParam)
				if err != nil {
					log.Fatal(err)
				}
				replyDetail.Data = isAliveParam
			} else if job.Type == common.PortScanJob {
				var portScanParam common.JobPortScanParam
				err = json.Unmarshal([]byte(job.Params), &portScanParam)
				if err != nil {
					log.Fatal(err)
				}
				replyDetail.Data = portScanParam
			}
			reply.Jobs = append(reply.Jobs, replyDetail)
		}
		ctx.Json(http.StatusOK, reply)
		return
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
		workerCount := 0
		db.Table("workers").Where("updated_at > ?", time.Now().Add(-3*common.HeartbeatInterval)).Count(&workerCount)
		log.Println("active workers", workerCount)

		if workerCount == 0 {
			workerCount = 1
		}
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

				tasks = append(tasks, Task{Type: common.IsAliveTask, WorkerId: NotAllocatedWorkerId, State: common.Queued, Params: buf})
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

				tasks = append(tasks, Task{Type: common.PortScanTask, WorkerId: NotAllocatedWorkerId, State: common.Queued, Params: buf})
			}
		}
		job := Job{Type: jobType, Params: jobParams.Data, Tasks: tasks}
		db.Create(&job)
		ctx.Json(http.StatusOK, common.JobSubmitReply{job.Id})
		return
	}

}

func (s *Server) Setup() {
	// update all inprogress tasks with workerId -1
	if result := s.db.Model(&Task{}).Where("state = ?", common.InProgress).Update("worker_id", NotAllocatedWorkerId); result.Error != nil {
		log.Fatalln(result.Error)
	}
}

func (s *Server) handleJobID(ctx *Context) {
	r := ctx.Request
	params := ctx.Params
	db := s.db
	//w := ctx.Response
	var id int
	id, err := strconv.Atoi(params[0])
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("%s /jobs/%s JobId: %d", r.Method, params[0], id)

	//Checks if JobId exists
	var count int
	db.Table("jobs").Where("id = ?", id).Count(&count)
	if count == 0 {
		ctx.Text(http.StatusBadRequest, "Job doesn't exists")
		return
	}

	switch r.Method {
	case "GET":
		var reply common.JobDetailReply
		var replyDetail common.WorkerTaskData
		var jobType common.JobType
		var params string

		/*Getting Job information*/
		row := db.Table("jobs").Select("type,params").Where("id = ?", id).Row()
		row.Scan(&jobType, &params)

		reply.JobInfo.JobId = id
		reply.JobInfo.Type = jobType
		reply.JobInfo.JobCompletedAt = time.Time{}

		if jobType == common.IsAliveJob {
			var isAliveParam common.JobIsAliveParam
			err = json.Unmarshal([]byte(params), &isAliveParam)
			if err != nil {
				ctx.Error(http.StatusBadRequest, err)
			}
			reply.JobInfo.Data = isAliveParam
		} else if jobType == common.PortScanJob {
			var portScanParam common.JobPortScanParam
			err = json.Unmarshal([]byte(params), &portScanParam)
			if err != nil {
				ctx.Error(http.StatusBadRequest, err)
			}
			reply.JobInfo.Data = portScanParam
		}
		/*Getting tasks*/
		var taskCount int
		var taskCreate time.Time
                var taskComplete time.Time
                var tempJobComplete time.Time
                var taskDuration time.Duration

		rows, err := db.Table("tasks").Select("id, type, state, worker_id, created_at, completed_at,  result").Where("job_id = ?", id).Count(&taskCount).Rows()
		if err != nil {
			ctx.Error(http.StatusBadRequest, err)
			return
		}
		defer rows.Close()
		var taskId int
		var workerId int
		var result string
		var taskType common.TaskType
		var taskState common.TaskState
		var workerName string
		var workerAddress string
		workerName = ""
		workerAddress = ""
		var completedCount int
		var queuedCount int
		completedCount = 0
		queuedCount = 0
		taskDuration = 0
		for rows.Next() {
			rows.Scan(&taskId, &taskType, &taskState, &workerId, &taskCreate, &taskComplete, &result)
			reply.JobInfo.JobCreatedAt = taskCreate

			//Getting worker name
			if workerId != -1 {
				row := db.Table("workers").Select("name, address").Where("id = ?", workerId).Row()
				row.Scan(&workerName, &workerAddress)
			}
			//Creating Reply Struct
			replyDetail.TaskId = taskId
			replyDetail.TaskState = taskState
			replyDetail.WorkerId = workerId
			replyDetail.WorkerName = workerName
			replyDetail.WorkerAddress = workerAddress
			replyDetail.CreateTime = taskCreate
			replyDetail.CompleteTime = taskComplete
			replyDetail.Data = struct{}{}
			if taskState == common.Complete {
				if taskDuration < taskComplete.Sub(taskCreate) {
                                        taskDuration = taskComplete.Sub(taskCreate)
                                        tempJobComplete = taskComplete
                                }
				completedCount++
				if taskType == common.IsAliveTask {
					var isAliveResult common.IsAliveResult
					err = json.Unmarshal([]byte(result), &isAliveResult)
					if err != nil {
						log.Fatal(err)
					}
					replyDetail.Data = isAliveResult
				} else if taskType == common.PortScanTask {
					var portScanResult common.PortScanResult
					err = json.Unmarshal([]byte(result), &portScanResult)
					if err != nil {
						log.Fatal(err)
					}
					replyDetail.Data = portScanResult
				}
			} else if taskState == common.Queued {
				queuedCount++
			} else if taskState == common.InProgress {
				reply.JobInfo.JobState = common.JobInProgress
			}

			reply.Data = append(reply.Data, replyDetail)
		}
		//Find if all tasks are completed
		if completedCount == taskCount {
			//Finds Job completion time from the max task completion time
			reply.JobInfo.JobCompletedAt = tempJobComplete
			reply.JobInfo.JobState = common.Completed
		} else if queuedCount == taskCount {
			reply.JobInfo.JobState = common.NotStarted
		}
		ctx.Json(http.StatusOK, reply)
		return
	case "DELETE":
		if success := db.Delete(&Job{Id: id}).Error == nil && db.Delete(Task{}, "job_id = ?", id).Error == nil; success {
			ctx.Text(http.StatusOK, "Deleted job successfully")
		} else {
			ctx.Text(http.StatusBadRequest, "Failed to delete job")
		}

		return
	}
}

func main() {
	var serverAddr string
	var rpcAddr string
	var dbPath string
	var clean bool
	flag.StringVar(&serverAddr, "serverAddr", "localhost:8080", "address of http service")
	flag.StringVar(&rpcAddr, "rpcAddr", "localhost:8081", "address of rpc service")
	flag.StringVar(&dbPath, "db", "test.db", "database path")
	flag.BoolVar(&clean, "clean", false, "cleans old database if it exists")
	flag.Parse()
	fmt.Println("starting server")
	var wg sync.WaitGroup

	// remove the old database
	if clean {
		err := os.Remove(dbPath)

		if err != nil {
			log.Fatal(err)
		}
	}
	gob.Register(common.IsAliveParam{})
	gob.Register(common.IsAliveResult{})
	gob.Register(common.PortScanParam{})
	gob.Register(common.PortScanResult{})

	// setup database
	db, err := gorm.Open("sqlite3", dbPath)
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()
	db.LogMode(true)

	// check the ui exists
	if _, err := os.Stat("ui/build"); os.IsNotExist(err) {
		log.Fatal("Failed to find ui build")
	}

	// Migrate the schema
	db.AutoMigrate(&Job{})
	db.AutoMigrate(&Task{})
	db.AutoMigrate(&Worker{})

	server := new(Server)
	server.db = db

	// star custom Mux, for dynamic routing from client-server interactions
	server.Setup()
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
	go server.Schedule(service)

	wg.Wait()
}
