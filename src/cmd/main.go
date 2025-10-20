package main

import (
	"fmt"
	"os"
	"register"
	"strconv"
)

func main() {
	config := register.CreateConfig("")
	config.Init()

	if len(os.Args) > 1 && len(os.Args[1]) > 0 {
		config.AdminUsername = os.Args[1]
	}
	if len(os.Args) > 2 && len(os.Args[2]) > 0 {
		config.AdminPassword = os.Args[2]
	}
	if len(os.Args) > 3 && len(os.Args[3]) > 0 {
		config.Hostport = os.Args[3]
	}
	if len(os.Args) > 4 && len(os.Args[4]) > 0 {
		config.Domain = os.Args[4]
	}
	if len(os.Args) > 5 && len(os.Args[5]) > 0 {
		config.Static = os.Args[5]
	}
	if len(os.Args) > 6 && len(os.Args[6]) > 0 {
		defaultmax, err := strconv.ParseInt(os.Args[6], 10, 64)
		if err != nil {
			fmt.Println("failed to convert ", os.Args[6], " to number")
		} else {
			config.DefaultMax = defaultmax
		}
	}
	if len(os.Args) > 7 && len(os.Args[7]) > 0 {
		stopregistration, err := strconv.ParseBool(os.Args[7])
		if err != nil {
			fmt.Println("failed to convert ", os.Args[7], " to bool")
		} else {
			config.StopRegistration = stopregistration
		}
	}

	if len(os.Args) > 1 {
		if err := config.WriteConfigDetails(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	register := register.CreateRegisterIlugc(config)
	if err := register.Init(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	register.Run()
}
