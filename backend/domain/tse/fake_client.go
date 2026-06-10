package tse

import (
	"context"
	"time"
)

type FakeClient struct {
	StartResponse      StartResult
	StartErr           error
	UpdateErr          error
	FinishResponse     FinishResult
	FinishErr          error
	ConnectionResponse VerbindungStatus
	ConnectionErr      error
	ArtificialDelay    time.Duration
}

func (f FakeClient) wait(ctx context.Context) error {
	if f.ArtificialDelay <= 0 {
		return nil
	}
	timer := time.NewTimer(f.ArtificialDelay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (f FakeClient) StartTransaction(ctx context.Context, _ string, _ string, _ string) (StartResult, error) {
	if err := f.wait(ctx); err != nil {
		return StartResult{}, err
	}
	if f.StartErr != nil {
		return StartResult{}, f.StartErr
	}
	return f.StartResponse, nil
}

func (f FakeClient) UpdateTransaction(ctx context.Context, _ string, _ int, _ string) error {
	if err := f.wait(ctx); err != nil {
		return err
	}
	return f.UpdateErr
}

func (f FakeClient) FinishTransaction(ctx context.Context, _ string, _ int, _ string, _ string) (FinishResult, error) {
	if err := f.wait(ctx); err != nil {
		return FinishResult{}, err
	}
	if f.FinishErr != nil {
		return FinishResult{}, f.FinishErr
	}
	return f.FinishResponse, nil
}

func (f FakeClient) TestConnection(ctx context.Context) (VerbindungStatus, error) {
	if err := f.wait(ctx); err != nil {
		return VerbindungStatus{}, err
	}
	if f.ConnectionErr != nil {
		return VerbindungStatus{}, f.ConnectionErr
	}
	return f.ConnectionResponse, nil
}
