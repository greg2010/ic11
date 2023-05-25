package printer

import (
	"fmt"

	"github.com/spf13/cobra"
)

// CobraPrinter is as concrete Printer interface to be used with Cobra framework
type CobraPrinter struct {
	cmd     *cobra.Command
	verbose bool
}

func (cp *CobraPrinter) Print(i ...interface{}) {
	cp.cmd.Print(i...)
}
func (cp *CobraPrinter) Printf(format string, i ...interface{}) {
	cp.cmd.Printf(format, i...)
}

func (cp *CobraPrinter) Println(i ...interface{}) {
	cp.cmd.Println(i...)
}

func (cp *CobraPrinter) PrintVerbose(i ...interface{}) {
	if !cp.verbose {
		return
	}

	cp.cmd.Print(i...)
}
func (cp *CobraPrinter) PrintVerbosef(format string, i ...interface{}) {
	if !cp.verbose {
		return
	}

	cp.cmd.Printf(format, i...)
}

func (cp *CobraPrinter) PrintVerboseln(i ...interface{}) {
	if !cp.verbose {
		return
	}

	cp.cmd.Println(i...)
}

func (cp *CobraPrinter) PrintError(i ...interface{}) {
	cp.cmd.PrintErrf("error: %v", i...)
}

func (cp *CobraPrinter) PrintErrorf(format string, i ...interface{}) {
	form := fmt.Sprintf("error: %s", format)
	cp.cmd.PrintErrf(form, i...)
}

func (cp *CobraPrinter) PrintErrorln(i ...interface{}) {
	cp.cmd.PrintErrf("error: %v\n", i...)
}

func NewCobraPrinter(cmd *cobra.Command, printVerbose bool) *CobraPrinter {
	return &CobraPrinter{
		cmd:     cmd,
		verbose: printVerbose,
	}
}
