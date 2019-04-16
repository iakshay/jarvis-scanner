package main

import (
	"log"
	"net/http"
	"io/ioutil"
	"io"
)

func main() {
    http.HandleFunc("/job/", handler)
    http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {


/*	keys, ok := r.URL.Query()["key"]
	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'key' is missing")
		return
	}

	// Query()["key"] will return an array of items, 
	// we only want the single item.
	key := keys[0]

	log.Println("Url Param 'key' is: " + string(key))

	log.Println("Url Param")
//	io.WriteString(w, "Job ")
	w.Write([]byte("Received a GET request\n"))
*/

	switch r.Method {
	case "GET":
		log.Println("get")
//		w.Write([]byte("Received a GET request\n"))
		io.WriteString(w,"value")

	case "POST":
		log.Println("post")
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(string(reqBody))
		w.Write([]byte("Received a POST request\n"))
	}
}
