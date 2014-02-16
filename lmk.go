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

	cmd := flag.Args()

	executable, args := getExecutableAndArgs(cmd)
	log.Printf("Running %s", cmd)
	err := run(executable, args...)

	icon, msg := getIconAndMessage(err, cmd)
	startNotificationLoop(icon, msg)
	waitForEnter()
}

func run (executable string, args ...string) error {
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

func getExecutableAndArgs(cmd []string) (string, []string) {
	if len(cmd) == 0 {
		log.Fatalf("No command was provided to lmk")
	}

	executable, lookErr := exec.LookPath(cmd[0])
	if lookErr != nil {
		log.Fatal(lookErr)
	}
	return executable, cmd[1:]
}

func getIconAndMessage(err error, cmd []string) (icon, msg string) {
	if err != nil {
		icon = "software-update-urgent"
		msg = fmt.Sprintf("%s has errored!", cmd)
	} else {
		icon = "emblem-default"
		if *flagMessage != "" {
			msg = *flagMessage
		} else {
			msg = fmt.Sprintf(defaultMessage, cmd)
		}
	}

	return
}

func startNotificationLoop(icon, msg string) {
	go func() {
		for {
			log.Print("Notifying")
			exec.Command("notify-send", "-i", icon, "--", "Heads up!", msg).Run()
			time.Sleep(time.Second * 30)
		}
	}()
}

func waitForEnter() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	in := bufio.NewReader(os.Stdin)
	_, err := in.ReadString('\n')
	if err != nil {
		panic(err)
	}
}
