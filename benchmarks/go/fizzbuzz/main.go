package main

import (
	"fmt"
	"math"
)

func fizzbuzz(n float64) string {
	res := ""
	for i := 1.0; i <= n; i = i + 1 {
		if math.Mod(i, 3) == 0 && math.Mod(i, 5) == 0 {
			res = res + "FizzBuzz"
		} else if math.Mod(i, 3) == 0 {
			res = res + "Fizz"
		} else if math.Mod(i, 5) == 0 {
			res = res + "Buzz"
		}
	}
	return res
}

func main() {
	fizzbuzz(100000)
	fmt.Println("done")
}
