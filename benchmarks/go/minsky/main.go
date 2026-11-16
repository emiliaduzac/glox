package main

import "fmt"

func minskyEven(n float64) float64 {
	r0 := n
	pc := 0
	halt := false

	for !halt {
		switch pc {
		case 0:
			if r0 == 0 {
				pc = 2
			} else {
				r0 = r0 - 1
				pc = 1
			}
		case 1:
			if r0 == 0 {
				pc = 3
			} else {
				r0 = r0 - 1
				pc = 0
			}
		case 2:
			r0 = r0 + 1
			pc = 3
		case 3:
			halt = true
		}
	}

	return r0
}

func main() {
	fmt.Println(minskyEven(200000))
}
