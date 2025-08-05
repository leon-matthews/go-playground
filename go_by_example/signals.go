package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Signal notification sends os.Signal values on a channel.
	// Create a (buffered) channel to receive these notifications.
	sigs := make(chan os.Signal, 1)

	// Pass no filters to receive all signals
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	done := make(chan bool)

	// Wait for first recieved signal
	go func() {
		sig := <-sigs
		fmt.Println("Received:", sig.String())
		done <- true
	}()

	fmt.Println("Awaiting signal")

	// Wait for goroutine to finish
	<-done

	fmt.Println("Exiting normally")
}
