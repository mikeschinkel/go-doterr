package syncdoterr

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/mikeschinkel/go-cliutil"
)

type Prompter interface {
	Confirm(prompt string) (bool, error)
}

type StdinPrompter struct {
	reader *bufio.Reader
	writer cliutil.Writer
}

func NewStdinPrompter(writer cliutil.Writer) *StdinPrompter {
	return &StdinPrompter{
		reader: bufio.NewReader(os.Stdin),
		writer: writer,
	}
}

func (p *StdinPrompter) Confirm(prompt string) (ok bool, err error) {
	var line string
	var readErr error

	p.writer.Printf("%s", prompt)

	line, readErr = p.reader.ReadString('\n')
	if readErr != nil {
		if readErr != io.EOF {
			err = readErr
			goto end
		}
	}

	line = strings.TrimSpace(line)
	if strings.EqualFold(line, "y") || strings.EqualFold(line, "yes") {
		ok = true
		goto end
	}

end:
	return ok, err
}
