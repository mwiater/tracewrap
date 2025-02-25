package main

import (
	"fmt"
)

// fibonacci calculates the nth Fibonacci number recursively.
func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func main() {
	fmt.Println("Recursive Fibonacci Example Project")
	n := 10
	result := fibonacci(n)
	fmt.Printf("Fibonacci(%d) = %d\n", n, result)
}
