package txerr

import "fmt"

type CommitOutcomeUnknownError struct {
	Op  string
	Err error
}

func (e *CommitOutcomeUnknownError) Error() string {
	return fmt.Sprintf("%s commit outcome unknown: %v", e.Op, e.Err)
}

func (e *CommitOutcomeUnknownError) Unwrap() error {
	return e.Err
}

type PostCommitPublishError struct {
	Op        string
	Err       error
	Committed bool
}

func (e *PostCommitPublishError) Error() string {
	return fmt.Sprintf("%s post-commit publish failed: %v", e.Op, e.Err)
}

func (e *PostCommitPublishError) Unwrap() error {
	return e.Err
}
