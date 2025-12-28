package syncdoterr

import (
	"bytes"
	"strings"
)

func replacePackageLine(source []byte, pkg string) (updated []byte, err error) {
	var lines [][]byte
	var line []byte
	var lineStr string
	var fields []string
	var buf bytes.Buffer
	var replaced bool
	var i int

	lines = bytes.Split(source, []byte("\n"))
	for i, line = range lines {
		lineStr = strings.TrimSpace(string(line))

		if !replaced {
			fields = strings.Fields(lineStr)
			if len(fields) == 2 && fields[0] == "package" && isIdent(fields[1]) {
				line = []byte("package " + pkg)
				replaced = true
			}
		}

		_, err = buf.Write(line)
		if err != nil {
			goto end
		}
		if i < len(lines)-1 {
			err = buf.WriteByte('\n')
			if err != nil {
				goto end
			}
		}
	}

	if !replaced {
		err = NewErr(ErrMissingPackageDecl,
			"package", pkg,
		)
		goto end
	}

	updated = buf.Bytes()

end:
	return updated, err
}
