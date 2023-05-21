package cmd

import (
	"os"

	"github.com/greg2010/ic11/internal/filereader"
	"github.com/greg2010/ic11/internal/ic11"
	"github.com/greg2010/ic11/internal/printer"
	"github.com/spf13/cobra"
)

var emitLabels bool
var noExprOpt bool
var noJumpOpt bool
var noVarOpt bool
var noDeviceAliases bool
var noComputeHashes bool
var optimize int
var verbose bool
var out string
var rootCmd = &cobra.Command{
	Use:   "ic11c file1 file2",
	Short: "A µC -> MIPS compiler",
	Long: `ic11c is a compiler for µC to MIPS dialect used by the game Stationeers.
run ic11c help for details on how to use it.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.SetOut(os.Stdout)
		printer := printer.NewCobraPrinter(cmd, verbose)
		if len(args) == 0 {
			printer.PrintErrorln("no input files")
			os.Exit(1)
		}

		reader, err := filereader.New(args...)
		if err != nil {
			printer.PrintErrorln(err)
			os.Exit(1)
		}
		defer reader.Close()

		conf := getCompilerConfig(printer)

		compiler, err := ic11.NewCompiler(reader.GetReaders(), conf, printer)
		if err != nil {
			printer.PrintErrorf("parsing failed: %v\n", err)
			os.Exit(1)
		}

		compiled, err := compiler.Compile()
		if err != nil {
			printer.PrintErrorf("compilation terminated due to error: %v\n", err)
			os.Exit(1)
		}

		err = writeToFile(out, compiled)
		if err != nil {
			printer.PrintErrorln(err)
			os.Exit(1)
		}

		printer.PrintVerboseln("compilation successful")
	},
}

func writeToFile(fname string, contents string) error {
	file, err := os.Create(fname)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(contents)
	return err
}

func getCompilerConfig(printer printer.Printer) ic11.CompilerOpts {
	var conf ic11.CompilerOpts
	if optimize == 2 {
		conf = ic11.AllCompilerOpts()
		printer.PrintVerboseln("using all compiler optimizations")
	} else {
		conf = ic11.NoCompilerOpts()
		printer.PrintVerboseln("disabling compiler optimizations")
	}

	if emitLabels {
		conf.OptimizeLabels = false
	}

	if noExprOpt {
		conf.PrecomputeExprs = false
	}

	if noJumpOpt {
		conf.OptimizeJumps = false
	}

	if noVarOpt {
		conf.PropagateVariables = false
	}

	if noComputeHashes {
		conf.PrecomputeHashes = false
	}

	return conf
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		rootCmd.PrintErr(err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolVar(&emitLabels, "emit-labels", false, "Emit labels. If set to false, replaces all labels with absolute addresses.")
	rootCmd.Flags().BoolVar(&noExprOpt, "no-expr-opt", false, "Dp not precompute expressions at compile-time.")
	rootCmd.Flags().BoolVar(&noJumpOpt, "no-jump-opt", false, "Do not emit special jump instructions (bne bgt etc).")
	rootCmd.Flags().BoolVar(&noVarOpt, "no-var-opt", false, "Do not propagate known variables to reduce the number of register allocations.")
	rootCmd.Flags().BoolVar(&noDeviceAliases, "no-device-aliases", false, "Do not emit device alias instructions.")
	rootCmd.Flags().BoolVar(&noComputeHashes, "no-compute-hashes", false, "Do not precompute hashes at compile time.")
	rootCmd.Flags().IntVarP(&optimize, "optimize", "O", 2, "Set optimization level preset. 0 -- no optimizations, 2 -- full optimization.")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging.")
	rootCmd.Flags().StringVarP(&out, "out", "o", "a.out", "Filename to write output to.")
}
