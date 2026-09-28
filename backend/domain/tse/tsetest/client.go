// Package tsetest holds fakes of the TSE clients for unit tests. Only test files import it.
package tsetest

import (
	"context"
	"time"

	"github.com/nicograef/jotti/backend/domain/tse"
)

type FakeClient struct {
	StartResponse      tse.StartResult
	StartErr           error
	FinishResponse     tse.FinishResult
	FinishErr          error
	RetrieveResponse   tse.RetrieveResult
	RetrieveErr        error
	ConnectionResponse tse.VerbindungStatus
	ConnectionErr      error
	UmgebungResponse   tse.Umgebung
	UmgebungErr        error
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

func (f FakeClient) StartTransaction(ctx context.Context, _ string) (tse.StartResult, error) {
	if err := f.wait(ctx); err != nil {
		return tse.StartResult{}, err
	}
	if f.StartErr != nil {
		return tse.StartResult{}, f.StartErr
	}
	return f.StartResponse, nil
}

func (f FakeClient) FinishTransaction(ctx context.Context, _ string, _ string, _ string) (tse.FinishResult, error) {
	if err := f.wait(ctx); err != nil {
		return tse.FinishResult{}, err
	}
	if f.FinishErr != nil {
		return tse.FinishResult{}, f.FinishErr
	}
	return f.FinishResponse, nil
}

func (f FakeClient) RetrieveTransaction(ctx context.Context, _ string) (tse.RetrieveResult, error) {
	if err := f.wait(ctx); err != nil {
		return tse.RetrieveResult{}, err
	}
	if f.RetrieveErr != nil {
		return tse.RetrieveResult{}, f.RetrieveErr
	}
	return f.RetrieveResponse, nil
}

func (f FakeClient) TestConnection(ctx context.Context) (tse.VerbindungStatus, error) {
	if err := f.wait(ctx); err != nil {
		return tse.VerbindungStatus{}, err
	}
	if f.ConnectionErr != nil {
		return tse.VerbindungStatus{}, f.ConnectionErr
	}
	return f.ConnectionResponse, nil
}

func (f FakeClient) Umgebung(ctx context.Context) (tse.Umgebung, error) {
	if err := f.wait(ctx); err != nil {
		return "", err
	}
	if f.UmgebungErr != nil {
		return "", f.UmgebungErr
	}
	return f.UmgebungResponse, nil
}
