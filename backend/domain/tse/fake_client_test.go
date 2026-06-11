//go:build unit

package tse

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestFakeClient_Success(t *testing.T) {
	fake := FakeClient{
		StartResponse:      StartResult{TransactionNumber: 10, SignatureCounter: 12},
		FinishResponse:     FinishResult{TransactionNumber: 10, SignatureCounter: 13, Signature: "abc"},
		ConnectionResponse: VerbindungStatus{Umgebung: UmgebungTest, TSSState: "INITIALIZED"},
	}

	start, err := fake.StartTransaction(context.Background(), "8e9e7b56-31a8-43e3-9b29-d92a2b78b561")
	if err != nil {
		t.Fatalf("expected no start error, got %v", err)
	}
	if start.TransactionNumber != 10 {
		t.Fatalf("expected transaction number 10, got %d", start.TransactionNumber)
	}

	finish, err := fake.FinishTransaction(context.Background(), "8e9e7b56-31a8-43e3-9b29-d92a2b78b561", "Kassenbeleg-V1", "Beleg^0.00")
	if err != nil {
		t.Fatalf("expected no finish error, got %v", err)
	}
	if finish.Signature != "abc" {
		t.Fatalf("expected signature abc, got %q", finish.Signature)
	}

	status, err := fake.TestConnection(context.Background())
	if err != nil {
		t.Fatalf("expected no connection error, got %v", err)
	}
	if status.Umgebung != UmgebungTest {
		t.Fatalf("expected TEST environment, got %s", status.Umgebung)
	}
}

func TestFakeClient_ConfiguredErrors(t *testing.T) {
	expectedErr := errors.New("boom")
	fake := FakeClient{
		StartErr:      expectedErr,
		FinishErr:     expectedErr,
		ConnectionErr: expectedErr,
	}

	if _, err := fake.StartTransaction(context.Background(), "8e9e7b56-31a8-43e3-9b29-d92a2b78b561"); !errors.Is(err, expectedErr) {
		t.Fatalf("expected start error boom, got %v", err)
	}
	if _, err := fake.FinishTransaction(context.Background(), "8e9e7b56-31a8-43e3-9b29-d92a2b78b561", "Kassenbeleg-V1", ""); !errors.Is(err, expectedErr) {
		t.Fatalf("expected finish error boom, got %v", err)
	}
	if _, err := fake.TestConnection(context.Background()); !errors.Is(err, expectedErr) {
		t.Fatalf("expected connection error boom, got %v", err)
	}
}

func TestFakeClient_Timeout(t *testing.T) {
	fake := FakeClient{ArtificialDelay: 150 * time.Millisecond}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := fake.TestConnection(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded, got %v", err)
	}
}
