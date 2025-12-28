package syncdoterr

import (
	"bytes"
	"os"
	"sort"

	"github.com/mikeschinkel/go-cliutil"
	"github.com/mikeschinkel/go-dt"
)

type SyncArgs struct {
	SourceDir dt.DirPath
	RootDir   dt.DirPath
	Writer    cliutil.Writer
	Prompter  Prompter
}

type SyncResult struct {
	Updated   int
	Skipped   int
	Cancelled bool
	Targets   int
}

func SyncDoterr(args *SyncArgs) (result *SyncResult, err error) {
	var sourceFile dt.Filepath
	var sourceExists bool
	var rootExists bool
	var targets []dt.Filepath
	var confirm bool
	var kvs []ErrKV

	if args == nil {
		panic("Missing SyncArgs")
	}
	if args.Writer == nil {
		panic("Missing SyncArgs.Writer")
	}
	if args.Prompter == nil {
		panic("Missing SyncArgs.Prompter")
	}

	sourceFile = dt.FilepathJoin(args.SourceDir, DotErrFilename)
	kvs = AppendKV(kvs, func() ErrKV {
		return AnyKV("doterr_file", toTilde(sourceFile))
	})

	sourceExists, err = sourceFile.Exists()
	if err != nil {
		err = NewErr(ErrSync, ErrSource, err)
		goto end
	}
	if !sourceExists {
		err = NewErr(ErrSource, dt.ErrFileNotExists)
		goto end
	}

	kvs = AppendKV(kvs, func() ErrKV {
		return AnyKV("target_dir", toTilde(args.RootDir))
	})
	rootExists, err = args.RootDir.Exists()
	if err != nil {
		err = NewErr(ErrTarget, err)
		goto end
	}
	if !rootExists {
		err = NewErr(ErrTarget, dt.ErrDirDoesNotExist)
		goto end
	}

	args.Writer.Printf("\nScanning %s for %s files...\n",
		toTilde(args.RootDir),
		DotErrFilename,
	)
	targets, err = scanTargets(args.RootDir, sourceFile)
	if err != nil {
		err = NewErr(ErrScan, err)
		goto end
	}

	result = &SyncResult{
		Targets: len(targets),
	}

	if len(targets) == 0 {
		args.Writer.Printf("No doterr.go files found under %s\n", args.RootDir)
		goto end
	}

	args.Writer.Printf("\nSource file:\n  %s\n", toTilde(sourceFile))
	args.Writer.Printf("\nTarget files (%d):\n", len(targets))
	for _, target := range targets {
		args.Writer.Printf("  %s\n", toTilde(target))
	}

	confirm, err = args.Prompter.Confirm("\nUpdate these files? (y/N): ")
	if err != nil {
		err = NewErr(ErrPrompt, err)
		goto end
	}
	if !confirm {
		args.Writer.Printf("Cancelled.\n\n")
		result.Cancelled = true
		goto end
	}

	result.Updated, result.Skipped, err = updateTargets(&updateArgs{
		SourceFile: sourceFile,
		Targets:    targets,
		Writer:     args.Writer,
	})
	if err != nil {
		goto end
	}

end:
	if err != nil {
		err = WithErr(err, ErrSync, kvs)
	}
	return result, err
}

func scanTargets(rootDir dt.DirPath, sourceFile dt.Filepath) (targets []dt.Filepath, err error) {
	var entry dt.DirEntry
	var entryErr error
	var file dt.Filepath
	var sourceAbs dt.Filepath
	var targetAbs dt.Filepath

	sourceAbs, err = sourceFile.Abs()
	if err != nil {
		goto end
	}

	for entry, entryErr = range rootDir.WalkFiles() {
		if entryErr != nil {
			err = entryErr
			goto end
		}

		if entry.Filename() != DotErrFilename {
			continue
		}

		file = entry.Filepath()
		targetAbs, err = file.Abs()
		if err != nil {
			goto end
		}

		if targetAbs == sourceAbs {
			continue
		}

		targets = append(targets, file)
	}

	sort.Slice(targets, func(i, j int) bool {
		return string(targets[i]) < string(targets[j])
	})

end:
	return targets, err
}

func updateTargets(args *updateArgs) (updated, skipped int, err error) {
	var sourceData []byte
	var targetData []byte
	var target dt.Filepath
	var pkgName string
	var transformed []byte
	var info os.FileInfo

	sourceData, err = args.SourceFile.ReadFile()
	if err != nil {
		err = NewErr(ErrSource,
			"path", args.SourceFile,
			err,
		)
		goto end
	}

	for _, target = range args.Targets {
		var perm os.FileMode

		pkgName, err = detectPackageName(target)
		if err != nil {
			//goland:noinspection GoDirectComparisonOfErrors
			// These are known sentinel errors, they will not be wrapped
			switch err {
			case ErrOnlyDotErrFilesFound:
				fallthrough
			case ErrPackageNameNotFound:
				pkgName = string(target.Dir().Base())
				//goland:noinspection GoMaybeNil
				warning(args.Writer, "Could not detect package name in %s:\n\t- REASON: %s.\n\t- USING:  '%s' instead.",
					toTilde(target),
					err.Error(),
					pkgName,
				)
				err = nil
			default:
				err = NewErr(ErrPackageName,
					"doterr_file", toTilde(target),
					err,
				)
				goto end
			}
		}

		transformed, err = replacePackageLine(sourceData, pkgName)
		if err != nil {
			err = NewErr(ErrSource,
				"path", args.SourceFile,
				err,
			)
			goto end
		}

		targetData, err = target.ReadFile()
		if err != nil {
			warning(args.Writer, "Could not read target %s", toTilde(target))
			err = nil
		}

		if bytes.Equal(transformed, targetData) {
			args.Writer.Printf("\n  %s skipped", toTilde(target))
			skipped++
			continue
		}

		info, err = target.Stat()
		switch {
		case err != nil:
			warning(args.Writer, "Could not stat target %s", toTilde(target))
			err = nil
			perm = os.ModePerm
		default:
			perm = info.Mode().Perm()
		}

		err = target.WriteFile(transformed, perm)
		if err != nil {
			err = NewErr(ErrTarget,
				"path", target,
				err,
			)
			goto end
		}

		args.Writer.Printf("\n  %s updated", toTilde(target))

		updated++
	}

end:
	return updated, skipped, err
}

type updateArgs struct {
	SourceFile dt.Filepath
	Targets    []dt.Filepath
	Writer     cliutil.Writer
}
