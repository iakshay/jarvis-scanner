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
	"net"
)

func Usage() {
	fmt.Printf("Usage: \nFor task=list; general format of execution is:\n"+
		"\t./client -task=list\n\n"+
		"For task=view; general format of execution is:\n"+
		"\t./client -task=view -id=5\n\n"+
		"For task=delete; general format of execution is:\n"+
		"\t./client -task=delete -id=4\n\n"+
		"For task=submit; general format of execution is:\n"+
		"\t./client -task=submit -type=isAlive -ip=1.2.3.4/255\n"+
		"\t./client -task=submit -type=portScan -ip=1.2.3.4 -mode=Syn -start=100 -end=200\n\n")

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
	var rangeStart string
	var rangeEnd string


	flag.StringVar(&taskName, "task", "", "Enter the type of task: {list, view, delete or submit}")
	flag.IntVar(&jobId, "id", 0, "Enter the id of the job: { greater than 0}")
	flag.StringVar(&jobType, "type", "", "Enter the type of the job: {isAlive or portScan}")
	flag.StringVar(&IP, "ip", "", "Enter the ip address or range of ip address : {1.2.3.4/255}")
	flag.StringVar(&scanMode, "mode", "", "Enter the mode of the scanning : {Normal,SYN or FIN}")
	flag.StringVar(&rangeStart, "start", "", "Enter the start port of range of port for scanning")
	flag.StringVar(&rangeEnd, "end", "", "Enter the end port of range of port for scanning")



	flag.Parse()

	if taskName == "list" {
		fmt.Println("Task name is: ",taskName)
		//creating get request for list
		resp, err := http.Get("http://localhost:8080/jobs/")
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

	} else if taskName == "view" && jobId > 0 {
		fmt.Printf("value of task: %s and id: %d\n",taskName, jobId)
		resp, err := http.Get("http://localhost:8080/jobs/"+strconv.Itoa(jobId))
		if err != nil {
                        log.Fatalln(err)
                }
		defer resp.Body.Close()
                body, err := ioutil.ReadAll(resp.Body)
                if err != nil {
                        log.Fatalln(err)
                }
                log.Println("\nReturned value from server:\n",string(body))
	}  else if taskName == "delete" && jobId > 0 {
		client := &http.Client{}

		// Create request
		req, err := http.NewRequest("DELETE", "http://localhost:8080/jobs/"+strconv.Itoa(jobId), nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		// Fetch Request
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()

		// Read Response Body
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
		fmt.Println(err)
		return
		}
		fmt.Println("response Body : ", string(respBody))
	} else if taskName == "submit" {
		if jobType == "isAlive" && IP != "" {
			fmt.Printf("value of task:%s, jobtype:%s, ip:%s\n",taskName, jobType, IP)
			//creating json form
			message := map[string]interface{}{
		                "Type" : 0,
			        "Data" :map[string]interface{} {"IpBlock" : IP},
			}
/*			var jsonStr = `{
				"Type": 0,
				"Data": {"IpBlock": IP}
			}
			`
			jsonMap := make(map[string]interface{})
			err := json.Unmarshal([]byte(jsonStr), &jsonMap)
			if err != nil {
				panic(err)
			}*/
			bytesRepresentation, err := json.Marshal(message)
			if err != nil {
				log.Fatalln(err)
			}
			//creating post request for isAlive
			resp, err := http.Post("http://localhost:8080/jobs/", "application/json", bytes.NewBuffer(bytesRepresentation))
			if err != nil {
				log.Fatalln(err)
			}
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			log.Println("\nReturned value from server: ",string(body))

		} else if jobType == "portScan" && IP != "" && scanMode != "" && rangeStart != "" && rangeEnd != ""{
			fmt.Printf("value of task:%s, jobtype:%s, ip:%s, scanMode:%s, rangeStart:%s, rangeEnd%s\n",taskName, jobType, IP, scanMode, rangeStart, rangeEnd)
			//Converting variables for sending
			var IpAddress net.IP
			IpAddress = net.ParseIP(IP)
			var ScanType int
			if scanMode == "Normal"{
				ScanType = 0
			} else if scanMode == "Syn" {
				ScanType = 1
			} else if scanMode == "Fin" {
				ScanType = 2
			} else {
				 os.Exit(0)
			}
			ValStart, err := strconv.Atoi(rangeStart)
			if err != nil {
                                log.Fatalln(err)
                        }
			ValEnd, err := strconv.Atoi(rangeEnd)
			if err != nil {
                                log.Fatalln(err)
                        }
			//creating post request for portScan
			message := map[string]interface{}{
                                "Type": 1,
				"Data": map[string]interface{}{ "Ip":IpAddress, "ScanType": ScanType, "Start" : ValStart, "End": ValEnd},
			}
                        bytesRepresentation, err := json.Marshal(message)
                        if err != nil {
                                log.Fatalln(err)
                        }
			resp, err := http.Post("http://localhost:8080/jobs/", "application/json", bytes.NewBuffer(bytesRepresentation))
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

