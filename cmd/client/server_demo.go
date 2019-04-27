package main

import (
	"flag"
	"fmt"
	"encoding/json"
	common "github.com/iakshay/jarvis-scanner"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"io"
	"io/ioutil"
	"log"
//	"math"
//	"net"
	"net/http"
	"net/rpc"
	"regexp"
	"strconv"
//	"strings"
	"sync"
	"time"
)

type Task struct {
	Id          int
	JobId       int
	State       common.TaskState
	Params      []byte	// Allows for UnMarshalling to struct objects, as needed
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

	connections map[int]*rpc.Client
	Routes      []Route
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
		args := &common.SendTaskArgs{TaskData: common.IsAliveParam{}, TaskType: common.IsAliveTask, TaskId: 1}
		var reply common.SendTaskReply
		client.Call("Worker.SendTask", args, reply)
		break
	}
}
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
		freeWorkers := make([]Worker, numWorkers)
		freeWorkerIndex := 0

		queuedTasks, inProgressTasks, completeTasks := make([]Task, 0, numWorkers), make([]Task, 0, numWorkers), make([]Task, 0, numWorkers)

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
				freeWorkers[freeWorkerIndex] = worker
				freeWorkerIndex += 1
			}
		}

		index := 0
		for (index < numWorkers) && (queuedTasks[index] != nil) && (freeWorkers[index] != nil) {
			if (time.Now().Sub(freeWorkers[index].UpdatedAt)) > common.LifeCycle {
				db.Delete(&(freeWorkers[index]))
			}
		}

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
/*
func ipTo32Bit(IP net.IP) int {
	total := 0
	power := 0

	for _, level := range ipArray {
		shift := math.Pow(2, power)
		total += (level * shift)
		power += 8
	}

	return num
}

func bitsToIP(value int) net.IP {
	arr := make([]int, 4)
	divisor := math.Pow(2, 32)

	for i := 0; i < 4; i++ {
		section := 0
		for j := 0; j < 8; j++ {
			quotient := value / divisor
			if quotient == 1 {
				section += math.Pow(2, j)
			}
			divisor = divisor/2
		}
		arr[i] = section
	}

	return net.IPv4(arr[0], arr[1], arr[2], arr[3])

}
*/
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
		//Parsing get request
                tasks, ok := r.URL.Query()["task"]

                if !ok || len(tasks[0]) < 1 {
                log.Println("Task is missing")
                return
                }

                task := tasks[0]

		if task == "list" {
			rows, err := db.Table("jobs").Rows()
			if err != nil {
				log.Fatal(err)
			}
			for rows.Next() {
				var job Job
				rows.Scan(&job.Id, &job.Params)
				io.WriteString(w,"JobId: " + strconv.Itoa(job.Id) +" param:"+string(job.Params)+"\n")
			}
			return
		} else if task == "view" {
			id, ok := r.URL.Query()["id"]
			if !ok || len(id[0]) < 1 {
				log.Println("Id is missing")
				return
			}
			idval := id[0]
			rows, err := db.Table("jobs").Where("Id = ?", idval).Rows()
			if err != nil {
				log.Fatal(err)
			}
			for rows.Next() {
				var jobs Job
				rows.Scan(&jobs.Id, &jobs.Params)
				io.WriteString(w,"JobId: " + strconv.Itoa(jobs.Id) +" param:"+string(jobs.Params)+"\n")
			}
			fmt.Printf("Task is: %s , Id: is %s\n",task,idval)
			return
		}  else if task == "delete" {
			id, ok := r.URL.Query()["id"]
			if !ok || len(id[0]) < 1 {
				log.Println("Id is missing")
				return
			}
			idval := id[0]
			fmt.Printf("Task is: %s , Id: is %s\n",task,idval)
			return
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
				fmt.Printf("value of type is: %d\n",jobParams.Type)

				typVal := jobParams.Type
//				var workerCount int
//				db.Table("workers").Count(&workerCount)
//				tasks := make([]Task, workerCount)
				if typVal == common.IsAliveJob {
					var IsAlive common.JobIsAliveParam;
					err = json.Unmarshal([]byte(jobParams.Data), &IsAlive)
					if err != nil {
						log.Fatal(err)
					}
					fmt.Printf("Value of Ipblock: %s\n",IsAlive.IpBlock)
				} else {
					var PortScan common.JobPortScanParam
					err = json.Unmarshal([]byte(jobParams.Data), &PortScan)
					if err != nil {
						log.Fatal(err)
					}
					fmt.Printf("Value of Ip: %v\n",PortScan.Ip)
					fmt.Printf("Value of ScanType: %d\n",PortScan.Type)
					fmt.Printf("Value of StartPort: %d\n",PortScan.StartPort)
					fmt.Printf("Value of EndPort: %d\n",PortScan.EndPort)
				}
					/*
					IPSplit := strings.Split(jobParams.IpBlock, "/")
					// IP struct stores most recent 32-bit IP address in final four bytes of array
					IPBlock := net.ParseIP(IPSplit[0])[12:16]
					IPMask := strconv.Atoi(IPSplit[1])
					subnetSize := math.Pow(2, 32 - IPMask)
					quotientWork := subnetSize/workerCount
					remainderWork := math.Mod(subnetSize, workerCount)
					IP32Rep := ipTo32Bit(IPBlock)
					for i := 0; i <  workerCount; i++ {
						nextIP32Base := IP32Rep + quotientWork
						if remainderWork > 0 {
							nextIP32Base += 1
							remainderWork = remainderWork - 1
						}
						startIP := bitsToIP(IP32Rep)
						endIP := bitsToIP(nextIP32Base) - 1
						IpRange := IpRange{startIP, endIP}
						taskParamData := IsAliveParam{IpRange}
						buf, e := json.Marshal(taskParamData)
						if e != nil {
							log.Fatal(e)
						}

						task := Task{common.Queued, buf}
						db.Create(&task)
						tasks[i] = task
						IP32Rep = nextIP32Base
					}

				} else {
					IP := jobParams.Ip
					Type := jobParams.Type
					Range := jobParams.Range
						Start := Range.Start
					End := Range.End
					rangeLength := (End - Start) + 1
					quotientWork := (rangeLength/workerCount) - 1
					remainderWork := math.Mod(rangeLength, workerCount)
					currStart := Start
					var currEnd uint16
					for i := 0; i < workerCount; i++ {
						currEnd = currStart + quotientWork
						if remainderWork > 0 {
							currEnd += 1
							remainderWork = remainderWork - 1
						}
						taskRange := common.PortRange{currStart, currEnd}
						taskParamData := common.PortScanParam{Type, IP, taskRange}

						buf, e := json.Marshal(taskParamData)
						if e != nil {
							log.Fatal(e)
						}

						currStart = currEnd + 1

						task := Task{Queued, buf}
						db.Create(&task)
						tasks[i] = task
					}
				}*/
				job := Job{Params: b}
				db.Create(&job)
		return

	}

}

func main() {
	var serverAddr string
	var dbPath string
	flag.StringVar(&serverAddr, "serverAddr", "localhost:8080", "address of the server")
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
	server.connections = make(map[int]*rpc.Client)

	/*Inserting dummy rows in job*/
/*	message := map[string]interface{}{
                                "type": "isAlive",
                                "ip": "1.2",
				"mode": "sys",
				"range":"100-200",
                        }
	bytesRepresentation, err := json.Marshal(message)
                        if err != nil {
                                log.Fatalln(err)
                        }
	job := Job{Id:3,Params: bytesRepresentation}
	db.Create(&job)
*/
	server.Handle("/jobs/", func(ctx *Context) {
		server.handleJobs(ctx)
	})

	wg.Add(1)
	go http.ListenAndServe(serverAddr, server)

	// start rpc server
/*
	rpc.Register(server)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", serverAddr)
	if e != nil {
		log.Fatal("listen error:", e)
	}

	wg.Add(1)
	go http.Serve(l, nil)
*/
	// start thread for Scheduler aspect of Server
	//go server.Schedule()

	time.Sleep(3 * time.Second)
	// start dummy task on one worker
//	server.startTask()
	wg.Wait()
}

