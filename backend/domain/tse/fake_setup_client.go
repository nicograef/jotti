package tse

import "context"

// RegistrierterClient haelt die Argumente eines RegistriereClient-Aufrufs fuer
// Assertions in den Orchestrator-Tests fest.
type RegistrierterClient struct {
	TssID        string
	ClientID     string
	SerialNumber string
}

// FakeSetupClient ist das Test-Double fuer SetupClient — analog zu FakeClient.
// Die Methoden haben Pointer-Receiver, damit die Aufzeichnungsfelder
// (CreateTSSCalls, RegistrierteClients, ...) ueber den Interface-Wert hinweg
// sichtbar bleiben.
type FakeSetupClient struct {
	UmgebungResponse Umgebung
	TSSResponse      []TSSInfo
	TSSErr           error
	ClientsByTSS     map[string][]ClientInfo
	ClientsErr       error

	CreateTSSResponse    TSSErstellt
	CreateTSSErr         error
	HoleAdminPUKResponse string
	HoleAdminPUKErr      error
	PersonalisiereErr    error
	SetzeAdminPINErr     error
	AuthAdminErr         error
	InitialisiereErr     error
	RegistriereErr       error

	// Aufzeichnung fuer Assertions.
	CreateTSSCalls      int
	HoleAdminPUKCalls   int
	GesetzteAdminPIN    string
	AuthentifiziertePIN string
	RegistrierteClients []RegistrierterClient
}

var _ SetupClient = (*FakeSetupClient)(nil)

func (f *FakeSetupClient) ListTSS(context.Context) (Umgebung, []TSSInfo, error) {
	if f.TSSErr != nil {
		return "", nil, f.TSSErr
	}
	return f.UmgebungResponse, f.TSSResponse, nil
}

func (f *FakeSetupClient) ListClients(_ context.Context, tssID string) ([]ClientInfo, error) {
	if f.ClientsErr != nil {
		return nil, f.ClientsErr
	}
	return f.ClientsByTSS[tssID], nil
}

func (f *FakeSetupClient) CreateTSS(context.Context) (TSSErstellt, error) {
	f.CreateTSSCalls++
	if f.CreateTSSErr != nil {
		return TSSErstellt{}, f.CreateTSSErr
	}
	return f.CreateTSSResponse, nil
}

func (f *FakeSetupClient) HoleAdminPUK(context.Context, string) (string, error) {
	f.HoleAdminPUKCalls++
	if f.HoleAdminPUKErr != nil {
		return "", f.HoleAdminPUKErr
	}
	return f.HoleAdminPUKResponse, nil
}

func (f *FakeSetupClient) PersonalisiereTSS(context.Context, string) error {
	return f.PersonalisiereErr
}

func (f *FakeSetupClient) SetzeAdminPIN(_ context.Context, _, _, pin string) error {
	if f.SetzeAdminPINErr != nil {
		return f.SetzeAdminPINErr
	}
	f.GesetzteAdminPIN = pin
	return nil
}

func (f *FakeSetupClient) AuthentifiziereAdmin(_ context.Context, _, pin string) error {
	if f.AuthAdminErr != nil {
		return f.AuthAdminErr
	}
	f.AuthentifiziertePIN = pin
	return nil
}

func (f *FakeSetupClient) InitialisiereTSS(context.Context, string) error {
	return f.InitialisiereErr
}

func (f *FakeSetupClient) RegistriereClient(_ context.Context, tssID, clientID, serialNumber string) error {
	if f.RegistriereErr != nil {
		return f.RegistriereErr
	}
	f.RegistrierteClients = append(f.RegistrierteClients, RegistrierterClient{
		TssID:        tssID,
		ClientID:     clientID,
		SerialNumber: serialNumber,
	})
	return nil
}
