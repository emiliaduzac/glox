package main

import "fmt"

func fib(n float64) float64 {
	if n <= 1 {
		return n
	}
	return fib(n-2) + fib(n-1)
}

func main() {
	fmt.Println(fib(30))
}
