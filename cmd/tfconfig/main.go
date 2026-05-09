package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/OpenUdon/tfconfig"
)

func main() {
	configDir := flag.String("config-dir", ".", "Terraform/OpenTofu configuration directory")
	compact := flag.Bool("compact", false, "emit compact JSON")
	flag.Parse()

	doc, err := tfconfig.LoadDir(*configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tfconfig: %v\n", err)
		os.Exit(1)
	}

	var out []byte
	if *compact {
		out, err = doc.JSON()
	} else {
		out, err = doc.JSONIndent("", "  ")
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "tfconfig: encode JSON: %v\n", err)
		os.Exit(1)
	}
	out = append(out, '\n')
	if _, err := os.Stdout.Write(out); err != nil {
		fmt.Fprintf(os.Stderr, "tfconfig: write JSON: %v\n", err)
		os.Exit(1)
	}

	if hasErrorDiagnostics(doc.Diagnostics) {
		os.Exit(2)
	}
	for _, mod := range doc.Modules {
		if hasErrorDiagnostics(mod.Diagnostics) {
			os.Exit(2)
		}
	}
}

func hasErrorDiagnostics(diags []tfconfig.Diagnostic) bool {
	for _, diag := range diags {
		if diag.Severity == tfconfig.DiagnosticError {
			return true
		}
	}
	return false
}
