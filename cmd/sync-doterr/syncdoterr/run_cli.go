package syncdoterr

import (
	"errors"
	"os"

	"github.com/mikeschinkel/go-cliutil"
	"github.com/mikeschinkel/go-dt"
)

func RunCLI() {
	var err error
	var globalOptions *cliutil.GlobalOptions
	var args []string
	var writer cliutil.Writer
	var rootArg string
	var rootDir dt.DirPath
	var cwd dt.DirPath
	var result *SyncResult
	var prompter Prompter

	err = dt.EnsureUserHomeDir()
	if err != nil {
		goto end
	}

	globalOptions, args, err = cliutil.ParseGlobalOptions(os.Args)
	if err != nil {
		cliutil.Stderrf("Invalid option(s): %v\n", err)
		os.Exit(cliutil.ExitOptionsParseError)
	}

	writer = cliutil.NewWriter(&cliutil.WriterArgs{
		Quiet:     globalOptions.Quiet(),
		Verbosity: globalOptions.Verbosity(),
	})

	if len(args) != 1 {
		err = NewErr(ErrInvalidArgs,
			"arg_count", len(args),
			"want", 1,
		)
		goto end
	}

	rootArg = args[0]
	rootDir, err = dt.ParseDirPath(rootArg)
	if err != nil {
		err = NewErr(ErrInvalidArgs, dt.ErrInvalidDirectory,
			"root_arg", rootArg,
			err,
		)
		goto end
	}

	cwd, err = dt.Getwd()
	if err != nil {
		err = NewErr(ErrCmd, dt.ErrAccessingWorkingDir, err)
		goto end
	}

	prompter = NewStdinPrompter(writer)

	result, err = SyncDoterr(&SyncArgs{
		SourceDir: cwd,
		RootDir:   rootDir,
		Writer:    writer,
		Prompter:  prompter,
	})
	if err != nil {
		goto end
	}

	if result != nil {
		if result.Cancelled {
			goto end
		}
		writer.Printf("\n\n")
		writer.Printf("Updated: %d\n", result.Updated)
		writer.Printf("Skipped: %d", result.Skipped)
		if result.Skipped != 0 {
			writer.Printf(" (skipped files were already up-to-date unless warning provided)\n")
		}
		writer.Printf("\n")
	}

end:
	if err != nil {
		if errors.Is(err, ErrInvalidArgs) {
			writer.Errorf("Usage: sync-doterr <root-dir>\n")
		}
		writer.Errorf("Error: %v\n", err)
		os.Exit(cliutil.ExitUnknownRuntimeError)
	}
}
