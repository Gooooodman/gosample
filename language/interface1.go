package main

import (
	"fmt"
)

type IServer interface {
	Start()
	Stop()
}

type IRestartAble interface {
	Restart()
}

type ServerTypeA struct {
}

func (s *ServerTypeA) Start() {
	fmt.Println("ServerTypeA start")
}

func (s *ServerTypeA) Stop() {
	fmt.Println("ServerTypeA Stop")
}

func (s *ServerTypeA) Restart() {
	fmt.Println("ServerTypeA Restart")
}

type ServerTypeB struct {
}

func (s *ServerTypeB) Start() {
	fmt.Println("ServerTypeB Start")
}

func (s *ServerTypeB) Stop() {
	fmt.Println("ServerTypeB Stop")
}

func main() {
	var sa IServer = new(ServerTypeA)
	sa.Start()
	sa.Stop()

	s1, ok1 := sa.(IRestartAble)
	if ok1 {
		s1.Restart()
	} else {
		fmt.Println("ServerTypeA not Restart")
	}

	var sb IServer = new(ServerTypeB)
	sb.Start()
	sb.Stop()

	s2, ok2 := sb.(IRestartAble)
	if ok2 {
		s2.Restart()
	} else {
		fmt.Println("ServerTypeB not Restart")
	}
}
