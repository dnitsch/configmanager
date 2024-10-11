package cmd_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	cmd "github.com/dnitsch/configmanager/cmd/configmanager"
)

type cmdTestInput struct {
	args        []string
	errored     bool
	exactOutput string
	output      []string
}

func cmdRunTestHelper(t *testing.T, testInput *cmdTestInput) {
	t.Helper()

	logOut := &bytes.Buffer{}
	logErr := &bytes.Buffer{}

	cmd := cmd.NewRootCmd(logOut, logErr)
	os.Args = append([]string{os.Args[0]}, testInput.args...)
	errOut := &bytes.Buffer{}
	stdOut := &bytes.Buffer{}
	cmd.Cmd.SetArgs(testInput.args)
	cmd.Cmd.SetErr(errOut)
	cmd.Cmd.SetOut(stdOut)

	if err := cmd.Execute(context.TODO()); err != nil {
		if testInput.errored {
			return
		}
		t.Fatalf("\ngot: %v\nwanted <nil>\n", err)
	}

	if testInput.errored && errOut.Len() < 1 {
		t.Errorf("\ngot: nil\nwanted an error to be thrown")
	}
	if len(testInput.output) > 0 {
		for _, v := range testInput.output {
			if !strings.Contains(logOut.String(), v) {
				t.Errorf("\ngot: %s\vnot found in: %v", logOut.String(), v)
			}
		}
	}
	if testInput.exactOutput != "" && logOut.String() != testInput.exactOutput {
		t.Errorf("output mismatch\ngot: %s\n\nwanted: %s", logOut.String(), testInput.exactOutput)
	}
}
