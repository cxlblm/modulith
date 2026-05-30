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

func TestContestCanSubmitAtHonorsPublicationWindowAndGracePeriod(t *testing.T) {
	start := time.Date(2026, 6, 1, 20, 0, 0, 0, time.UTC)
	c, err := NewDraft("daily quiz", start, 30, []QuestionSnapshot{
		{QuestionID: "q1", Prompt: "first"},
		{QuestionID: "q2", Prompt: "second"},
	})
	if err != nil {
		t.Fatalf("NewDraft() error = %v", err)
	}
	if c.CanSubmitAt(start, 30*time.Second) {
		t.Fatal("CanSubmitAt(draft) = true, want false")
	}
	if err := c.Publish(); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	tests := []struct {
		name string
		now  time.Time
		want bool
	}{
		{name: "before start", now: start.Add(-time.Nanosecond), want: false},
		{name: "at start", now: start, want: true},
		{name: "at end plus grace", now: start.Add(90 * time.Second), want: true},
		{name: "after grace", now: start.Add(90*time.Second + time.Nanosecond), want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.CanSubmitAt(tt.now, 30*time.Second); got != tt.want {
				t.Fatalf("CanSubmitAt(%s) = %t, want %t", tt.now, got, tt.want)
			}
		})
	}
}

func TestContestDueQuestionIDs(t *testing.T) {
	start := time.Date(2026, 6, 1, 20, 0, 0, 0, time.UTC)
	c, err := NewDraft("daily quiz", start, 30, []QuestionSnapshot{
		{QuestionID: "q1", Prompt: "first"},
		{QuestionID: "q2", Prompt: "second"},
		{QuestionID: "q3", Prompt: "third"},
	})
	if err != nil {
		t.Fatalf("NewDraft() error = %v", err)
	}

	due := c.DueQuestionIDs(start.Add(30 * time.Second))
	if !due["q1"] || !due["q2"] {
		t.Fatalf("DueQuestionIDs() = %+v, want q1 and q2 due", due)
	}
	if due["q3"] {
		t.Fatalf("DueQuestionIDs() = %+v, want q3 not due", due)
	}
}

func TestQuestionSnapshotEvaluateAnswer(t *testing.T) {
	tests := []struct {
		name        string
		snapshot    QuestionSnapshot
		answer      Answer
		answerFound bool
		want        AnswerEvaluation
	}{
		{
			name: "choice correct",
			snapshot: QuestionSnapshot{
				QuestionID:      "q1",
				Prompt:          "choose",
				Type:            "choice",
				CorrectOptionID: "b",
			},
			answer:      Answer{OptionID: "b"},
			answerFound: true,
			want:        AnswerEvaluation{Correct: true},
		},
		{
			name: "choice missing option",
			snapshot: QuestionSnapshot{
				QuestionID:      "q1",
				Prompt:          "choose",
				Type:            "choice",
				CorrectOptionID: "b",
			},
			answerFound: true,
			want:        AnswerEvaluation{Missing: true},
		},
		{
			name: "blank normalizes answer",
			snapshot: QuestionSnapshot{
				QuestionID:      "q2",
				Prompt:          "blank",
				Type:            "blank",
				AcceptedAnswers: []string{"hello world"},
			},
			answer:      Answer{Text: " Hello   World "},
			answerFound: true,
			want:        AnswerEvaluation{Correct: true},
		},
		{
			name: "not found is missing",
			snapshot: QuestionSnapshot{
				QuestionID:      "q2",
				Prompt:          "blank",
				Type:            "blank",
				AcceptedAnswers: []string{"hello world"},
			},
			answerFound: false,
			want:        AnswerEvaluation{Missing: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.snapshot.EvaluateAnswer(tt.answer, tt.answerFound); got != tt.want {
				t.Fatalf("EvaluateAnswer() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
