package api

import "net/http"

// routeMux ist ein http.ServeMux, der die registrierten Pfade zusätzlich
// mitschreibt. So bleibt die Liste der Pfade eines Bereichs die einzige Quelle:
// jeder per HandleFunc registrierte Endpunkt landet automatisch in Paths() und
// damit in der Berechtigungs-Matrix — es kann keinen Endpunkt geben, der nicht
// erfasst ist.
type routeMux struct {
	mux   *http.ServeMux
	paths []string
}

func newRouteMux() *routeMux {
	return &routeMux{mux: http.NewServeMux()}
}

// HandleFunc registriert einen Handler und merkt sich den Pfad.
func (m *routeMux) HandleFunc(pattern string, handler http.HandlerFunc) {
	m.mux.HandleFunc(pattern, handler)
	m.paths = append(m.paths, pattern)
}

// Handler gibt den zugrunde liegenden Mux als http.Handler zurück.
func (m *routeMux) Handler() http.Handler {
	return m.mux
}

// Paths liefert die registrierten Pfade in Registrierungsreihenfolge.
func (m *routeMux) Paths() []string {
	return m.paths
}
