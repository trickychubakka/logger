// Package staticlint -- multichecker, статический анализатор.
// Файл analysistyperegister.go, создание analysisTypeRegistry -- реестра golang.org/x/tools/go/analysis/passes checker-ов.
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stdversion"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"golang.org/x/tools/go/analysis/passes/waitgroup"
)

type analysisTypeRegistry map[string]*analysis.Analyzer

func createAnalysisTypesRegistry() (analysisTypeRegistry, error) {
	var typeRegistry = make(map[string]*analysis.Analyzer)

	typeRegistry["appends"] = appends.Analyzer
	typeRegistry["asmdecl"] = asmdecl.Analyzer
	typeRegistry["assign"] = assign.Analyzer
	typeRegistry["atomic"] = atomic.Analyzer
	typeRegistry["atomicalign"] = atomicalign.Analyzer
	typeRegistry["bools"] = bools.Analyzer
	typeRegistry["buildssa"] = buildssa.Analyzer
	typeRegistry["buildtag"] = buildtag.Analyzer
	typeRegistry["cgocall"] = cgocall.Analyzer
	typeRegistry["composite"] = composite.Analyzer
	typeRegistry["copylock"] = copylock.Analyzer
	typeRegistry["ctrlflow"] = ctrlflow.Analyzer
	typeRegistry["deepequalerrors"] = deepequalerrors.Analyzer
	typeRegistry["defers"] = defers.Analyzer
	typeRegistry["directive"] = directive.Analyzer
	typeRegistry["errorsas"] = errorsas.Analyzer
	typeRegistry["fieldalignment"] = fieldalignment.Analyzer
	typeRegistry["findcall"] = findcall.Analyzer
	typeRegistry["framepointer"] = framepointer.Analyzer
	typeRegistry["httpmux"] = httpmux.Analyzer
	typeRegistry["httpresponse"] = httpresponse.Analyzer
	typeRegistry["ifaceassert"] = ifaceassert.Analyzer
	typeRegistry["inspect"] = inspect.Analyzer
	typeRegistry["loopclosure"] = loopclosure.Analyzer
	typeRegistry["lostcancel"] = lostcancel.Analyzer
	typeRegistry["nilfunc"] = nilfunc.Analyzer
	typeRegistry["nilness"] = nilness.Analyzer
	typeRegistry["pkgfact"] = pkgfact.Analyzer
	typeRegistry["printf"] = printf.Analyzer
	typeRegistry["reflectvaluecompare"] = reflectvaluecompare.Analyzer
	typeRegistry["shadow"] = shadow.Analyzer
	typeRegistry["shift"] = shift.Analyzer
	typeRegistry["sigchanyzer"] = sigchanyzer.Analyzer
	typeRegistry["slog"] = slog.Analyzer
	typeRegistry["sortslice"] = sortslice.Analyzer
	typeRegistry["stdmethods"] = stdmethods.Analyzer
	typeRegistry["stdversion"] = stdversion.Analyzer
	typeRegistry["stringintconv"] = stringintconv.Analyzer
	typeRegistry["structtag"] = structtag.Analyzer
	typeRegistry["testinggoroutine"] = testinggoroutine.Analyzer
	typeRegistry["tests"] = tests.Analyzer
	typeRegistry["timeformat"] = timeformat.Analyzer
	typeRegistry["unmarshal"] = unmarshal.Analyzer
	typeRegistry["unreachable"] = unreachable.Analyzer
	typeRegistry["unsafeptr"] = unsafeptr.Analyzer
	typeRegistry["unusedresult"] = unusedresult.Analyzer
	typeRegistry["unusedwrite"] = unusedwrite.Analyzer
	typeRegistry["usesgenerics"] = usesgenerics.Analyzer
	typeRegistry["waitgroup"] = waitgroup.Analyzer

	return typeRegistry, nil
}
