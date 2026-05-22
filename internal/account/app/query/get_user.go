package query

import "context"

type GetUser struct {
	UserID string
}

type GetUserHandler struct {
	ReadModel ReadModel
}

func (h GetUserHandler) Handle(ctx context.Context, q GetUser) (UserDTO, error) {
	return h.ReadModel.GetUser(ctx, q.UserID)
}
