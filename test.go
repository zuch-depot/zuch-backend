package main

import "fmt"

func test() {
	a := neighbourTracks(4, 4, 4)
	for i := range a {
		fmt.Println(a[i])
	}
}
