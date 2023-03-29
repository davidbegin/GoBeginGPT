package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func home(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "home\n")

	// Can we do something!
	// that our other program will know about
}

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")

	fmt.Printf("Request: %s", req.Body)

	now := time.Now() // current local time
	// Convert some strings to bytes
	nsec := now.UnixNano()
	b := []byte(fmt.Sprintf("%v", nsec))

	// write the whole body at once
	err := ioutil.WriteFile("output.txt", b, 0644)
	if err != nil {
		panic(err)
	}
}

func headers(w http.ResponseWriter, req *http.Request) {

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func main() {

	// So we just need to do something
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/headers", headers)
	http.HandleFunc("/", home)

	http.ListenAndServe(":8080", nil)
}

// PUSHER_APP_KEY=a6a7b7662238ce4494d5
// PUSHER_APP_ID=1555452
// PUSHER_APP_CLUSTER=mt1
// Pusher channel name is api_client_status_update_ + api_secret:
// api_client_status_update_oDWtE_tWvsPZNlw5BdDFtj2qlwM1sVQH

// func main2() {

// 	pusherClient := pusher.Client{
// 		AppId:   "APP_ID",
// 		Key:     "APP_KEY",
// 		Secret:  "APP_SECRET",
// 		Cluster: "APP_CLUSTER",
// 	}

// 	data := map[string]string{"message": "hello world"}
// 	pusherClient.Trigger("my-channel", "my-event", data)
// }
