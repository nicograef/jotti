package tse_repo

// signaturWorkerTrigger wakes the Signatur-Worker after each commit that adds a Signaturauftrag.
// Buffered with a non-blocking send: a lost trigger is harmless because the worker's polling tick is the fallback.
var signaturWorkerTrigger = make(chan struct{}, 1)

// NotifySignaturWorker wakes the Signatur-Worker without blocking.
func NotifySignaturWorker() {
	select {
	case signaturWorkerTrigger <- struct{}{}:
	default:
	}
}

// SignaturWorkerTrigger returns the channel the Signatur-Worker listens on.
func SignaturWorkerTrigger() <-chan struct{} {
	return signaturWorkerTrigger
}
