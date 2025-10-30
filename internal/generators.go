package internal

import (
	"go/parser"
	"go/token"
)

func GetDocComment(src string) (string, error) {
	cli := token.NewFileSet()
	ast, err := parser.ParseFile(cli, src, nil, parser.ParseComments)
	if err != nil {
		return "", err
	}

	var doccomment string
	// Don't use .Text() as it messes up whitespace by shrinking multiple empty lines
	for _, line := range ast.Doc.List {
		doccomment += line.Text
	}
	return doccomment, nil
}
