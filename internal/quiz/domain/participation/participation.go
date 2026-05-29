package participation

import "modular_monolith/internal/shared/bizid"

type ParticipationUUID string

func (uuid ParticipationUUID) String() string { return string(uuid) }

type Status string

const (
	StatusActive     Status = "active"
	StatusEliminated Status = "eliminated"
	StatusPassed     Status = "passed"
)

type QuestionRef struct {
	ID string
}

type AnswerRecord struct {
	QuestionID  string
	Correct     bool
	UsedRevival bool
	Passed      bool
}

type AnswerOutcome struct {
	QuestionID  string
	Correct     bool
	UsedRevival bool
	Status      Status
}

type Participation struct {
	uuid         ParticipationUUID
	contestID    string
	userID       string
	questions    []QuestionRef
	revivalCards int
	status       Status
	answers      []AnswerRecord
}

func New(contestID string, userID string, questions []QuestionRef, revivalCards int) (*Participation, error) {
	if contestID == "" || userID == "" || len(questions) == 0 || revivalCards < 0 {
		return nil, ErrInvalidParticipation
	}
	for _, ref := range questions {
		if ref.ID == "" {
			return nil, ErrInvalidParticipation
		}
	}
	return &Participation{
		uuid:         ParticipationUUID(bizid.New()),
		contestID:    contestID,
		userID:       userID,
		questions:    append([]QuestionRef(nil), questions...),
		revivalCards: revivalCards,
		status:       StatusActive,
	}, nil
}

func Rehydrate(uuid ParticipationUUID, contestID string, userID string, questions []QuestionRef, revivalCards int, status Status, answers []AnswerRecord) *Participation {
	return &Participation{
		uuid:         uuid,
		contestID:    contestID,
		userID:       userID,
		questions:    append([]QuestionRef(nil), questions...),
		revivalCards: revivalCards,
		status:       status,
		answers:      append([]AnswerRecord(nil), answers...),
	}
}

func (p *Participation) Submit(questionID string, correct bool) (AnswerOutcome, error) {
	if p.status == StatusEliminated {
		return AnswerOutcome{}, ErrParticipantEliminated
	}
	if p.status == StatusPassed {
		return AnswerOutcome{}, ErrQuestionAlreadyAnswered
	}
	if !p.hasQuestion(questionID) {
		return AnswerOutcome{}, ErrInvalidParticipation
	}
	if p.hasAnswered(questionID) {
		return AnswerOutcome{}, ErrQuestionAlreadyAnswered
	}
	record := AnswerRecord{QuestionID: questionID, Correct: correct, Passed: correct}
	if !correct {
		if p.revivalCards <= 0 {
			p.status = StatusEliminated
			p.answers = append(p.answers, record)
			return AnswerOutcome{QuestionID: questionID, Correct: false, Status: p.status}, nil
		}
		p.revivalCards--
		record.UsedRevival = true
		record.Passed = true
	}
	p.answers = append(p.answers, record)
	if p.allPassed() {
		p.status = StatusPassed
	}
	return AnswerOutcome{QuestionID: questionID, Correct: correct, UsedRevival: record.UsedRevival, Status: p.status}, nil
}

func (p *Participation) UUID() ParticipationUUID { return p.uuid }
func (p *Participation) ContestID() string       { return p.contestID }
func (p *Participation) UserID() string          { return p.userID }
func (p *Participation) Status() Status          { return p.status }
func (p *Participation) RevivalCards() int       { return p.revivalCards }
func (p *Participation) Questions() []QuestionRef {
	return append([]QuestionRef(nil), p.questions...)
}
func (p *Participation) Answers() []AnswerRecord {
	return append([]AnswerRecord(nil), p.answers...)
}

func (p *Participation) hasQuestion(questionID string) bool {
	for _, question := range p.questions {
		if question.ID == questionID {
			return true
		}
	}
	return false
}

func (p *Participation) hasAnswered(questionID string) bool {
	for _, answer := range p.answers {
		if answer.QuestionID == questionID {
			return true
		}
	}
	return false
}

func (p *Participation) allPassed() bool {
	if len(p.answers) < len(p.questions) {
		return false
	}
	passedByQuestion := make(map[string]bool, len(p.answers))
	for _, answer := range p.answers {
		if answer.Passed {
			passedByQuestion[answer.QuestionID] = true
		}
	}
	for _, question := range p.questions {
		if !passedByQuestion[question.ID] {
			return false
		}
	}
	return true
}
