package mysql

import "time"

type QuestionModel struct {
	ID                  uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned"`
	UUID                string `gorm:"type:char(36);not null;uniqueIndex"`
	Type                string `gorm:"size:32;not null;index"`
	Prompt              string `gorm:"type:text;not null"`
	MaterialKind        string `gorm:"size:32"`
	AudioURL            string `gorm:"size:1024"`
	PassageText         string `gorm:"type:text"`
	CorrectOptionID     string `gorm:"size:64"`
	AcceptedAnswersJSON string `gorm:"type:text;not null"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type QuestionOptionModel struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned"`
	QuestionUUID string `gorm:"type:char(36);not null;index"`
	OptionID     string `gorm:"size:64;not null"`
	Text         string `gorm:"size:512;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ContestModel struct {
	ID                 uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned"`
	UUID               string `gorm:"type:char(36);not null;uniqueIndex"`
	Title              string `gorm:"size:255;not null"`
	StartTime          time.Time
	PerQuestionSeconds int    `gorm:"not null"`
	Status             string `gorm:"size:32;not null;index"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ContestQuestionModel struct {
	ID                  uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned"`
	ContestUUID         string `gorm:"type:char(36);not null;index"`
	Position            int    `gorm:"not null"`
	QuestionUUID        string `gorm:"type:char(36);not null;index"`
	Prompt              string `gorm:"type:text;not null"`
	Type                string `gorm:"size:32;not null"`
	MaterialKind        string `gorm:"size:32"`
	AudioURL            string `gorm:"size:1024"`
	PassageText         string `gorm:"type:text"`
	CorrectOptionID     string `gorm:"size:64"`
	AcceptedAnswersJSON string `gorm:"type:text;not null"`
	OptionsJSON         string `gorm:"type:text;not null"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type ParticipationModel struct {
	ID            uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned"`
	UUID          string `gorm:"type:char(36);not null;uniqueIndex"`
	ContestUUID   string `gorm:"type:char(36);not null;uniqueIndex:idx_participation_contest_user"`
	UserUUID      string `gorm:"type:char(36);not null;uniqueIndex:idx_participation_contest_user"`
	QuestionsJSON string `gorm:"type:text;not null"`
	RevivalCards  int    `gorm:"not null"`
	Status        string `gorm:"size:32;not null;index"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ParticipationAnswerModel struct {
	ID                uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned"`
	ParticipationUUID string `gorm:"type:char(36);not null;index"`
	QuestionUUID      string `gorm:"type:char(36);not null;index"`
	Correct           bool   `gorm:"not null"`
	UsedRevival       bool   `gorm:"not null"`
	Passed            bool   `gorm:"not null"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type RevivalCardModel struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned"`
	UserUUID  string `gorm:"type:char(36);not null;uniqueIndex"`
	Balance   int    `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func Models() []any {
	return []any{
		&QuestionModel{},
		&QuestionOptionModel{},
		&ContestModel{},
		&ContestQuestionModel{},
		&ParticipationModel{},
		&ParticipationAnswerModel{},
		&RevivalCardModel{},
	}
}
