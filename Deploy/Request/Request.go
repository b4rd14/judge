package main

import (
	"GO/Judge/Requester"
)

func main() {
	err := requester.Request()
	if err != nil {
		return
	}
}
