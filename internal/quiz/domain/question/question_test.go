package question

import "testing"

func TestQuestionGradesChoiceAndBlankAnswers(t *testing.T) {
	choice, err := NewChoice("listen and choose", []Option{
		{ID: "a", Text: "Alpha"},
		{ID: "b", Text: "Bravo"},
	}, "b", Material{Kind: MaterialListening, AudioURL: "https://example.test/audio.mp3"})
	if err != nil {
		t.Fatalf("NewChoice() error = %v", err)
	}
	if !choice.Grade(Answer{OptionID: "b"}) {
		t.Fatal("choice.Grade(correct) = false, want true")
	}
	if choice.Grade(Answer{OptionID: "a"}) {
		t.Fatal("choice.Grade(wrong) = true, want false")
	}

	blank, err := NewBlank("fill from reading", []string{"Hello World"}, Material{Kind: MaterialReading, PassageText: "A short passage"})
	if err != nil {
		t.Fatalf("NewBlank() error = %v", err)
	}
	if !blank.Grade(Answer{Text: " hello   world "}) {
		t.Fatal("blank.Grade(normalized answer) = false, want true")
	}
	if blank.Grade(Answer{Text: "goodbye"}) {
		t.Fatal("blank.Grade(wrong) = true, want false")
	}
}

func TestNewQuestionValidatesRequiredFields(t *testing.T) {
	if _, err := NewChoice("", []Option{{ID: "a", Text: "Alpha"}}, "a", Material{}); err == nil {
		t.Fatal("NewChoice(empty prompt) error = nil, want error")
	}
	if _, err := NewChoice("choose", []Option{{ID: "a", Text: "Alpha"}}, "missing", Material{}); err == nil {
		t.Fatal("NewChoice(missing correct option) error = nil, want error")
	}
	if _, err := NewBlank("blank", []string{""}, Material{}); err == nil {
		t.Fatal("NewBlank(empty answer) error = nil, want error")
	}
	if _, err := NewBlank("listen", []string{"answer"}, Material{Kind: MaterialListening}); err == nil {
		t.Fatal("NewBlank(listening without audio URL) error = nil, want error")
	}
}
