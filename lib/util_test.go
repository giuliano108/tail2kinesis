package lib

import (
	"os"
	"os/exec"
	"testing"
)

func TestParseArgsEmpty(t *testing.T) {
	if os.Getenv("UNDER_TEST") == "1" {
		ParseArgs([]string{})
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestParseArgs")
	cmd.Env = append(os.Environ(), "UNDER_TEST=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("Should've exited, but it didn't (err: %v)", err)
}

func TestParseArgs(t *testing.T) {
	args := ParseArgs([]string{"--region", "dummyregion", "--stream-name", "somestream", "somefile.log"})
	t.Logf("%#v", args)
}
