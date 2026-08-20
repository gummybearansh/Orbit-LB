package cluster

import (
	"time"
	"sync"
	"strconv"
	"net/http"
	"fmt"
)

var (
	nextPort = 8000
	portMutex sync.Mutex
)


func GetPort() string {
	portMutex.Lock()
	defer portMutex.Unlock()

	nextPort += 1 
	return strconv.Itoa(nextPort)
}


func SpawnServer(port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world")
	})

	server := &http.Server{
		Addr:         ":" + port,           // Set the network address (TCP)
		Handler: mux,
	}

	// i would cancel this timer / reset the timer if there is a call made to this server
	time.AfterFunc(1 * time.Minute, func () {
		server.Close()
	}) 

	// blocking function - will keep the goroutine alive for 10 minutes 
	server.ListenAndServe()
}
