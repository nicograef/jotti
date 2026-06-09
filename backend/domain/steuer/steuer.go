package steuer

import z "github.com/Oudwins/zog"

type Steuersatz string

const (
	RegelSteuersatz      Steuersatz = "regel"
	ErmaessigtSteuersatz Steuersatz = "ermaessigt"
	BefreitSteuersatz    Steuersatz = "befreit"
	KombiSteuersatz      Steuersatz = "kombi"
)

type Aufteilung struct {
	Satz   Steuersatz
	Brutto int
	Netto  int
	Steuer int
}

var SteuersatzSchema = z.StringLike[Steuersatz]().OneOf(
	[]Steuersatz{RegelSteuersatz, ErmaessigtSteuersatz, BefreitSteuersatz, KombiSteuersatz},
	z.Message("Ungueltiger Steuersatz"),
)

func (s Steuersatz) Prozent() int {
	switch s {
	case RegelSteuersatz:
		return 19
	case ErmaessigtSteuersatz:
		return 7
	case BefreitSteuersatz:
		return 0
	default:
		return -1
	}
}

func Aufteilen(brutto int, satz Steuersatz) []Aufteilung {
	if brutto < 0 {
		return nil
	}

	switch satz {
	case KombiSteuersatz:
		ermaessigtBrutto := roundHalfUpDivide(brutto*70, 100)
		regelBrutto := brutto - ermaessigtBrutto

		return []Aufteilung{
			aufteilenEinzeln(ermaessigtBrutto, ErmaessigtSteuersatz),
			aufteilenEinzeln(regelBrutto, RegelSteuersatz),
		}
	case RegelSteuersatz, ErmaessigtSteuersatz, BefreitSteuersatz:
		return []Aufteilung{aufteilenEinzeln(brutto, satz)}
	default:
		return nil
	}
}

func aufteilenEinzeln(brutto int, satz Steuersatz) Aufteilung {
	steuerbetrag := steuerAnteil(brutto, satz.Prozent())

	return Aufteilung{
		Satz:   satz,
		Brutto: brutto,
		Netto:  brutto - steuerbetrag,
		Steuer: steuerbetrag,
	}
}

func steuerAnteil(brutto int, prozent int) int {
	if brutto <= 0 || prozent <= 0 {
		return 0
	}

	return roundHalfUpDivide(brutto*prozent, 100+prozent)
}

func roundHalfUpDivide(numerator int, denominator int) int {
	if denominator <= 0 {
		return 0
	}

	if numerator >= 0 {
		return (numerator + denominator/2) / denominator
	}

	return (numerator - denominator/2) / denominator
}
