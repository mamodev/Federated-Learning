package main

import (
	"go-lib/cmd/server/demo"
	"os"
	"strconv"
)

func readArgs() int{
	if len(os.Args) != 2 {
		panic("Usage: server <number of clients>")
	}

	numClients, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic("Error parsing number of clients")
	}

	return numClients
}

func main() {	
	numClients:= readArgs()
	demo.StartDemo(numClients)
}