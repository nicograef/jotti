// Package http maps the position-enrichment sentinels for every Kasse handler that enriches positions.
package http

import (
	"github.com/nicograef/jotti/backend/api/helper"
	"github.com/nicograef/jotti/backend/api/kasse/enrichment"
)

// Fehler maps the product and Pfand checks of the position enrichment.
var Fehler = []helper.ErrorCode{
	{Err: enrichment.ErrKeinPfandProdukt, Code: "kein_pfand_produkt"},
	{Err: enrichment.ErrProduktNotFound, Code: "produkt_not_found"},
	{Err: enrichment.ErrVarianteNichtAktiv, Code: "variante_nicht_aktiv"},
}
