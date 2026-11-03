package commands

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/Autumn-27/ScopeSentry/cmd/scopesentry/output"
)

func TestProjectCreateHelpScopeAndTargets(t *testing.T) {
	fake := &mockClient{}
	var stdout, stderr bytes.Buffer
	code := RunProjectCommand(context.Background(), fake, []string{"create", "--help"}, &stdout, &stderr)
	if code != output.ExitUsageError || stdout.Len() != 0 || len(fake.calledTools) != 0 {
		t.Fatalf("help must remain local on stderr: code=%d stdout=%q calls=%v", code, stdout.String(), fake.calledTools)
	}
	for _, guidance := range []string{
		"actual newline characters (not commas or literal backslash-n)",
		"existing name replaces complete scope",
		"complete successful read",
		"preserve all existing targets",
		"stop if the read is incomplete",
	} {
		if !strings.Contains(stderr.String(), guidance) {
			t.Errorf("missing %q in project create help:\n%s", guidance, stderr.String())
		}
	}
	if strings.Contains(stderr.String(), "multiline or comma-separated") {
		t.Error("project target help must not advertise comma-separated targets")
	}
}
