package txerr

import (
	"errors"
	"testing"
)

func TestPostCommitPublishError_UnwrapsCause(t *testing.T) {
	cause := errors.New("event bus unavailable")
	err := &PostCommitPublishError{Op: "order.Repository.Save", Err: cause, Committed: true}

	if !errors.Is(err, cause) {
		t.Fatalf("errors.Is(err, cause) = false, want true")
	}
	if !err.Committed {
		t.Fatal("Committed = false, want true")
	}
}

func TestCommitOutcomeUnknownError_UnwrapsCause(t *testing.T) {
	cause := errors.New("connection reset")
	err := &CommitOutcomeUnknownError{Op: "order.Repository.Save", Err: cause}

	if !errors.Is(err, cause) {
		t.Fatalf("errors.Is(err, cause) = false, want true")
	}
}
