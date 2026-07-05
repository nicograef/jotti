package tse

import "context"

// RegistrierterClient haelt die Argumente eines RegistriereClient-Aufrufs fuer
// Assertions in den Orchestrator-Tests fest.
type RegistrierterClient struct {
	TssID        string
	ClientID     string
	SerialNumber string
}

// ReaktivierterClient haelt die Argumente eines ReaktiviereClient-Aufrufs fest.
// Eine Reaktivierung traegt keine serial_number — sie aktiviert den vorhandenen
// Client unter seiner ID wieder.
type ReaktivierterClient struct {
	TssID    string
	ClientID string
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

	CreateTSSResponse   TSSErstellt
	CreateTSSErr        error
	GetAdminPUKResponse string
	GetAdminPUKErr      error
	StammdatenResponse  TSSStammdaten
	StammdatenErr       error
	PersonalisiereErr   error
	SetAdminPINErr      error
	AuthAdminErr        error
	InitialisiereErr    error
	RegistriereErr      error
	ReaktiviereErr      error

	// Aufzeichnung fuer Assertions.
	CreateTSSCalls      int
	GetAdminPUKCalls    int
	StammdatenCalls     int
	StammdatenTssID     string
	AuthAdminCalls      int
	GesetzteAdminPIN    string
	GesetzterAdminPUK   string
	AuthentifiziertePIN string
	RegistrierteClients []RegistrierterClient
	ReaktivierteClients []ReaktivierterClient
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

func (f *FakeSetupClient) RetrieveTSSStammdaten(_ context.Context, tssID string) (TSSStammdaten, error) {
	f.StammdatenCalls++
	f.StammdatenTssID = tssID
	if f.StammdatenErr != nil {
		return TSSStammdaten{}, f.StammdatenErr
	}
	return f.StammdatenResponse, nil
}

func (f *FakeSetupClient) CreateTSS(context.Context) (TSSErstellt, error) {
	f.CreateTSSCalls++
	if f.CreateTSSErr != nil {
		return TSSErstellt{}, f.CreateTSSErr
	}
	return f.CreateTSSResponse, nil
}

func (f *FakeSetupClient) GetAdminPUK(context.Context, string) (string, error) {
	f.GetAdminPUKCalls++
	if f.GetAdminPUKErr != nil {
		return "", f.GetAdminPUKErr
	}
	return f.GetAdminPUKResponse, nil
}

func (f *FakeSetupClient) PersonalisiereTSS(context.Context, string) error {
	return f.PersonalisiereErr
}

func (f *FakeSetupClient) SetAdminPIN(_ context.Context, _, puk, pin string) error {
	if f.SetAdminPINErr != nil {
		return f.SetAdminPINErr
	}
	f.GesetzteAdminPIN = pin
	f.GesetzterAdminPUK = puk
	return nil
}

func (f *FakeSetupClient) AuthentifiziereAdmin(_ context.Context, _, pin string) error {
	f.AuthAdminCalls++
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

func (f *FakeSetupClient) ReaktiviereClient(_ context.Context, tssID, clientID string) error {
	if f.ReaktiviereErr != nil {
		return f.ReaktiviereErr
	}
	f.ReaktivierteClients = append(f.ReaktivierteClients, ReaktivierterClient{
		TssID:    tssID,
		ClientID: clientID,
	})
	return nil
}
