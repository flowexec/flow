package store_test

import (
	"testing"

	"github.com/flowexec/flow/v2/pkg/store"
)

func TestRunEnvValue(t *testing.T) {
	t.Run("returns a plain value unchanged", func(t *testing.T) {
		t.Setenv(store.RunClientEnv, "claude-code")
		if got := store.RunEnvValue(store.RunClientEnv); got != "claude-code" {
			t.Errorf("expected %q, got %q", "claude-code", got)
		}
	})

	t.Run("resolves a ${NAME} reference from the environment", func(t *testing.T) {
		// A harness that cannot interpolate its settings file stores the reference verbatim.
		t.Setenv("HARNESS_SESSION_ID", "3f9a-conversation")
		t.Setenv(store.RunSessionEnv, "${HARNESS_SESSION_ID}")
		if got := store.RunEnvValue(store.RunSessionEnv); got != "3f9a-conversation" {
			t.Errorf("expected the referenced value, got %q", got)
		}
	})

	t.Run("yields empty when the reference points at nothing", func(t *testing.T) {
		// Never the literal: a constant `${...}` on every run would group unrelated runs
		// together, which is worse than recording no session at all.
		t.Setenv(store.RunSessionEnv, "${NOT_SET_ANYWHERE}")
		if got := store.RunEnvValue(store.RunSessionEnv); got != "" {
			t.Errorf("expected empty, got %q", got)
		}
	})

	t.Run("leaves values that only look like references alone", func(t *testing.T) {
		for _, value := range []string{"${}", "$NAME", "{NAME}", "prefix-${NAME}", "${NAME}-suffix"} {
			t.Setenv(store.RunSessionEnv, value)
			if got := store.RunEnvValue(store.RunSessionEnv); got != value {
				t.Errorf("expected %q unchanged, got %q", value, got)
			}
		}
	})

	t.Run("is empty when the variable is unset", func(t *testing.T) {
		t.Setenv(store.RunSessionEnv, "")
		if got := store.RunEnvValue(store.RunSessionEnv); got != "" {
			t.Errorf("expected empty, got %q", got)
		}
	})
}
