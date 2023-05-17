package printer

// Printer is a generic printer interface
type Printer interface {
	Print(i ...interface{})
	Printf(format string, i ...interface{})
	Println(i ...interface{})
	PrintVerbose(i ...interface{})
	PrintVerbosef(format string, i ...interface{})
	PrintVerboseln(i ...interface{})
	PrintError(i ...interface{})
	PrintErrorf(format string, i ...interface{})
	PrintErrorln(i ...interface{})
}
