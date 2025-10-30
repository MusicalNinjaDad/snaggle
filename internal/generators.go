package internal

import (
	"go/parser"
	"go/token"
)

type DocComment struct {
	Text  string // Exact contents as multi-line string
	Start token.Position
	End   token.Position
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

	startpos := cli.Position(ast.Doc.Pos())
	endpos := cli.Position(ast.Doc.End())
	return DocComment{Text: doccomment, Start: startpos, End: endpos}, nil
}
