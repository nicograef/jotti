//go:build unit

package kasse

import "testing"

// Der Kassenabschluss vergleicht beim Wiederanlauf den Soll-Bestand ohne die
// abschluss-eigene Differenzbuchung gegen den im Kassensturz protokollierten Wert. Die
// Ableitung muss deshalb genau die vier Komponenten summieren, die Entnahme als Abzug.
func TestKassenbestand_SollBestandOhneDifferenzCents(t *testing.T) {
	cases := []struct {
		name    string
		bestand Kassenbestand
		want    int
		// wantDifferenz ist die gebuchte Differenz (Soll − Ist).
		wantDifferenz int
	}{
		{
			name: "ohne gebuchte Differenz summiert die vier Komponenten",
			bestand: Kassenbestand{
				SollBestandCents:    34000,
				AnfangsbestandCents: 15000,
				BareinnahmenCents:   17000,
				EinlagenCents:       3000,
				EntnahmenCents:      1000,
			},
			want:          34000,
			wantDifferenz: 0,
		},
		{
			name: "mit gebuchter Differenz bleibt der Bestand der Buchungen",
			bestand: Kassenbestand{
				// Fehlbetrag 500 gebucht: SollBestandCents ist an den Ist-Bestand angeglichen.
				SollBestandCents:    33500,
				AnfangsbestandCents: 15000,
				BareinnahmenCents:   17000,
				EinlagenCents:       3000,
				EntnahmenCents:      1000,
			},
			want:          34000,
			wantDifferenz: 500,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.bestand.SollBestandOhneDifferenzCents()
			if got != tc.want {
				t.Errorf("SollBestandOhneDifferenzCents() = %d, erwartet %d", got, tc.want)
			}
			if differenz := got - tc.bestand.SollBestandCents; differenz != tc.wantDifferenz {
				t.Errorf("Abstand zu SollBestandCents = %d, erwartet %d", differenz, tc.wantDifferenz)
			}
		})
	}
}
