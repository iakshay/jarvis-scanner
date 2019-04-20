package main

import (
	"log"
	"net/http"
	"io/ioutil"
	"io"
	"fmt"
	"encoding/json"
	"bytes"
)

/*Creating a struct to parse json of post request*/
type submit struct {
	TypeName    string   `json:"type,omitempty"`
	IP string  `json:"ip,omitempty"`
	Mode string  `json:"mode,omitempty"`
	Range string  `json:"range,omitempty"`
}


func main() {
    http.HandleFunc("/job/", handler)
    http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {


	switch r.Method {
	case "GET":
		//Parsing get request
		tasks, ok := r.URL.Query()["task"]

		if !ok || len(tasks[0]) < 1 {
		log.Println("Task is missing")
		return
		}

		task := tasks[0]

		if task == "list" {
			fmt.Printf("Task is: %s\n",task)
			//Code for showing list of Jobs
			io.WriteString(w,"\nReceived a GET request for task = "+string(task)+"\n")
		} else if task == "view" {
			id, ok := r.URL.Query()["id"]
			if !ok || len(id[0]) < 1 {
				log.Println("Id is missing")
				return
			}
			idval := id[0]
			fmt.Printf("Task is: %s , Id: is %s\n",task,idval)
			//writing to response body
			io.WriteString(w,"\nReceived a GET request for task: "+task+" and id is "+idval+"\n")
			//code for showing specific JobId
		} else if task == "delete" {
                        id, ok := r.URL.Query()["id"]
                        if !ok || len(id[0]) < 1 {
                                log.Println("Id is missing")
				return
                        }
                        idval := id[0]
                        fmt.Printf("Task is: %s , Id is: %s\n",task,idval)
			io.WriteString(w,"\nReceived a GET request for task: "+task+" and id is "+idval+"\n")
			//Code for deleting specific JobId
                } else{
			log.Println("Invalid Input")
			return
		}

	case "POST":
		//parsing json
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		new := bytes.NewReader(reqBody)
		decoder := json.NewDecoder(new)
		jsonParse := &submit{}
		err1 := decoder.Decode(jsonParse)
		if err1 != nil {
		        log.Fatal(err1)
		}
		if jsonParse.TypeName == "isAlive" {
			fmt.Printf("Submitted job type is: %s and ip is: %s\n",jsonParse.TypeName,jsonParse.IP)
			//io.WriteString(w,"\nReceived a post request for jobtype: "+jsonParse.TypeName+" and ip is "+jsonParse.IP+"\n")
			io.WriteString(w,"\nJob "+jsonParse.TypeName+" is successfully posted\n")
			//code for isAlive
		} else if jsonParse.TypeName == "portScan" {
			fmt.Printf("Submitted job type is: %s and ip is: %s\n",jsonParse.TypeName,jsonParse.IP)
			fmt.Printf("Mode is: %s and Range is: %s\n",jsonParse.Mode,jsonParse.Range)
			//io.WriteString(w,"\nReceived a post request "+jsonParse.TypeName+", "+jsonParse.IP+", "+jsonParse.Mode+
			//		", "+jsonParse.Range+"\n")
			io.WriteString(w,"\nJob "+jsonParse.TypeName+" is successfully posted\n")
			//code for portscan
		} else {
			log.Println("Invalid Input")
			return
		}
	}
}
