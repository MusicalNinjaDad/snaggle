package internal

import (
	"go/parser"
	"go/token"
)

type DocComment struct {
	Text string
}

func GetDocComment(src string) (DocComment, error) {
	cli := token.NewFileSet()
	ast, err := parser.ParseFile(cli, src, nil, parser.ParseComments)
	if err != nil {
		return DocComment{}, err
	}

	var doccomment string
	// Don't use .Text() as it messes up whitespace by shrinking multiple empty lines
	for _, line := range ast.Doc.List {
		doccomment += line.Text
	}
	return DocComment{Text: doccomment}, nil
}
