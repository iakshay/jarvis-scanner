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
	"math"
	"net"
	"net/http"
	"net/rpc"
	"regexp"
	"strconv"
	"strings"
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
			divisor = divisor / 2
		}
		arr[i] = section
	}

	return net.IPv4(arr[0], arr[1], arr[2], arr[3])

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
						io.WriteString(w, "Job "+strconv.Itoa(task.JobId)+": complete")
					} else {
						io.WriteString(w, "Job "+strconv.Itoa(task.JobId)+": incomplete")
					}
					completed = true
					io.WriteString(w, "\n\n")
				}
			}

			if task.State != Complete {
				completed = false
			}

			io.WriteString(w, task.Params+"\n")
		}
		return
	case "POST":
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		var jobParams JobSubmitParam
		err = json.Unmarshal(b, &jobParams)
		if err != nil {
			log.Fatal(err)
		}

		typVal := jobParams.Type
		var workerCount int
		db.Table("workers").Count(&workerCount)
		tasks := make([]Task, workerCount)
		if typVal == IsAliveJob {
			IPSplit := strings.Split(jobParams.IpBlock, "/")
			// IP struct stores most recent 32-bit IP address in final four bytes of array
			IPBlock := net.ParseIP(IPSplit[0])[12:16]
			IPMask := strconv.Atoi(IPSplit[1])
			subnetSize := math.Pow(2, 32-IPMask)
			quotientWork := subnetSize / workerCount
			remainderWork := math.Mod(subnetSize, workerCount)
			IP32Rep := ipTo32Bit(IPBlock)
			for i := 0; i < workerCount; i++ {
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

				task := Task{Queued, buf}
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
			quotientWork := (rangeLength / workerCount) - 1
			remainderWork := math.Mod(rangeLength, workerCount)
			currStart := Start
			var currEnd uint16
			for i := 0; i < workerCount; i++ {
				currEnd = currStart + quotientWork
				if remainderWork > 0 {
					currEnd += 1
					remainderWork = remainderWork - 1
				}
				taskRange := PortRange{currStart, currEnd}
				taskParamData := PortScanParam{Type, IP, taskRange}
				buf, e := json.Marshal(taskParamData)
				if e != nil {
					log.Fatal(e)
				}

				currStart = currEnd + 1

				task := Task{Queued, buf}
				db.Create(&task)
				tasks[i] = task
			}
		}
		job := Job{Params: b, Tasks: tasks}
		db.Create(&job)
		return
	}

}

func (s *Server) handleJobID(ctx *Context) {
	r := ctx.Request

	switch r.Method {
	case "GET":
		return
	case "DELETE":
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

	wg.Add(1)
	go http.ListenAndServe(serverAddr, server)

	// start rpc server
	rpc.Register(server)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", serverAddr)
	if e != nil {
		log.Fatal("listen error:", e)
	}

	wg.Add(1)
	go http.Serve(l, nil)

	time.Sleep(3 * time.Second)
	// start dummy task on one worker
	server.startTask()
	wg.Wait()
}
