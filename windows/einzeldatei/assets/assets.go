// Package assets bettet ein, was scripts/einzeldatei-assets.sh nach files/ legt:
// postgres.zip, den App-Build unter app/ und die Migrationen unter migrations/.
package assets

import "embed"

// FS enthält files/; ohne Assets nur files/.gitkeep, damit das Modul auch ohne Build-Lauf kompiliert.
//
//go:embed all:files
var FS embed.FS
