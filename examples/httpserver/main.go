package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// homeHandler handles the root "/" route.
func homeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Entering homeHandler")
	fmt.Fprintf(w, "Welcome to the traceswrap example server!")
	log.Println("Exiting homeHandler")
}

// helloHandler handles the "/hello" route.
func helloHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Entering helloHandler")

	// Create a channel to receive the asynchronous result.
	resultChan := make(chan string)

	// Spawn a goroutine to simulate asynchronous processing.
	go func() {
		// Simulate some work with a sleep.
		time.Sleep(500 * time.Millisecond)
		resultChan <- "Hello, world!"
	}()

	// Wait for the result and write it to the response.
	result := <-resultChan
	fmt.Fprintf(w, result)

	log.Println("Exiting helloHandler")
}

// timeHandler handles the "/time" route.
func timeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Entering timeHandler")

	resultChan := make(chan string)
	go func() {
		// Simulate a slight delay before fetching the time.
		time.Sleep(250 * time.Millisecond)
		resultChan <- time.Now().Format(time.RFC1123)
	}()

	currentTime := <-resultChan
	fmt.Fprintf(w, "Current server time: %s", currentTime)

	log.Println("Exiting timeHandler")
}

func main() {
	// Register routes with their respective handler functions.
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/time", timeHandler)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server error: ", err)
	}
}
