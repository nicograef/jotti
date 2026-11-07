// Package zeichen bounds string lengths in characters, as the frontend's Zod schemas and the official
// field lengths count them. zog's Min and Max count bytes, so an umlaut would tighten every limit.
package zeichen

import (
	"unicode/utf8"

	z "github.com/Oudwins/zog"
)

// Min is the zog test "at least n characters".
func Min(n int) z.BoolTFunc[*string] {
	return func(wert *string, _ z.Ctx) bool {
		return utf8.RuneCountInString(*wert) >= n
	}
}

// Max is the zog test "at most n characters".
func Max(n int) z.BoolTFunc[*string] {
	return func(wert *string, _ z.Ctx) bool {
		return utf8.RuneCountInString(*wert) <= n
	}
}
