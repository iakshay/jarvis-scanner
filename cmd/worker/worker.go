package main

import (
	"flag"
	"fmt"
	"github.com/google/gopacket/routing"
	common "github.com/iakshay/jarvis-scanner"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

type Worker struct {
	Id     int
	Client *rpc.Client

	common.SendTaskArgs
}

func (worker *Worker) DoTask() {
	fmt.Println("Doing Task")
	args := &common.CompleteTaskArgs{}
	var reply common.CompleteTaskReply

	switch worker.TaskType {
	case common.IsAliveTask:
		isAliveParam, ok := worker.TaskData.(common.IsAliveParam)
		if !ok {
			log.Fatal("Invalid param data")
		}

		args.Result = common.IsAlive(isAliveParam.IpRange)
	case common.PortScanTask:
		portScanParam, ok := worker.TaskData.(common.PortScanParam)
		if !ok {
			log.Fatal("Invalid param data")
		}
		if portScanParam.Type == common.NormalScan {
			args.Result = common.NormalPortScan(portScanParam.Ip, portScanParam.PortRange, 3*time.Second)

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
			args.Result, err = s.Scan(portScanParam.Type, portScanParam.PortRange)
		}
	default:
		log.Fatal("Invalid task type")
	}

	worker.Client.Call("Server.CompleteTask", args, &reply)
}

func (worker *Worker) SendTask(args *common.SendTaskArgs, reply *common.SendTaskReply) error {
	if worker.TaskId != -1 {
		log.Fatal("already process task!")
	}

	worker.TaskData = args.TaskData
	worker.TaskId = args.TaskId
	go worker.DoTask()

	return nil
}

func (worker *Worker) RunHearbeat() {
	args := &common.HeartbeatArgs{WorkerId: worker.Id}
	var reply common.HeartbeatReply
	for {
		fmt.Println("Sending hearbeat")
		worker.Client.Call("Server.Heartbeat", args, &reply)
		time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	var serverAddr string
	var workerAddr string
	flag.StringVar(&serverAddr, "serverAddr", "localhost:8080", "address of the server")
	flag.StringVar(&workerAddr, "workerAddr", "localhost:7070", "address of the worker")
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
	fmt.Println("Starting worker")
	rpc.Register(worker)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", workerAddr)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)

	// register worker
	args := &common.RegisterWorkerArgs{Name: "worker", Address: workerAddr}
	var reply common.RegisterWorkerReply

	err = client.Call("Server.RegisterWorker", args, &reply)
	if err != nil {
		log.Fatal("RegisterWorker error:", err)
	}
	worker.Id = reply.WorkerId

	// heartbeat go routine
	go worker.RunHearbeat()

	wg.Wait()
}
