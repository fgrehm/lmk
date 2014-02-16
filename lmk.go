package main

import (
	"bufio"
	"fmt"
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

var (
	flagMessage = flag.String("m", "", "")

	defaultMessage = "%s has completed successfully"
)

var usage = `Usage: lmk [options...] command

Options:
  -m  Message to display in case of success, defaults to "[command] has completed successfully"
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}

	flag.Parse()
	if flag.NArg() < 1 {
		usageAndExit("")
	}

	if len(flag.Args()) == 0 {
		log.Fatalf("No command was provided to lmk")
	}

	// https://gobyexample.com/execing-processes
	flagArgs := flag.Args()
	executable, lookErr := exec.LookPath(flagArgs[0])
	if lookErr != nil {
		log.Fatal(lookErr)
	}
	executableArgs := flagArgs[1:]

	log.Printf("Running %s", flagArgs)
	err := run(executable, executableArgs...)

	var icon, msg string
	if err != nil {
		icon = "software-update-urgent"
		msg = fmt.Sprintf("%s has errored!", flagArgs)
	} else {
		icon = "emblem-default"
		if *flagMessage != "" {
			msg = *flagMessage
		} else {
			msg = fmt.Sprintf(defaultMessage, flagArgs)
		}
	}
	go func() {
		for {
			log.Print("Notifying")
			exec.Command("notify-send", "-i", icon, "--", "Heads up!", msg).Run()
			time.Sleep(time.Second * 30)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

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

func usageAndExit(message string) {
	if message != "" {
		fmt.Fprintf(os.Stderr, message)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}
