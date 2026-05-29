package module

type QuizModule interface{}

type quizModule struct{}

func NewQuizModule() QuizModule {
	return &quizModule{}
}
