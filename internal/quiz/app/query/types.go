package query

import "time"

type QuestionDTO struct {
	ID        string              `json:"id"`
	Type      string              `json:"type"`
	Prompt    string              `json:"prompt"`
	Options   []QuestionOptionDTO `json:"options,omitempty"`
	Material  MaterialDTO         `json:"material"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type QuestionOptionDTO struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type MaterialDTO struct {
	Kind        string `json:"kind,omitempty"`
	AudioURL    string `json:"audio_url,omitempty"`
	PassageText string `json:"passage_text,omitempty"`
}

type ArenaDTO struct {
	ContestID          string             `json:"contest_id"`
	Title              string             `json:"title"`
	Status             string             `json:"status"`
	StartTime          time.Time          `json:"start_time"`
	EndTime            time.Time          `json:"end_time"`
	PerQuestionSeconds int                `json:"per_question_seconds"`
	Questions          []ArenaQuestionDTO `json:"questions"`
}

type ArenaQuestionDTO struct {
	ID        string              `json:"id"`
	Type      string              `json:"type"`
	Prompt    string              `json:"prompt"`
	Options   []QuestionOptionDTO `json:"options,omitempty"`
	Material  MaterialDTO         `json:"material"`
	StartTime time.Time           `json:"start_time"`
	EndTime   time.Time           `json:"end_time"`
}
