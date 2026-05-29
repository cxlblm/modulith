package query

import "context"

type ListQuestions struct{}

type ListQuestionsHandler struct {
	ReadModel ReadModel
}

func (h ListQuestionsHandler) Handle(ctx context.Context, q ListQuestions) ([]QuestionDTO, error) {
	return h.ReadModel.ListQuestions(ctx)
}
