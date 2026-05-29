package question

import (
	"strings"

	"modular_monolith/internal/shared/bizid"
)

type QuestionUUID string

func (uuid QuestionUUID) String() string { return string(uuid) }

type Type string

const (
	TypeChoice Type = "choice"
	TypeBlank  Type = "blank"
)

type MaterialKind string

const (
	MaterialNone      MaterialKind = ""
	MaterialListening MaterialKind = "listening"
	MaterialReading   MaterialKind = "reading"
)

type Material struct {
	Kind        MaterialKind
	AudioURL    string
	PassageText string
}

type Option struct {
	ID   string
	Text string
}

type Answer struct {
	OptionID string
	Text     string
}

type Question struct {
	uuid            QuestionUUID
	typ             Type
	prompt          string
	options         []Option
	correctOptionID string
	acceptedAnswers []string
	material        Material
}

func NewChoice(prompt string, options []Option, correctOptionID string, material Material) (*Question, error) {
	q := &Question{
		uuid:            QuestionUUID(bizid.New()),
		typ:             TypeChoice,
		prompt:          prompt,
		options:         copyOptions(options),
		correctOptionID: correctOptionID,
		material:        material,
	}
	if !q.valid() {
		return nil, ErrInvalidQuestion
	}
	return q, nil
}

func NewBlank(prompt string, acceptedAnswers []string, material Material) (*Question, error) {
	q := &Question{
		uuid:            QuestionUUID(bizid.New()),
		typ:             TypeBlank,
		prompt:          prompt,
		acceptedAnswers: normalizedAnswers(acceptedAnswers),
		material:        material,
	}
	if !q.valid() {
		return nil, ErrInvalidQuestion
	}
	return q, nil
}

func Rehydrate(uuid QuestionUUID, typ Type, prompt string, options []Option, correctOptionID string, acceptedAnswers []string, material Material) *Question {
	return &Question{
		uuid:            uuid,
		typ:             typ,
		prompt:          prompt,
		options:         copyOptions(options),
		correctOptionID: correctOptionID,
		acceptedAnswers: normalizedAnswers(acceptedAnswers),
		material:        material,
	}
}

func (q *Question) Grade(answer Answer) bool {
	switch q.typ {
	case TypeChoice:
		return answer.OptionID != "" && answer.OptionID == q.correctOptionID
	case TypeBlank:
		got := normalizeAnswer(answer.Text)
		for _, accepted := range q.acceptedAnswers {
			if got == accepted {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func (q *Question) UUID() QuestionUUID      { return q.uuid }
func (q *Question) Type() Type              { return q.typ }
func (q *Question) Prompt() string          { return q.prompt }
func (q *Question) Material() Material      { return q.material }
func (q *Question) Options() []Option       { return copyOptions(q.options) }
func (q *Question) CorrectOptionID() string { return q.correctOptionID }
func (q *Question) AcceptedAnswers() []string {
	return append([]string(nil), q.acceptedAnswers...)
}

func (q *Question) valid() bool {
	if q == nil || q.prompt == "" || !validMaterial(q.material) {
		return false
	}
	switch q.typ {
	case TypeChoice:
		if len(q.options) < 2 || q.correctOptionID == "" {
			return false
		}
		seen := make(map[string]bool, len(q.options))
		correctExists := false
		for _, option := range q.options {
			if option.ID == "" || option.Text == "" || seen[option.ID] {
				return false
			}
			seen[option.ID] = true
			if option.ID == q.correctOptionID {
				correctExists = true
			}
		}
		return correctExists
	case TypeBlank:
		if len(q.acceptedAnswers) == 0 {
			return false
		}
		for _, answer := range q.acceptedAnswers {
			if answer == "" {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func validMaterial(material Material) bool {
	switch material.Kind {
	case MaterialNone:
		return material.AudioURL == "" && material.PassageText == ""
	case MaterialListening:
		return material.AudioURL != ""
	case MaterialReading:
		return material.PassageText != ""
	default:
		return false
	}
}

func copyOptions(options []Option) []Option {
	return append([]Option(nil), options...)
}

func normalizedAnswers(answers []string) []string {
	out := make([]string, 0, len(answers))
	for _, answer := range answers {
		out = append(out, normalizeAnswer(answer))
	}
	return out
}

func normalizeAnswer(answer string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(answer))), " ")
}
