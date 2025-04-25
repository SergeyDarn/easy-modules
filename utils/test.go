package utils

import (
	"testing"
)

func TestPanic(t *testing.T, testName string, functionToTest func()) {
	t.Helper()

	defer func() { _ = recover() }()

	functionToTest()

	t.Errorf(PrepareDangerOutput("Expected test %s to error."), testName)
}
