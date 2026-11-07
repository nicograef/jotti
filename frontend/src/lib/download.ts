// Der Export ist authentifiziert und lässt sich daher nicht über ein einfaches
// <a href> laden; stattdessen ein temporärer, geklickter Download-Link.
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
