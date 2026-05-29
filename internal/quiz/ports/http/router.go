package http

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"

	"modular_monolith/internal/platform/httpserver"
	"modular_monolith/internal/quiz/app"
	"modular_monolith/internal/quiz/app/command"
	"modular_monolith/internal/quiz/app/query"
)

type Handler struct {
	app *app.Application
}

func Register(e *echo.Echo, app *app.Application) {
	h := &Handler{app: app}
	e.POST("/quiz/questions", h.createQuestion)
	e.GET("/quiz/questions", h.listQuestions)
	e.POST("/quiz/contests", h.createContest)
	e.POST("/quiz/contests/:contest_id/publish", h.publishContest)
	e.GET("/quiz/contests/:contest_id/arena", h.getArena)

	g := e.Group("/quiz/contests/:contest_id", httpserver.RequireUserAuth())
	g.POST("/answers", h.submitAnswer)
	g.POST("/reward/claim", h.claimReward)
}

func (h *Handler) createQuestion(c *echo.Context) error {
	var req struct {
		Type    string `json:"type" validate:"required"`
		Prompt  string `json:"prompt" validate:"required"`
		Options []struct {
			ID   string `json:"id" validate:"required"`
			Text string `json:"text" validate:"required"`
		} `json:"options"`
		CorrectOptionID string   `json:"correct_option_id"`
		AcceptedAnswers []string `json:"accepted_answers"`
		Material        struct {
			Kind        string `json:"kind"`
			AudioURL    string `json:"audio_url"`
			PassageText string `json:"passage_text"`
		} `json:"material"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(req); err != nil {
		return err
	}
	options := make([]command.QuestionOption, 0, len(req.Options))
	for _, option := range req.Options {
		options = append(options, command.QuestionOption{ID: option.ID, Text: option.Text})
	}
	result, err := h.app.Commands.CreateQuestion.Handle(c.Request().Context(), command.CreateQuestion{
		Type:            req.Type,
		Prompt:          req.Prompt,
		Options:         options,
		CorrectOptionID: req.CorrectOptionID,
		AcceptedAnswers: req.AcceptedAnswers,
		Material: command.QuestionMaterial{
			Kind:        req.Material.Kind,
			AudioURL:    req.Material.AudioURL,
			PassageText: req.Material.PassageText,
		},
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, result)
}

func (h *Handler) listQuestions(c *echo.Context) error {
	result, err := h.app.Queries.ListQuestions.Handle(c.Request().Context(), query.ListQuestions{})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) createContest(c *echo.Context) error {
	var req struct {
		Title              string    `json:"title" validate:"required"`
		StartTime          time.Time `json:"start_time" validate:"required"`
		PerQuestionSeconds int       `json:"per_question_seconds"`
		QuestionIDs        []string  `json:"question_ids" validate:"required,min=1"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(req); err != nil {
		return err
	}
	result, err := h.app.Commands.CreateContest.Handle(c.Request().Context(), command.CreateContest{
		Title:              req.Title,
		StartTime:          req.StartTime,
		PerQuestionSeconds: req.PerQuestionSeconds,
		QuestionIDs:        req.QuestionIDs,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, result)
}

func (h *Handler) publishContest(c *echo.Context) error {
	if err := h.app.Commands.PublishContest.Handle(c.Request().Context(), command.PublishContest{ContestID: c.Param("contest_id")}); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) getArena(c *echo.Context) error {
	result, err := h.app.Queries.GetArena.Handle(c.Request().Context(), query.GetArena{ContestID: c.Param("contest_id")})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) submitAnswer(c *echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	var req struct {
		Answers []struct {
			QuestionID string `json:"question_id"`
			OptionID   string `json:"option_id"`
			Text       string `json:"text"`
		} `json:"answers" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(req); err != nil {
		return err
	}
	answers := make([]command.SubmittedAnswer, 0, len(req.Answers))
	for _, answer := range req.Answers {
		answers = append(answers, command.SubmittedAnswer{
			QuestionID: answer.QuestionID,
			OptionID:   answer.OptionID,
			Text:       answer.Text,
		})
	}
	result, err := h.app.Commands.SubmitAnswer.Handle(c.Request().Context(), command.SubmitAnswer{
		ContestID: c.Param("contest_id"),
		UserID:    userID,
		Answers:   answers,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}

func (h *Handler) claimReward(c *echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	result, err := h.app.Commands.ClaimReward.Handle(c.Request().Context(), command.ClaimReward{ContestID: c.Param("contest_id"), UserID: userID})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, result)
}

func currentUserID(c *echo.Context) (string, error) {
	userID, ok := httpserver.UserUUID(c)
	if !ok {
		return "", echo.NewHTTPError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
	}
	return userID, nil
}
