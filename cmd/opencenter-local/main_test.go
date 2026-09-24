package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRustFSCLIUsesCredentialFileInsteadOfPlaintextFlags(t *testing.T) {
	root := newRootCmd()
	var output bytes.Buffer
	root.SetOut(&output)
	root.SetErr(&output)
	root.SetArgs([]string{"rustfs", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	if !strings.Contains(text, "--credentials-file") {
		t.Fatalf("help omitted safe credential file option: %s", text)
	}
	for _, forbidden := range []string{"--access-key", "--secret-key"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("help exposes plaintext credential flag %q: %s", forbidden, text)
		}
	}
}
