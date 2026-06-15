package tse

import "context"

// FakeSetupClient ist das Test-Double fuer SetupClient — analog zu FakeClient.
type FakeSetupClient struct {
	UmgebungResponse Umgebung
	TSSResponse      []TSSInfo
	TSSErr           error
	ClientsByTSS     map[string][]ClientInfo
	ClientsErr       error
}

func (f FakeSetupClient) ListTSS(context.Context) (Umgebung, []TSSInfo, error) {
	if f.TSSErr != nil {
		return "", nil, f.TSSErr
	}
	return f.UmgebungResponse, f.TSSResponse, nil
}

func (f FakeSetupClient) ListClients(_ context.Context, tssID string) ([]ClientInfo, error) {
	if f.ClientsErr != nil {
		return nil, f.ClientsErr
	}
	return f.ClientsByTSS[tssID], nil
}
