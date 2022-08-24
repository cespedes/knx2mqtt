package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	verbose := false
	flag.BoolVar(&verbose, "v", false, "add verbosity")
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Println("Must use 1 arg.")
		os.Exit(1)
	}
	filename := flag.Args()[0]
	fmt.Println("File:", filename)

	ets, err := Uncompress(filename)
	if err != nil {
		panic(err)
	}

	k, err := ParseProject(ets.Project)
	if err != nil {
		panic(err)
	}

	PrintProject(k, verbose)
}
