package main

import (
	"encoding/gob"
	"flag"
	"github.com/google/gopacket/routing"
	common "github.com/iakshay/jarvis-scanner"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type Worker struct {
	Id     int
	Client *rpc.Client

	common.SendTaskArgs
}

func (worker *Worker) doTask() {
	log.Println("Doing Task")
	args := &common.CompleteTaskArgs{}
	var reply common.CompleteTaskReply

	switch worker.TaskType {
	case common.IsAliveTask:
		isAliveParam, ok := worker.TaskData.(common.IsAliveParam)
		if !ok {
			log.Fatal("Invalid param data")
		}
		log.Println(isAliveParam)
		args.Result = common.IsAlive(isAliveParam.IpRange)
	case common.PortScanTask:
		portScanParam, ok := worker.TaskData.(common.PortScanParam)
		if !ok {
			log.Fatal("Invalid param data")
		}
		if portScanParam.Type == common.NormalScan {
			args.Result = common.NormalPortScan(portScanParam.Ip, portScanParam.PortRange, 6*time.Second)

		} else {
			router, err := routing.New()
			if err != nil {
				log.Fatal("routing error:", err)
			}
			s, err := common.NewScanner(portScanParam.Ip, router)
			if err != nil {
				log.Fatal("failed to create scanner:", err)
			}
			defer s.Close()
			log.Println(portScanParam.Ip, portScanParam.Type, portScanParam.PortRange)
			args.Result, err = s.Scan(portScanParam.Type, portScanParam.PortRange)
			if err != nil {
				log.Fatal("failed to scan:", err)
			}
		}
	default:
		log.Fatal("Invalid task type")
	}
	args.TaskId = worker.TaskId
	//log.Printf("%v %T", args.Result, args.Result)
	// reset worker state
	worker.resetState()
	if err := worker.Client.Call("Server.CompleteTask", args, &reply); err != nil {
		log.Fatal(err)
	}

	log.Println("completed task")
}

func (worker *Worker) resetState() {
	worker.TaskId = -1
}

func (worker *Worker) SendTask(args *common.SendTaskArgs, reply *common.SendTaskReply) error {
	if worker.TaskId != -1 {
		log.Fatal("already process task!")
	}

	worker.TaskData = args.TaskData
	worker.TaskId = args.TaskId
	worker.TaskType = args.TaskType
	go worker.doTask()

	return nil
}

func (worker *Worker) runHearbeat() {
	args := &common.HeartbeatArgs{WorkerId: worker.Id}
	var reply common.HeartbeatReply
	for {
		worker.Client.Call("Server.Heartbeat", args, &reply)
		time.Sleep(common.HeartbeatInterval)
	}
}

func main() {
	gob.Register(common.IsAliveParam{})
	gob.Register(common.IsAliveResult{})
	gob.Register(common.PortScanParam{})
	gob.Register(common.PortScanResult{})
	var serverAddr string
	var workerAddr string
	flag.StringVar(&serverAddr, "serverAddr", "localhost:8081", "address of the server")
	flag.StringVar(&workerAddr, "workerAddr", "localhost:7071", "address of the worker")
	flag.Parse()
	var wg sync.WaitGroup
	wg.Add(1)

	// keep trying to connect to the server
	client, err := rpc.DialHTTP("tcp", serverAddr)
	for err != nil {
		log.Println("dialing:", err)
		time.Sleep(5 * time.Second)
		client, err = rpc.DialHTTP("tcp", serverAddr)
	}

	worker := new(Worker)
	worker.Client = client
	worker.TaskId = -1

	// open upto requests from server
	log.Println("Starting worker")
	rpc.Register(worker)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", workerAddr)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)

	// register worker
	workerName, err := os.Hostname()
	if err != nil {
		workerName = "worker"
	}
	args := &common.RegisterWorkerArgs{Name: workerName, Address: workerAddr}
	var reply common.RegisterWorkerReply

	err = client.Call("Server.RegisterWorker", args, &reply)
	if err != nil {
		log.Fatal("RegisterWorker error:", err)
	}
	worker.Id = reply.WorkerId

	// heartbeat go routine
	go worker.runHearbeat()
	wg.Wait()
}
