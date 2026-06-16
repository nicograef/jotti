// triggerBrowserDownload speichert einen Blob unter dem gegebenen Dateinamen,
// indem es einen temporären Download-Link erzeugt und klickt. Nötig, weil der
// Export authentifiziert ist und daher nicht über einen einfachen <a href>
// geladen werden kann.
export function triggerBrowserDownload(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = filename
  document.body.appendChild(anchor)
  anchor.click()
  anchor.remove()
  URL.revokeObjectURL(url)
}
