package syncdoterr

import "errors"

var (
	ErrCmd                = errors.New("cmd error")
	ErrSync               = errors.New("sync error")
	ErrSource             = errors.New("source error")
	ErrTarget             = errors.New("target error")
	ErrScan               = errors.New("scan error")
	ErrPackageName        = errors.New("package name error")
	ErrPrompt             = errors.New("prompt error")
	ErrInvalidArgs        = errors.New("invalid args error")
	ErrMissingPackageDecl = errors.New("missing package declaration error")
)
