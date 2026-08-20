package main

import (
	"fmt"
	"net/http"
	"dashboard/internal/handlers"
)


func main(){
	mux := http.NewServeMux()

	mux.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		println("hit /")
		http.ServeFile(w, r, "index.html")
	})

	mux.HandleFunc("/spawn", handlers.HandleSpawn)

	mux.HandleFunc("/health", handlers.HandleHealth)

	mux.HandleFunc("/blast", handlers.HandleBlast)

	// actual binding call 
	if err := http.ListenAndServe(":3000", mux); err != nil {
		fmt.Println("server failed: ", err)
	}
}
