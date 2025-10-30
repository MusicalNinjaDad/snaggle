package internal

import (
	"bufio"
	"go/parser"
	"go/token"
	"io"
	"os"
	"strings"
)

type docComment struct {
	Text  string // Exact contents as multi-line string
	Start token.Position
	End   token.Position
}

func getDocComment(src string) (docComment, error) {
	cli := token.NewFileSet()
	ast, err := parser.ParseFile(cli, src, nil, parser.ParseComments)
	if err != nil {
		return docComment{}, err
	}

	var doccomment string
	// Don't use .Text() as it messes up whitespace by shrinking multiple empty lines
	for _, line := range ast.Doc.List {
		doccomment += line.Text
	}

	startpos := cli.Position(ast.Doc.Pos())
	endpos := cli.Position(ast.Doc.End())
	return docComment{Text: doccomment, Start: startpos, End: endpos}, nil
}

func SetDocComment(path string, comment string) error {
	comment = "/*\n" + comment + "*/\n"
	oldComment, err := getDocComment(path)
	if err != nil {
		return err
	}

	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	origContents, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	newContents := string(origContents[:oldComment.Start.Offset])
	newContents += comment
	newContents += string(origContents[oldComment.End.Offset+1:])

	return os.WriteFile(path, []byte(newContents), 0)
}

func ReplaceBetween(original io.Reader, marker1 string, marker2 string, replacement []byte) []byte {
	updated := ""
	scanner := bufio.NewScanner(original)
	wanted := true
	for scanner.Scan() {
		if wanted {
			updated += scanner.Text() + "\n"
			if strings.Contains(scanner.Text(), marker1) {
				wanted = false
				updated += string(replacement)
			}
		} else {
			if strings.Contains(scanner.Text(), marker2) {
				wanted = true
				updated += scanner.Text() + "\n"
			}
		}
	}
	return []byte(updated)
}
