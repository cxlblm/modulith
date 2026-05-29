package participation

import "testing"

func TestParticipationUsesRevivalCardAndPasses(t *testing.T) {
	p, err := New("contest-1", "user-1", []QuestionRef{{ID: "q1"}, {ID: "q2"}}, 1)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	first, err := p.Submit("q1", true)
	if err != nil {
		t.Fatalf("Submit(correct) error = %v", err)
	}
	if first.Status != StatusActive || first.UsedRevival {
		t.Fatalf("first outcome = %+v, want active without revival", first)
	}

	second, err := p.Submit("q2", false)
	if err != nil {
		t.Fatalf("Submit(wrong with revival) error = %v", err)
	}
	if second.Status != StatusPassed || !second.UsedRevival {
		t.Fatalf("second outcome = %+v, want passed with revival", second)
	}
	if p.RevivalCards() != 0 {
		t.Fatalf("RevivalCards() = %d, want 0", p.RevivalCards())
	}
}

func TestParticipationEliminatesWithoutRevivalCard(t *testing.T) {
	p, err := New("contest-1", "user-1", []QuestionRef{{ID: "q1"}, {ID: "q2"}}, 0)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	outcome, err := p.Submit("q1", false)
	if err != nil {
		t.Fatalf("Submit(wrong) error = %v", err)
	}
	if outcome.Status != StatusEliminated {
		t.Fatalf("Status = %q, want %q", outcome.Status, StatusEliminated)
	}
	if _, err := p.Submit("q2", true); err == nil {
		t.Fatal("Submit(after eliminated) error = nil, want error")
	}
}
