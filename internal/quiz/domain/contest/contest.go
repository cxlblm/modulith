package contest

import (
	"strings"
	"time"

	"modular_monolith/internal/shared/bizid"
)

type ContestUUID string

func (uuid ContestUUID) String() string { return string(uuid) }

type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
)

type MaterialSnapshot struct {
	Kind        string
	AudioURL    string
	PassageText string
}

type OptionSnapshot struct {
	ID   string
	Text string
}

type QuestionSnapshot struct {
	QuestionID      string
	Prompt          string
	Type            string
	Options         []OptionSnapshot
	CorrectOptionID string
	AcceptedAnswers []string
	Material        MaterialSnapshot
}

type Answer struct {
	OptionID string
	Text     string
}

type AnswerEvaluation struct {
	Missing bool
	Correct bool
}

type Schedule struct {
	QuestionID string
	StartTime  time.Time
	EndTime    time.Time
}

type Contest struct {
	uuid               ContestUUID
	title              string
	startTime          time.Time
	perQuestionSeconds int
	status             Status
	snapshots          []QuestionSnapshot
}

func NewDraft(title string, startTime time.Time, perQuestionSeconds int, snapshots []QuestionSnapshot) (*Contest, error) {
	if perQuestionSeconds == 0 {
		perQuestionSeconds = 30
	}
	c := &Contest{
		uuid:               ContestUUID(bizid.New()),
		title:              title,
		startTime:          startTime,
		perQuestionSeconds: perQuestionSeconds,
		status:             StatusDraft,
		snapshots:          copySnapshots(snapshots),
	}
	if !c.valid() {
		return nil, ErrInvalidContest
	}
	return c, nil
}

func Rehydrate(uuid ContestUUID, title string, startTime time.Time, perQuestionSeconds int, status Status, snapshots []QuestionSnapshot) *Contest {
	return &Contest{
		uuid:               uuid,
		title:              title,
		startTime:          startTime,
		perQuestionSeconds: perQuestionSeconds,
		status:             status,
		snapshots:          copySnapshots(snapshots),
	}
}

func (c *Contest) Publish() error {
	if !c.valid() {
		return ErrInvalidContest
	}
	c.status = StatusPublished
	c.snapshots = copySnapshots(c.snapshots)
	return nil
}

func (c *Contest) Timeline() []Schedule {
	out := make([]Schedule, 0, len(c.snapshots))
	for i, snapshot := range c.snapshots {
		start := c.startTime.Add(time.Duration(i*c.perQuestionSeconds) * time.Second)
		out = append(out, Schedule{
			QuestionID: snapshot.QuestionID,
			StartTime:  start,
			EndTime:    start.Add(time.Duration(c.perQuestionSeconds) * time.Second),
		})
	}
	return out
}

func (c *Contest) CanSubmitAt(now time.Time, gracePeriod time.Duration) bool {
	return c.IsPublished() && !now.Before(c.startTime) && !now.After(c.EndTime().Add(gracePeriod))
}

func (c *Contest) DueQuestionIDs(now time.Time) map[string]bool {
	due := make(map[string]bool, len(c.snapshots))
	for _, schedule := range c.Timeline() {
		if !now.Before(schedule.StartTime) {
			due[schedule.QuestionID] = true
		}
	}
	return due
}

func (c *Contest) EndTime() time.Time {
	return c.startTime.Add(time.Duration(len(c.snapshots)*c.perQuestionSeconds) * time.Second)
}

func (c *Contest) UUID() ContestUUID             { return c.uuid }
func (c *Contest) Title() string                 { return c.title }
func (c *Contest) StartTime() time.Time          { return c.startTime }
func (c *Contest) PerQuestionSeconds() int       { return c.perQuestionSeconds }
func (c *Contest) Status() Status                { return c.status }
func (c *Contest) Snapshots() []QuestionSnapshot { return copySnapshots(c.snapshots) }
func (c *Contest) IsPublished() bool             { return c.status == StatusPublished }

func (c *Contest) Snapshot(questionID string) (QuestionSnapshot, bool) {
	for _, snapshot := range c.snapshots {
		if snapshot.QuestionID == questionID {
			return copySnapshot(snapshot), true
		}
	}
	return QuestionSnapshot{}, false
}

func (c *Contest) valid() bool {
	if c == nil || c.title == "" || c.startTime.IsZero() || c.perQuestionSeconds <= 0 || len(c.snapshots) == 0 {
		return false
	}
	seen := make(map[string]bool, len(c.snapshots))
	for _, snapshot := range c.snapshots {
		if snapshot.QuestionID == "" || snapshot.Prompt == "" || seen[snapshot.QuestionID] {
			return false
		}
		seen[snapshot.QuestionID] = true
	}
	return true
}

func copySnapshots(snapshots []QuestionSnapshot) []QuestionSnapshot {
	out := make([]QuestionSnapshot, 0, len(snapshots))
	for _, snapshot := range snapshots {
		out = append(out, copySnapshot(snapshot))
	}
	return out
}

func copySnapshot(snapshot QuestionSnapshot) QuestionSnapshot {
	snapshot.Options = append([]OptionSnapshot(nil), snapshot.Options...)
	snapshot.AcceptedAnswers = append([]string(nil), snapshot.AcceptedAnswers...)
	return snapshot
}

func (snapshot QuestionSnapshot) EvaluateAnswer(answer Answer, answerFound bool) AnswerEvaluation {
	if !answerFound || !snapshot.hasUsableAnswer(answer) {
		return AnswerEvaluation{Missing: true}
	}
	return AnswerEvaluation{Correct: snapshot.grade(answer)}
}

func (snapshot QuestionSnapshot) hasUsableAnswer(answer Answer) bool {
	switch snapshot.Type {
	case "choice":
		return answer.OptionID != ""
	case "blank":
		return answer.Text != ""
	default:
		return false
	}
}

func (snapshot QuestionSnapshot) grade(answer Answer) bool {
	switch snapshot.Type {
	case "choice":
		return answer.OptionID != "" && answer.OptionID == snapshot.CorrectOptionID
	case "blank":
		got := normalizeAnswer(answer.Text)
		for _, accepted := range snapshot.AcceptedAnswers {
			if got == normalizeAnswer(accepted) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func normalizeAnswer(answer string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(answer))), " ")
}
