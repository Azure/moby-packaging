package tests

import (
	"context"
	"os"
	"testing"

	"dagger.io/dagger"
)

var (
	client *dagger.Client
)

func TestMain(m *testing.M) {
	var err error
	client, err = dagger.Connect(context.Background(), dagger.WithWorkdir("../"))
	if err != nil {
		panic(err)
	}
	code := m.Run()
	client.Close()
	os.Exit(code)
}
