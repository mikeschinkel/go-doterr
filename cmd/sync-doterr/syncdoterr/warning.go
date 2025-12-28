package syncdoterr

import (
	"github.com/mikeschinkel/go-cliutil"
)

func warning(w cliutil.Writer, format string, args ...any) {
	w.Errorf("\n\tWARNING: "+format, args...)
}
