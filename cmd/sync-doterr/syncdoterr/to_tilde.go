package syncdoterr

import (
	"github.com/mikeschinkel/go-dt"
)

type tildable interface {
	dt.DirPath | dt.Filepath | dt.EntryPath
}

func toTilde[T tildable](path T) dt.TildeEntryPath {
	return dt.EntryPathToTilde(dt.EntryPath(path), dt.OrFullPath)
}
