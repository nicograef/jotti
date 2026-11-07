package tse

import "testing"

func TestVerbindungStatus_Signiertfaehig(t *testing.T) {
	cases := []struct {
		status VerbindungStatus
		want   bool
	}{
		{VerbindungStatus{TSSState: TSSInitialized, ClientState: ClientRegistered, SeriennummerKorrekt: true}, true},
		{VerbindungStatus{TSSState: TSSInitialized, ClientState: ClientRegistered, SeriennummerKorrekt: false}, false},
		{VerbindungStatus{TSSState: TSSInitialized, ClientState: ClientDeregistered, SeriennummerKorrekt: true}, false},
		{VerbindungStatus{TSSState: TSSUninitialized, ClientState: ClientRegistered, SeriennummerKorrekt: true}, false},
	}
	for _, tc := range cases {
		if got := tc.status.Signiertfaehig(); got != tc.want {
			t.Errorf("%+v.Signiertfaehig() = %v, want %v", tc.status, got, tc.want)
		}
	}
}
