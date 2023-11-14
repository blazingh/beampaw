package helper

import (
	"io"

	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
)

func PrintProjectHeader(w io.Writer) {
	pterm.DefaultBasicText.WithWriter(w).Print("\n\n")
	pterm.DefaultBigText.
		WithWriter(w).
		WithLetters(
			putils.LettersFromStringWithStyle("BEAM ", pterm.FgCyan.ToStyle()),
			putils.LettersFromStringWithStyle("PAW", pterm.FgLightMagenta.ToStyle())).
		Render()
	pterm.DefaultBasicText.WithWriter(w).Print("\n\n")
}
