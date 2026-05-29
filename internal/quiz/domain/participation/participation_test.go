package participation

import "testing"

func TestParticipationRecordsUsedRevivalAndPasses(t *testing.T) {
	p, err := New("contest-1", "user-1", []QuestionRef{{ID: "q1"}, {ID: "q2"}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	first, err := p.Submit("q1", true, false)
	if err != nil {
		t.Fatalf("Submit(correct) error = %v", err)
	}
	if first.Status != StatusActive || first.UsedRevival {
		t.Fatalf("first outcome = %+v, want active without revival", first)
	}

	second, err := p.Submit("q2", false, true)
	if err != nil {
		t.Fatalf("Submit(wrong with revival) error = %v", err)
	}
	if second.Status != StatusPassed || !second.UsedRevival {
		t.Fatalf("second outcome = %+v, want passed with revival", second)
	}
}

func TestParticipationEliminatesWithoutRevivalCard(t *testing.T) {
	p, err := New("contest-1", "user-1", []QuestionRef{{ID: "q1"}, {ID: "q2"}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	outcome, err := p.Submit("q1", false, false)
	if err != nil {
		t.Fatalf("Submit(wrong) error = %v", err)
	}
	if outcome.Status != StatusEliminated {
		t.Fatalf("Status = %q, want %q", outcome.Status, StatusEliminated)
	}
	if _, err := p.Submit("q2", true, false); err == nil {
		t.Fatal("Submit(after eliminated) error = nil, want error")
	}
}

func TestParticipationCanSubmitRejectsInvalidOrAnsweredQuestion(t *testing.T) {
	p, err := New("contest-1", "user-1", []QuestionRef{{ID: "q1"}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := p.CanSubmit("missing"); err == nil {
		t.Fatal("CanSubmit(missing) error = nil, want error")
	}
	if _, err := p.Submit("q1", true, false); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if err := p.CanSubmit("q1"); err == nil {
		t.Fatal("CanSubmit(answered) error = nil, want error")
	}
}

func TestParticipationHasAnswered(t *testing.T) {
	p, err := New("contest-1", "user-1", []QuestionRef{{ID: "q1"}, {ID: "q2"}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if p.HasAnswered("q1") {
		t.Fatal("HasAnswered(q1) = true, want false")
	}
	if _, err := p.Submit("q1", true, false); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if !p.HasAnswered("q1") {
		t.Fatal("HasAnswered(q1) = false, want true")
	}
}

func TestParticipationSubmitAfterTerminalStatusDoesNotAppendAnswer(t *testing.T) {
	p, err := New("contest-1", "user-1", []QuestionRef{{ID: "q1"}, {ID: "q2"}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := p.Submit("q1", false, false); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	if _, err := p.Submit("q2", true, false); err == nil {
		t.Fatal("Submit(after eliminated) error = nil, want error")
	}
	if got := len(p.Answers()); got != 1 {
		t.Fatalf("answers = %d, want 1", got)
	}
}
