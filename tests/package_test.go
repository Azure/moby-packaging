package tests

import (
	"fmt"
	"testing"
)

func TestThis(t *testing.T) {
	t.Run("lol", func(t *testing.T) {
		fmt.Printf("t: %v\n", t)

		// t.FailNow()
	})
}
