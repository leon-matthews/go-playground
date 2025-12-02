package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type config struct {
	tick time.Duration
}

func (c *config) load() {
	c.tick = time.Second * 10
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	cfg := &config{}

	signalStream := make(chan os.Signal, 1)
	signal.Notify(signalStream, os.Interrupt, syscall.SIGHUP)

	defer func() {
		signal.Stop(signalStream)
		cancel()
	}()

	go func() {
		for {
			select {
			case s := <-signalStream:
				switch s {
				case syscall.SIGHUP:
					log.Println("SIGHUP received, reloading configuration")
					cfg.load()
				case os.Interrupt:
					log.Println("Interrupt signal received")
					cancel()
					os.Exit(1)
				}
			case <-ctx.Done():
				log.Println("Context done")
				os.Exit(1)
			}
		}
	}()

	if err := run(ctx, cfg, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg *config, out io.Writer) error {
	cfg.load()
	log.SetOutput(out)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.Tick(cfg.tick):
			log.Println("Still running")
		}
	}
}
