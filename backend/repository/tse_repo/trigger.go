package tse_repo

// signaturWorkerTrigger ist die In-Process-Benachrichtigung an den
// Signatur-Worker: Nach jedem Commit mit neuem Signaturauftrag wird er sofort
// angestossen. Der Kanal ist gepuffert und der Send non-blocking — ein
// verlorener Trigger ist unkritisch, der Polling-Tick des Workers bleibt
// Fallback.
var signaturWorkerTrigger = make(chan struct{}, 1)

// NotifySignaturWorker stoesst den Signatur-Worker non-blocking an.
func NotifySignaturWorker() {
	select {
	case signaturWorkerTrigger <- struct{}{}:
	default:
	}
}

// SignaturWorkerTrigger liefert den Kanal, auf den der Signatur-Worker lauscht.
func SignaturWorkerTrigger() <-chan struct{} {
	return signaturWorkerTrigger
}
