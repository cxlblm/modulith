package contest

import (
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
