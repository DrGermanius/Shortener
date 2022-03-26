// package main of staticlint provides code analyzing entrypoint
package main

import (
	"github.com/DrGermanius/exitanalyser"
	"github.com/go-critic/go-critic/checkers/analyzer"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// main provides checkers for analyzing code using `go run main.go directory`
func main() {
	var analyzers []*analysis.Analyzer

	analyzers = append(analyzers, asmdecl.Analyzer,

		// check for useless assignments
		assign.Analyzer,

		// check for common mistakes using the sync/atomic package
		atomic.Analyzer,

		// check for common mistakes involving boolean operators
		bools.Analyzer,

		// check that +build tags are well-formed and correctly located
		buildtag.Analyzer,

		// detect some violations of the cgo pointer passing rules
		cgocall.Analyzer,

		// check for unkeyed composite literals
		composite.Analyzer,

		// check for locks erroneously passed by value
		copylock.Analyzer,

		// report passing non-pointer or non-error values to errors.As
		errorsas.Analyzer,

		// report assembly that clobbers the frame pointer before saving it
		framepointer.Analyzer,

		// check for mistakes using HTTP responses
		httpresponse.Analyzer,

		// detect impossible interface-to-interface type assertions
		ifaceassert.Analyzer,

		// check references to loop variables from within nested functions
		loopclosure.Analyzer,

		// check cancel func returned by context.WithCancel is called
		lostcancel.Analyzer,

		// check for useless comparisons between functions and nil
		nilfunc.Analyzer,

		// check consistency of Printf format strings and arguments
		printf.Analyzer,

		// check for shifts that equal or exceed the width of the integer
		shift.Analyzer,

		// check for unbuffered channel of os.Signal
		sigchanyzer.Analyzer,

		// check signature of methods of well-known interfaces
		stdmethods.Analyzer,

		// check for string(int) conversions
		stringintconv.Analyzer,

		// check that struct field tags conform to reflect.StructTag.Get
		structtag.Analyzer,

		// check for common mistaken usages of tests and examples
		tests.Analyzer,

		// report calls to (*testing.T).Fatal from goroutines started by a test
		testinggoroutine.Analyzer,

		// report passing non-pointer or non-interface values to unmarshal
		unmarshal.Analyzer,

		// check for unreachable code
		unreachable.Analyzer,

		// check for invalid conversions of uintptr to unsafe.Pointer
		unsafeptr.Analyzer,

		// check for unused results of calls to some functions
		unusedresult.Analyzer,

		// the most opinionated Go source code linter
		analyzer.Analyzer,

		// check for os.Exit calls
		exitanalyser.Analyzer,
	)

	for _, v := range simple.Analyzers {
		analyzers = append(analyzers, v)
	}
	for _, v := range staticcheck.Analyzers {
		analyzers = append(analyzers, v)
	}
	for _, v := range stylecheck.Analyzers {
		analyzers = append(analyzers, v)
	}

	multichecker.Main(analyzers...)
}
