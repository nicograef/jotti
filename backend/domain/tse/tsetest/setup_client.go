package tsetest

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/tse"
)

type RegistrierterClient struct {
	TssID        string
	ClientID     string
	SerialNumber string
}

// ReaktivierterClient: eine Reaktivierung trägt keine serial_number, sie aktiviert
// den vorhandenen Client unter seiner ID wieder.
type ReaktivierterClient struct {
	TssID    string
	ClientID string
}

// FakeSetupClient hat Pointer-Receiver, damit die Aufzeichnungsfelder über den
// Interface-Wert hinweg sichtbar bleiben.
type FakeSetupClient struct {
	UmgebungResponse tse.Umgebung
	TSSResponse      []tse.TSSInfo
	TSSErr           error
	ClientsByTSS     map[string][]tse.ClientInfo
	ClientsErr       error

	CreateTSSResponse   tse.TSSErstellt
	CreateTSSErr        error
	GetAdminPUKResponse string
	GetAdminPUKErr      error
	StammdatenResponse  tse.Stammdaten
	StammdatenErr       error
	PersonalisiereErr   error
	SetAdminPINErr      error
	AuthAdminErr        error
	InitialisiereErr    error
	RegistriereErr      error
	ReaktiviereErr      error

	ErstellteTSS         []tse.TSSErstellt
	StammdatenTssID      string
	AdminAuthentifiziert bool
	GesetzteAdminPIN     string
	GesetzterAdminPUK    string
	AuthentifiziertePIN  string
	RegistrierteClients  []RegistrierterClient
	ReaktivierteClients  []ReaktivierterClient
}

var _ tse.SetupClient = (*FakeSetupClient)(nil)

func (f *FakeSetupClient) ListTSS(context.Context) (tse.Umgebung, []tse.TSSInfo, error) {
	if f.TSSErr != nil {
		return "", nil, f.TSSErr
	}
	return f.UmgebungResponse, f.TSSResponse, nil
}

func (f *FakeSetupClient) ListClients(_ context.Context, tssID string) ([]tse.ClientInfo, error) {
	if f.ClientsErr != nil {
		return nil, f.ClientsErr
	}
	return f.ClientsByTSS[tssID], nil
}

func (f *FakeSetupClient) RetrieveTSSStammdaten(_ context.Context, tssID string) (tse.Stammdaten, error) {
	f.StammdatenTssID = tssID
	if f.StammdatenErr != nil {
		return tse.Stammdaten{}, f.StammdatenErr
	}
	return f.StammdatenResponse, nil
}

func (f *FakeSetupClient) CreateTSS(context.Context) (tse.TSSErstellt, error) {
	if f.CreateTSSErr != nil {
		return tse.TSSErstellt{}, f.CreateTSSErr
	}
	f.ErstellteTSS = append(f.ErstellteTSS, f.CreateTSSResponse)
	return f.CreateTSSResponse, nil
}

func (f *FakeSetupClient) GetAdminPUK(context.Context, string) (string, error) {
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
	if f.AuthAdminErr != nil {
		return f.AuthAdminErr
	}
	f.AdminAuthentifiziert = true
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
