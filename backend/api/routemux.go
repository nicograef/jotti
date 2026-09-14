package api

import "net/http"

// routeMux schreibt jeden per HandleFunc registrierten Pfad mit: Paths() speist
// die Berechtigungs-Matrix, es kann also keinen unerfassten Endpunkt geben.
type routeMux struct {
	mux   *http.ServeMux
	paths []string
}

func newRouteMux() *routeMux {
	return &routeMux{mux: http.NewServeMux()}
}

func (m *routeMux) HandleFunc(pattern string, handler http.HandlerFunc) {
	m.mux.HandleFunc(pattern, handler)
	m.paths = append(m.paths, pattern)
}

func (m *routeMux) Handler() http.Handler {
	return m.mux
}

func (m *routeMux) Paths() []string {
	return m.paths
}
