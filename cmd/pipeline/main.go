package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"dagger.io/dagger"
	"golang.org/x/sys/unix"
)

func main() {
	fmt.Println("vim-go")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, unix.SIGTERM)
	defer cancel()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer client.Close()

	go func() {
		<-ctx.Done()
		client.Close()
	}()
}
