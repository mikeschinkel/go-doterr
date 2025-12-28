package syncdoterr

import (
	"bufio"
	"errors"
	"os"
	"strings"

	"github.com/mikeschinkel/go-dt"
)

var ErrOnlyDotErrFilesFound = errors.New("only doterr.go file found; no other .go files found in package")
var ErrPackageNameNotFound = errors.New("package name not found")

func detectPackageName(dotErrFile dt.Filepath) (name string, err error) {
	var entries []os.DirEntry
	var entry os.DirEntry
	var fullpath dt.Filepath
	var candidate string
	var ok bool
	var base string
	var targetBase dt.Filename
	var entryName string
	var entryFilename dt.Filename
	var entryExt dt.FileExt

	dir := dotErrFile.Dir()
	entries, err = dir.ReadDir()
	if err != nil {
		goto end
	}

	//Not needed because we will only get here if the dotErrFile exists
	//if len(entries) == 0 {
	//	goto end
	//}

	targetBase = dotErrFile.Base()

	for _, entry = range entries {
		if entry.IsDir() {
			continue
		}

		entryName = entry.Name()
		entryFilename = dt.Filename(entryName)
		entryExt = entryFilename.Ext()
		if entryExt != GoFileExt {
			continue
		}

		if entryFilename == targetBase {
			if len(entries) == 1 {
				// Not using NewErr() on purpose since this is an internal function and the sole
				// purpose is to communicate to this info to called, not to propagate the error.
				err = ErrOnlyDotErrFilesFound
				goto end
			}
			continue
		}

		fullpath = dt.FilepathJoin(dir, entryName)
		candidate, ok, err = packageNameFromFile(fullpath)
		if err != nil {
			goto end
		}
		if !ok {
			continue
		}
		if strings.HasSuffix(candidate, "_test") {
			continue
		}

		name = candidate
		goto end
	}

	base = string(dir.Base())
	name = base
	// Not using NewErr() on purpose since this is an internal function and the sole
	// purpose is to communicate to this info to called, not to propagate the error.
	err = ErrPackageNameNotFound

end:
	return name, err
}

func packageNameFromFile(path dt.Filepath) (name string, ok bool, err error) {
	var file *os.File
	var scanner *bufio.Scanner
	var line string
	var fields []string
	var closeErr error

	file, err = path.Open()
	if err != nil {
		goto end
	}

	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		line = strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "package ") {
			continue
		}

		fields = strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		if fields[0] != "package" {
			continue
		}
		if !isIdent(fields[1]) {
			continue
		}

		name = fields[1]
		ok = true
		goto end
	}

	err = scanner.Err()

end:
	if file != nil {
		closeErr = file.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return name, ok, err
}

func isIdent(s string) (ok bool) {
	var i int
	var b byte

	if s == "" {
		goto end
	}

	b = s[0]
	if !isIdentStart(b) {
		goto end
	}

	for i = 1; i < len(s); i++ {
		b = s[i]
		if !isIdentPart(b) {
			goto end
		}
	}

	ok = true

end:
	return ok
}

func isIdentStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

func isIdentPart(b byte) bool {
	return isIdentStart(b) || (b >= '0' && b <= '9')
}
