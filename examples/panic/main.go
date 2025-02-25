package main

import (
	"fmt"
)

// main is the entry point of the application.
// It prints an introductory message, runs the demo function that triggers a panic,
// and then, after recovering, waits briefly before self-terminating.
func main() {
	fmt.Println("=== Panic Recovery Demo Application ===")

	// Run the demo function that will panic and recover.
	demoPanicRecovery()

	// Continue execution after panic recovery.
	fmt.Println("Continuing execution after recovery.")
}

// demoPanicRecovery demonstrates panic triggering and recovery.
// It defers a function to recover from a panic, prints a message upon recovery,
// and then returns so that main can continue.
func demoPanicRecovery() {
	fmt.Println("Starting panic recovery test...")

	defer func() {
		if r := recover(); r != nil {
			// Print the recovered panic value.
			fmt.Printf("Recovered from panic: %v\n", r)
		}
	}()

	// This function will intentionally panic.
	triggerPanic()

	// This line should not be executed.
	fmt.Println("This message should never be printed.")
}

// triggerPanic deliberately triggers a panic with a custom message.
func triggerPanic() {
	fmt.Println("Triggering a panic now...")
	panic("Intentional panic for demonstration purposes")
}
