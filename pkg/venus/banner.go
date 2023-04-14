package venus

import "github.com/zan8in/gologger"

var Version = "0.0.1"

func ShowBanner() {
	gologger.Print().Msgf("\n|||\tV E N U S\t|||\t%s\n\n", Version)
}
