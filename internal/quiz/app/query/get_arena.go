package query

import "context"

type GetArena struct {
	ContestID string
}

type GetArenaHandler struct {
	ReadModel ReadModel
}

func (h GetArenaHandler) Handle(ctx context.Context, q GetArena) (ArenaDTO, error) {
	return h.ReadModel.GetArena(ctx, q.ContestID)
}
