package tse

import (
	"context"
	"time"
)

type FakeClient struct {
	StartResponse      StartResult
	StartErr           error
	FinishResponse     FinishResult
	FinishErr          error
	RetrieveResponse   RetrieveResult
	RetrieveErr        error
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

func (f FakeClient) StartTransaction(ctx context.Context, _ string) (StartResult, error) {
	if err := f.wait(ctx); err != nil {
		return StartResult{}, err
	}
	if f.StartErr != nil {
		return StartResult{}, f.StartErr
	}
	return f.StartResponse, nil
}

func (f FakeClient) FinishTransaction(ctx context.Context, _ string, _ string, _ string) (FinishResult, error) {
	if err := f.wait(ctx); err != nil {
		return FinishResult{}, err
	}
	if f.FinishErr != nil {
		return FinishResult{}, f.FinishErr
	}
	return f.FinishResponse, nil
}

func (f FakeClient) RetrieveTransaction(ctx context.Context, _ string) (RetrieveResult, error) {
	if err := f.wait(ctx); err != nil {
		return RetrieveResult{}, err
	}
	if f.RetrieveErr != nil {
		return RetrieveResult{}, f.RetrieveErr
	}
	return f.RetrieveResponse, nil
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
