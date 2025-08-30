package tools

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type CodeAnalysisTool struct {
	BaseTool
}

func NewCodeAnalysisTool() *CodeAnalysisTool {
	return &CodeAnalysisTool{
		BaseTool: BaseTool{
			name:        "code_analysis",
			description: "Analyze Go code structure and provide insights",
		},
	}
}

func (c *CodeAnalysisTool) Execute(ctx context.Context, input string) (string, error) {
	parts := strings.SplitN(input, " ", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("usage: <operation> <code or file path>")
	}

	operation := parts[0]
	target := parts[1]

	switch operation {
	case "analyze":
		return c.analyzeCode(target)
	case "functions":
		return c.listFunctions(target)
	case "imports":
		return c.listImports(target)
	case "complexity":
		return c.calculateComplexity(target)
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

func (c *CodeAnalysisTool) analyzeCode(code string) (string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", code, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse code: %w", err)
	}

	var analysis []string
	analysis = append(analysis, fmt.Sprintf("Package: %s", node.Name.Name))

	importCount := len(node.Imports)
	analysis = append(analysis, fmt.Sprintf("Imports: %d", importCount))

	var funcCount, methodCount int
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Recv == nil {
				funcCount++
			} else {
				methodCount++
			}
		}
		return true
	})

	analysis = append(analysis, fmt.Sprintf("Functions: %d", funcCount))
	analysis = append(analysis, fmt.Sprintf("Methods: %d", methodCount))

	return strings.Join(analysis, "\n"), nil
}

func (c *CodeAnalysisTool) listFunctions(code string) (string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", code, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse code: %w", err)
	}

	var functions []string
	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			signature := c.getFunctionSignature(fn)
			functions = append(functions, signature)
		}
		return true
	})

	if len(functions) == 0 {
		return "No functions found", nil
	}

	return "Functions found:\n" + strings.Join(functions, "\n"), nil
}

func (c *CodeAnalysisTool) getFunctionSignature(fn *ast.FuncDecl) string {
	var signature strings.Builder

	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		recv := fn.Recv.List[0]
		if starExpr, ok := recv.Type.(*ast.StarExpr); ok {
			if ident, ok := starExpr.X.(*ast.Ident); ok {
				signature.WriteString(fmt.Sprintf("(*%s) ", ident.Name))
			}
		} else if ident, ok := recv.Type.(*ast.Ident); ok {
			signature.WriteString(fmt.Sprintf("(%s) ", ident.Name))
		}
	}

	signature.WriteString(fn.Name.Name)
	signature.WriteString("(")

	for i, param := range fn.Type.Params.List {
		if i > 0 {
			signature.WriteString(", ")
		}
		for j, name := range param.Names {
			if j > 0 {
				signature.WriteString(", ")
			}
			signature.WriteString(name.Name)
		}
		if len(param.Names) > 0 {
			signature.WriteString(" ")
		}
		signature.WriteString(c.typeToString(param.Type))
	}

	signature.WriteString(")")

	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
		signature.WriteString(" ")
		if len(fn.Type.Results.List) > 1 {
			signature.WriteString("(")
		}
		for i, result := range fn.Type.Results.List {
			if i > 0 {
				signature.WriteString(", ")
			}
			signature.WriteString(c.typeToString(result.Type))
		}
		if len(fn.Type.Results.List) > 1 {
			signature.WriteString(")")
		}
	}

	return signature.String()
}

func (c *CodeAnalysisTool) typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + c.typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + c.typeToString(t.Elt)
	case *ast.SelectorExpr:
		return c.typeToString(t.X) + "." + t.Sel.Name
	default:
		return "interface{}"
	}
}

func (c *CodeAnalysisTool) listImports(code string) (string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", code, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse code: %w", err)
	}

	if len(node.Imports) == 0 {
		return "No imports found", nil
	}

	var imports []string
	for _, imp := range node.Imports {
		path := imp.Path.Value
		if imp.Name != nil {
			imports = append(imports, fmt.Sprintf("%s %s", imp.Name.Name, path))
		} else {
			imports = append(imports, path)
		}
	}

	return "Imports:\n" + strings.Join(imports, "\n"), nil
}

func (c *CodeAnalysisTool) calculateComplexity(code string) (string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", code, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse code: %w", err)
	}

	complexities := make(map[string]int)

	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			complexity := c.cyclomaticComplexity(fn)
			complexities[fn.Name.Name] = complexity
		}
		return true
	})

	if len(complexities) == 0 {
		return "No functions to analyze", nil
	}

	var result []string
	result = append(result, "Cyclomatic Complexity:")
	for name, complexity := range complexities {
		level := "Low"
		if complexity > 10 {
			level = "High"
		} else if complexity > 5 {
			level = "Medium"
		}
		result = append(result, fmt.Sprintf("  %s: %d (%s)", name, complexity, level))
	}

	return strings.Join(result, "\n"), nil
}

func (c *CodeAnalysisTool) cyclomaticComplexity(fn *ast.FuncDecl) int {
	complexity := 1

	ast.Inspect(fn, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt:
			complexity++
		case *ast.CaseClause:
			complexity++
		}
		return true
	})

	return complexity
}
