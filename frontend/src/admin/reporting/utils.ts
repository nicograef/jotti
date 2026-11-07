// Zeigt den eingefrorenen Username, ergänzt um den live aufgelösten Klarnamen.
export function formatServicekraft(userName: string, name: string): string {
  return name ? `${userName} (${name})` : userName
}
