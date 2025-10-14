package main

import (
	"fmt"
	"os"
	"register_ilugc"
)

func main() {
	hostport := ""
	static := ""

	if len(os.Args) > 1 {
		hostport = os.Args[1]
	}
	if len(os.Args) > 2 {
		static = os.Args[2]
	}
	register := register_ilugc.CreateRegisterIlugc(hostport, static)
	if err := register.Init(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	register.Run()
}
