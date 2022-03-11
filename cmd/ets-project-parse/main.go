package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Must use 1 arg.")
		os.Exit(1)
	}
	fmt.Println("File:", os.Args[1])

	ets, err := Uncompress(os.Args[1])
	if err != nil {
		panic(err)
	}

	k, err := ParseProject(ets.Project)
	if err != nil {
		panic(err)
	}

	PrintProject(k)
}
