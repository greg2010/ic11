package cmd

import (
	"os"

	"github.com/greg2010/ic11/internal/filereader"
	"github.com/greg2010/ic11/internal/ic11"
	"github.com/greg2010/ic11/internal/printer"
	"github.com/spf13/cobra"
)

var emitLabels bool
var precomputeExprs bool
var optimizeJumps bool
var propagateVars bool
var optimize bool
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

		printer.PrintVerbose("compilation successful")
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
	if optimize {
		conf = ic11.AllCompilerOpts()
		printer.PrintVerboseln("using all compiler optimizations")
	} else {
		conf = ic11.NoCompilerOpts()
		printer.PrintVerboseln("disabling compiler optimizations")
	}

	if emitLabels {
		conf.OptimizeLabels = false
	}

	if !precomputeExprs {
		conf.PrecomputeExprs = false
	}

	if !optimizeJumps {
		conf.OptimizeJumps = false
	}

	if !propagateVars {
		conf.PropagateVariables = false
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
	rootCmd.Flags().BoolVar(&precomputeExprs, "precompute-exprs", true, "Precompute expressions at compile-time. If set to true, computes simple expressions ahead of time.")
	rootCmd.Flags().BoolVar(&optimizeJumps, "optimize-jumps", true, "Emit special jump instructions (bne bgt etc).")
	rootCmd.Flags().BoolVar(&propagateVars, "propagate-vars", true, "Propagate known variables to reduce the number of register allocations.")
	rootCmd.Flags().BoolVarP(&optimize, "optimize", "O", true, "Enable all compiler optimizations (invidiual optimizations can be overriden with specific flags).")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging.")
	rootCmd.Flags().StringVarP(&out, "out", "o", "a.out", "Filename to write output to.")
}
