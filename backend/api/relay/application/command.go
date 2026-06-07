package application

import "context"

type Command struct {
	DruckauftragRepo druckauftragCommandRepo
}

type druckauftragCommandRepo interface {
	QuittiereGedruckteAuftraege(ctx context.Context, ids []int) error
}

func (c Command) QuittiereGedruckteAuftraege(ctx context.Context, ids []int) error {
	return c.DruckauftragRepo.QuittiereGedruckteAuftraege(ctx, ids)
}
