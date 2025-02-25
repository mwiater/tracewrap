package main

import (
	"fmt"
)

// add returns the sum of two integers.
func add(a, b int) int {
	return a + b
}

// subtract returns the difference between two integers.
func subtract(a, b int) int {
	return a - b
}

// multiply returns the product of two integers.
func multiply(a, b int) int {
	return a * b
}

// divide returns the quotient of two integers. If the divisor is zero, an error is returned.
func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("cannot divide by zero")
	}
	return a / b, nil
}

func main() {
	fmt.Println("Simple Math Sample Project")
	a, b := 10, 5

	fmt.Printf("%d + %d = %d\n", a, b, add(a, b))
	fmt.Printf("%d - %d = %d\n", a, b, subtract(a, b))
	fmt.Printf("%d * %d = %d\n", a, b, multiply(a, b))

	result, err := divide(a, b)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("%d / %d = %d\n", a, b, result)
	}
}
