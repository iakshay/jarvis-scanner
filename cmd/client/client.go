package main

import (
	"strconv"
	"flag"
	"fmt"
	"os"
	"net/http"
	"io/ioutil"
	"log"
	"bytes"
	"encoding/json"
)

func Usage() {
	fmt.Printf("Usage: \nFor task=list; general format of execution is:\n"+
		"\t./client -task=list\n\n"+
		"For task=view; general format of execution is:\n"+
		"\t./client -task=view id=5\n\n"+
		"For task=delete; general format of execution is:\n"+
		"\t./client -task=delete id=4\n\n"+
		"For task=submit; general format of execution is:\n"+
		"\t./client -task=submit -type=isAlive -ip=1.2.3.4/255\n"+
		"\t./client -task=submit -type=portScan -ip=1.2.3.4 -mode=SYN -range=100-200\n\n")

	fmt.Printf("Possible values for the flags are:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = Usage
	var taskName string
	var jobId int
	var jobType string
	var IP string
	var scanMode string
	var portRange string


	flag.StringVar(&taskName, "task", "", "Enter the type of task: {list, view, delete or submit}")
	flag.IntVar(&jobId, "id", 0, "Enter the id of the job: { greater than 0}")
	flag.StringVar(&jobType, "type", "", "Enter the type of the job: {isAlive or portScan}")
	flag.StringVar(&IP, "ip", "", "Enter the ip address or range of ip address : {1.2.3.4/255}")
	flag.StringVar(&scanMode, "mode", "", "Enter the mode of the scanning : {Normal,SYN or FIN}")
	flag.StringVar(&portRange, "range", "", "Enter the range of port for scanning : {100-200}")

	flag.Parse()

	if taskName == "list" {
		fmt.Println("Task name is: ",taskName)
		//creating get request for list
		resp, err := http.Get("http://localhost:8080/job/?task="+string(taskName))
	        if err != nil {
		        log.Fatalln(err)
	        }
	        defer resp.Body.Close()
		//reading the response body
		body, err := ioutil.ReadAll(resp.Body)
	        if err != nil {
	                log.Fatalln(err)
	        }
		log.Println("\nReturned value from server:\n",string(body))

	} else if (taskName == "view" || taskName == "delete") && jobId > 0{
		fmt.Printf("value of task: %s and id: %s\n",taskName, jobId)
		//creating get request for view and delete specific job ID
		resp, err := http.Get("http://localhost:8080/job/?task="+taskName+"&id="+strconv.Itoa(jobId))
		if err != nil {
                        log.Fatalln(err)
                }
                defer resp.Body.Close()
                body, err := ioutil.ReadAll(resp.Body)
                if err != nil {
                        log.Fatalln(err)
                }
		log.Println("\nReturned value from server:\n",string(body))
	} else if taskName == "submit" {
		if jobType == "isAlive" && IP != "" {
			fmt.Printf("value of task:%s, jobtype:%s, ip:%s\n",taskName, jobType, IP)
			//creating json form
			message := map[string]interface{}{
		                "type": jobType,
			        "ip": IP,
			}
			bytesRepresentation, err := json.Marshal(message)
			if err != nil {
				log.Fatalln(err)
			}
			//creating post request for isAlive
			resp, err := http.Post("http://localhost:8080/job/", "application/json", bytes.NewBuffer(bytesRepresentation))
			if err != nil {
				log.Fatalln(err)
			}
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			log.Println("\nReturned value from server: ",string(body))

		} else if jobType == "portScan" && IP != "" && scanMode != "" && portRange != ""{
			fmt.Printf("value of task:%s, jobtype:%s, ip:%s, scanMode:%s, portRange:%s\n",taskName, jobType, IP, scanMode, portRange)
			//creating post request for portScan
			message := map[string]interface{}{
                                "type": jobType,
                                "ip": IP,
				"mode": scanMode,
				"range":portRange,
                        }
                        bytesRepresentation, err := json.Marshal(message)
                        if err != nil {
                                log.Fatalln(err)
                        }
                        resp, err := http.Post("http://localhost:8080/job/", "application/json", bytes.NewBuffer(bytesRepresentation))
                        if err != nil {
                                log.Fatalln(err)
                        }
                        defer resp.Body.Close()
                        body, _ := ioutil.ReadAll(resp.Body)
			log.Println("\nReturned value from server: ",string(body))

		} else {
			flag.Usage()
			os.Exit(0)
		}
        } else {
		flag.Usage()
		os.Exit(0)
	}
}

