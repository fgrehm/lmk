package main

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"time"
)

func main() {
	log.Print("Will run CMD")
	runCmd()

	go func() {
		for {
			log.Print("Will notify...")
			exec.Command("notify-send", "Letting you know...", "That the command has finished").Run()
			time.Sleep(time.Minute * 1)
		}
	}()

	log.Print("Will wait for input")

	in := bufio.NewReader(os.Stdin)
	_, err := in.ReadString('\n')
	if err != nil {
		panic(err)
	}

	log.Print("Leaving...")
	// FIXME: Do we need to "kill" the goroutine? (NO)
	// FIXME: Trap Ctrl+C
}

func runCmd() {
	// https://gobyexample.com/execing-processes
	binary, lookErr := exec.LookPath(os.Args[1])
	if lookErr != nil {
		panic(lookErr)
	}
	args := os.Args[2:]
	// FIXME: Should handle errors and Ctrl+C here
	//
	system(binary, args...)
}

func system(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Run()
}
