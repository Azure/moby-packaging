package tests

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"

	"dagger.io/dagger"
)

var (
	client *dagger.Client
)

func TestMain(m *testing.M) {
	distro := flag.String("distro", os.Getenv("DISTRO"), "distro to test")

	flag.Parse()

	fmt.Println(*distro)

	var err error
	client, err = dagger.Connect(context.Background(), dagger.WithWorkdir("../"))
	if err != nil {
		panic(err)
	}

	code := m.Run()
	client.Close()
	os.Exit(code)
}
