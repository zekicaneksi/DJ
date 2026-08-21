package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	setUpLog()

	os.Exit(m.Run())
}
