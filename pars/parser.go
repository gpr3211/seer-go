package pars

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Helper functions to convert AST nodes to readable strings

var FunctionMap = make(map[string]FunctionDetails)

func ParseDir(dirPath string) error {

	fset := token.NewFileSet()
	// Map to store all functions across files

	// Walk through all directories and find Go files
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		fmt.Printf("Analyzing file: %s\n", path)

		// Read and parse the file
		src, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", path, err)
			return nil // Continue with other files
		}

		// Parse the source code into an AST
		file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
		if err != nil {
			fmt.Printf("Error parsing file %s: %v\n", path, err)
			return nil // Continue with other files
		}

		// Create an AST visitor to analyze the code
		visitor := &CodeAnalyzer{
			Functions: make(map[string]FunctionDetails),
			Fset:      fset,
		}

		// Visit all nodes in the AST
		ast.Walk(visitor, file)

		// Add functions to the global map with package and file info
		packageName := file.Name.Name
		for name, details := range visitor.Functions {
			fullName := fmt.Sprintf("%s.%s (%s)", packageName, name, path)
			FunctionMap[fullName] = details
		}

		return nil
	})
	return err
}

func exprToString(fset *token.FileSet, expr ast.Expr) string {
	if expr == nil {
		return "<nil>"
	}
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, expr)
	return buf.String()
}

func StmtToString(fset *token.FileSet, stmt ast.Stmt) string {
	if stmt == nil {
		return "<nil>"
	}

	var buf bytes.Buffer

	// Using type assertion to handle specific types of statements
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		buf.WriteString("Assignment: ")
		for _, lhs := range s.Lhs {
			buf.WriteString(fmt.Sprintf("%s ", lhs))
		}
		buf.WriteString(fmt.Sprintf("= "))
		for _, rhs := range s.Rhs {
			buf.WriteString(fmt.Sprintf("%s ", rhs))
		}
	case *ast.DeclStmt:
		buf.WriteString("Declaration: ")
		printer.Fprint(&buf, fset, s.Decl)
	case *ast.ExprStmt:
		buf.WriteString("Expression: ")
		printer.Fprint(&buf, fset, s.X)
	case *ast.IfStmt:
		buf.WriteString("If Condition: ")
		printer.Fprint(&buf, fset, s.Cond)
	case *ast.ForStmt:
		buf.WriteString("For Loop: Init: ")
		printer.Fprint(&buf, fset, s.Init)
		buf.WriteString("; Cond: ")
		printer.Fprint(&buf, fset, s.Cond)
		buf.WriteString("; Post: ")
		printer.Fprint(&buf, fset, s.Post)
	default:
		// Default case: generic print if statement type is unhandled
		printer.Fprint(&buf, fset, stmt)
	}

	return buf.String()
}

/*
	func funcTypeToString(fset *token.FileSet, funcType *ast.FuncType) string {
		var buf bytes.Buffer
		printer.Fprint(&buf, fset, funcType)
		return buf.String()
	}
*/
func FuncTypeToString(fset *token.FileSet, funcType *ast.FuncType) string {
	var buf bytes.Buffer

	// Extract parameters
	buf.WriteString("Parameters: ")
	for _, field := range funcType.Params.List {
		buf.WriteString(field.Names[0].Name + " " + fmt.Sprint(field.Type) + ", ")
	}

	// Extract return types
	if funcType.Results != nil {
		buf.WriteString(" | Returns: ")
		for _, result := range funcType.Results.List {
			buf.WriteString(fmt.Sprint(result.Type) + ", ")
		}
	}

	return buf.String()
}

// FunctionDetails stores detailed information about a function
type FunctionDetails struct {
	Doc       string            // Documentation comments
	Recv      string            // Receiver type if it's a method
	Name      string            // Function name
	Type      *ast.FuncType     // Function signature
	Variables []VariableDetails // Variables within the function
	Body      *ast.BlockStmt    // Function body}
}

// VariableDetails stores detailed information about a variable
type VariableDetails struct {
	Name string
	Type string
}

// CodeAnalyzer implements the ast.Visitor interface
type CodeAnalyzer struct {
	Functions map[string]FunctionDetails
	Fset      *token.FileSet
}

func (v *CodeAnalyzer) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	switch n := node.(type) {
	case *ast.FuncDecl:
		// Extract function details
		funcDetail := FunctionDetails{
			Doc:       extractDoc(n.Doc),
			Recv:      fieldListToString(v.Fset, n.Recv),
			Name:      n.Name.Name,
			Type:      n.Type,
			Variables: extractVariables(v.Fset, n.Body),
			Body:      n.Body}
		v.Functions[n.Name.Name] = funcDetail
	}

	return v
}
func extractDoc(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}
	return doc.Text()
}

func fieldListToString(fset *token.FileSet, fl *ast.FieldList) string {
	if fl == nil || len(fl.List) == 0 {
		return ""
	}
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, fl)
	return buf.String()
}
func extractVariables(fset *token.FileSet, body *ast.BlockStmt) []VariableDetails {
	var variables []VariableDetails

	ast.Inspect(body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			// Track variable declarations
			for i, expr := range stmt.Lhs {
				if ident, ok := expr.(*ast.Ident); ok {
					varType := "unknown"
					if i < len(stmt.Rhs) {
						varType = exprToString(fset, stmt.Rhs[i])
					}
					if ident.Name == "_" && varType == "uknown" {
						fmt.Println(" WARNING POTENTIAL IGNORED ERR", ident.NamePos)
					}
					variables = append(variables, VariableDetails{Name: ident.Name, Type: varType})
				}
			}
		}
		return true
	})

	return variables
}
