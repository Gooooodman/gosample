package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	fmt.Println("Welcome to the playground!")

	fmt.Println("The time is", time.Now())

	fmt.Println("And if you try to open a file:")
	fmt.Println(os.Open("D:\\WorspaceEugene\\gopath\\src\\github.com\\fengbaoxp\\gosample\\tour\\tour2.go"))

	fmt.Println("Or access the network:")
	fmt.Println(net.Dial("tcp", "google.com"))
}
