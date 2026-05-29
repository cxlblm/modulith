package contest

import (
	"testing"
	"time"
)

func TestContestPublishFreezesQuestionSnapshotsAndBuildsTimeline(t *testing.T) {
	start := time.Date(2026, 6, 1, 20, 0, 0, 0, time.UTC)
	snapshots := []QuestionSnapshot{
		{QuestionID: "q1", Prompt: "first"},
		{QuestionID: "q2", Prompt: "second"},
	}
	c, err := NewDraft("daily quiz", start, 30, snapshots)
	if err != nil {
		t.Fatalf("NewDraft() error = %v", err)
	}

	if err := c.Publish(); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	snapshots[0].Prompt = "changed after publish"

	if c.Status() != StatusPublished {
		t.Fatalf("Status() = %q, want %q", c.Status(), StatusPublished)
	}
	if c.EndTime() != start.Add(time.Minute) {
		t.Fatalf("EndTime() = %s, want %s", c.EndTime(), start.Add(time.Minute))
	}
	if got := c.Snapshots()[0].Prompt; got != "first" {
		t.Fatalf("Snapshots()[0].Prompt = %q, want frozen prompt", got)
	}

	timeline := c.Timeline()
	if len(timeline) != 2 {
		t.Fatalf("Timeline() length = %d, want 2", len(timeline))
	}
	if timeline[1].StartTime != start.Add(30*time.Second) || timeline[1].EndTime != start.Add(time.Minute) {
		t.Fatalf("second schedule = %s-%s, want %s-%s", timeline[1].StartTime, timeline[1].EndTime, start.Add(30*time.Second), start.Add(time.Minute))
	}
}
