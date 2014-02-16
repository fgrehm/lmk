package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

func main() {
	if len(os.Args) == 1 {
		log.Fatalf("No command was provided to lmk")
	}

	// https://gobyexample.com/execing-processes
	executable, lookErr := exec.LookPath(os.Args[1])
	if lookErr != nil {
		log.Fatal(lookErr)
	}
	args := os.Args[2:]

	log.Printf("Running %s", os.Args[1:])
	err := run(executable, args...)

	var icon, msg string
	if err != nil {
		icon = "software-update-urgent"
		msg = fmt.Sprintf("That %s has errored!", os.Args[1:])
	} else {
		icon = "emblem-default"
		msg = fmt.Sprintf("That %s has finished", os.Args[1:])
	}
	go func() {
		for {
			log.Print("Notifying")
			exec.Command("notify-send", "-i", icon, "--", "Letting you know...", msg).Run()
			time.Sleep(time.Second * 30)
		}
	}()

	in := bufio.NewReader(os.Stdin)
	_, err = in.ReadString('\n')
	if err != nil {
		panic(err)
	}
}

var run = func(executable string, args ...string) error {
	cmd := exec.Command(executable, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	return cmd.Run()
}
