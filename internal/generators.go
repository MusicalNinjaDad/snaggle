package internal

import (
	"go/parser"
	"go/token"
	"io"
	"os"
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

func SetDocComment(src string, comment string) error {
	oldComment, err := GetDocComment(src)
	if err != nil {
		return err
	}

	srcRO, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcRO.Close()

	origSrc, err := io.ReadAll(srcRO)
	if err != nil {
		return err
	}

	srcRW, err := os.OpenFile(src, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer srcRW.Close()

	newContents := "/*\n"
	newContents += comment
	newContents += "*/\n"
	newContents += string(origSrc[oldComment.End.Offset+1:])

	_, err = srcRW.WriteAt([]byte(newContents), int64(oldComment.Start.Offset))
	if err != nil {
		return err
	}

	return nil
}
