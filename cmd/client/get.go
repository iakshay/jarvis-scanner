package main

import (
	"net/http"
	"io/ioutil"
	"log"
)

func main() {

//	resp, err := http.Get("https://httpbin.org/get")
//	resp, err := http.Get("http://localhost:8080/?key=hello%20golangcode.com:")
	resp, err := http.Get("http://localhost:8080/job/")
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
                log.Fatalln(err)
        }

	log.Println(string(body))
}
