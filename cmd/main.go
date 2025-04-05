package main

import (
	"flag"
	"fmt"
	"github.com/gpr3211/seer-go/pars"
)

// Example code to analyze
func main() {
	// Create a FileSet to hold position information for parsed files

	dirPath := flag.String("dir", ".", "Directory path to analyze")
	flag.Parse()

	pars.ParseDir(*dirPath)
	fmt.Println("\n========== Analysis Results ==========")
	for name, details := range pars.FunctionMap {
		fmt.Printf("\nFunction: %s\n", name)
		fmt.Printf("Documentation: %s\n", details.Doc)

		if details.Recv != "" {
			fmt.Printf("Receiver: %s\n", details.Recv)
		}

		fmt.Println("Variables:")
		for _, variable := range details.Variables {
			fmt.Printf("  - Name: %s, \n", variable.Name)
		}
		// Optionally print the body (can be large)
		// fmt.Printf("Body: %s\n", stmtToString(fset, details.Body))
		fmt.Println("----------------------------------------")
	}

}
