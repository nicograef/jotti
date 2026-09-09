//go:build unit

package kasse

import "testing"

// Der Kassenabschluss vergleicht beim Wiederanlauf den Soll-Bestand ohne die
// abschluss-eigene Differenzbuchung gegen den im Kassensturz protokollierten
// Soll-Bestand. Die Ableitung muss deshalb genau die vier Komponenten summieren —
// die Entnahme als einziger Abzug — und darf von SollBestandCents nur um eine
// gebuchte Differenz abweichen.
func TestKassenbestand_SollBestandOhneDifferenzCents(t *testing.T) {
	cases := []struct {
		name string
		// bestand trägt in beiden Fällen dieselben vier Komponenten; nur
		// SollBestandCents unterscheidet sich um die gebuchte Differenz.
		bestand Kassenbestand
		want    int
		// wantDifferenz ist die gebuchte Differenz (Soll − Ist), also der Abstand
		// zwischen dem Bestand ohne sie und SollBestandCents.
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
				// Fehlbetrag von 500 gebucht: SollBestandCents ist an den gezählten
				// Ist-Bestand angeglichen, die vier Komponenten sind unberührt.
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
