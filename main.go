package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	nextPort = 8000
	portMutex sync.Mutex
)

func main(){
	mux := http.NewServeMux()

	mux.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		println("hit /")
		http.ServeFile(w, r, "index.html")
	})

	mux.HandleFunc("/spawn", handleSpawn);

	// actual binding call 
	if err := http.ListenAndServe(":3000", mux); err != nil {
		fmt.Println("server failed: ", err)
	}
}

func handleSpawn(w http.ResponseWriter, r *http.Request){
	// get the port 
	port := getPort()

	// spawn the actual server 
	go spawnServer(port)
	// can't return value from goroutine??? 
	// if err != nil {
	// 	fmt.Fprintf(w, "<h2>Failed to spawn server</h2>")
	// }

	// return the string that the server has been spawned 
	fmt.Fprintf(w, "<h2>Server spawned on port: %v</h2>", port)
}

func getPort() string {
	portMutex.Lock()
	defer portMutex.Unlock()

	nextPort += 1 
	return strconv.Itoa(nextPort)
}


func spawnServer(port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world")
	})

	server := &http.Server{
		Addr:         ":" + port,           // Set the network address (TCP)
		Handler: mux,
	}

	// i would cancel this timer / reset the timer if there is a call made to this server
	time.AfterFunc(10 * time.Minute, func () {
		server.Close()
	}) 

	// blocking function - will keep the goroutine alive for 10 minutes 
	server.ListenAndServe()
}
