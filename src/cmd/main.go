package main

import (
	"fmt"
	"os"
	"register_ilugc"
)

func main() {
	address := ""
	if len(os.Args) > 1 {
		address = os.Args[1]
	}
	register := register_ilugc.CreateRegisterIlugc(address)
	if err := register.Init(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	register.Run()
}
