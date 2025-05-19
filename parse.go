package fuzzypatch

import (
	"fmt"
	"iter"
	"strconv"
	"strings"
)

type tokenType int

const (
	startSearchType   tokenType = iota // "<<<<<<< SEARCH line:n
	textSeparatorType                  // "======="
	endReplaceType                     // ">>>>>>> REPLACE
	textType                           // any other line (incl. blank)
	invalidType
	EOF
)

const (
	startSearchPrefix = "<<<<<<< SEARCH"
	textSeparator     = "======="
	endReplace        = ">>>>>>> REPLACE"
)

func tokenTypeString(typ tokenType) string {
	switch typ {
	case startSearchType:
		return "StartSearchType: " + startSearchPrefix + " line:n"
	case textSeparatorType:
		return "TextSeparatorType: " + textSeparator
	case endReplaceType:
		return "EndReplaceType: " + endReplace
	case textType:
		return "TextType"
	case invalidType:
		return "InvalidType"
	case EOF:
		return "EOF"
	default:
		return "Unknown"
	}
}

type token struct {
	Type tokenType
	Line int
	Text string
}

func tokenize(input string) iter.Seq[token] {
	return func(yield func(token) bool) {
		lineNo := 0
		for line := range strings.Lines(input) {
			trim := strings.TrimRight(line, "\r\n")
			switch {
			case strings.HasPrefix(line, startSearchPrefix):
				if !yield(token{startSearchType, lineNo, line}) {
					return
				}
			case trim == textSeparator:
				if !yield(token{textSeparatorType, lineNo, line}) {
					return
				}
			case trim == endReplace:
				if !yield(token{endReplaceType, lineNo, line}) {
					return
				}
			default:
				if !yield(token{textType, lineNo, line}) {
					return
				}
			}
			lineNo++
		}
		yield(token{EOF, 0, ""})
	}
}

type parser struct {
	current token
	next    func() (token, bool)
}

func (p *parser) read() token {
	current := p.current
	tok, ok := p.next()
	if !ok {
		p.current = token{Type: invalidType}
	} else {
		p.current = tok
	}
	return current
}

func (p *parser) expect(typ tokenType) (token, error) {
	tok := p.read()
	if tok.Type != typ {
		return token{}, fmt.Errorf("expected %s, got %s: %q (line %d))",
			tokenTypeString(typ),
			tokenTypeString(tok.Type),
			tok.Text,
			tok.Line,
		)
	}
	return tok, nil
}

func (p *parser) parseStartSearch() (int, error) {
	for p.current.Type == textType && strings.TrimSpace(p.current.Text) == "" {
		p.read()
	}
	tok, err := p.expect(startSearchType)
	if err != nil {
		return 0, err
	}
	suffix, _ := strings.CutPrefix(tok.Text, startSearchPrefix)
	lineStr, ok := strings.CutPrefix(strings.TrimSpace(suffix), "line:")
	if !ok {
		return 0, fmt.Errorf("expected %s, got %q", tokenTypeString(startSearchType), tok.Text)
	}
	line, err := strconv.Atoi(lineStr)
	if err != nil {
		return 0, fmt.Errorf("expected %s, got %q: %w", tokenTypeString(startSearchType), tok.Text, err)
	}
	return line, nil
}

func (p *parser) parseDiff() (Diff, error) {
	var diff Diff
	var err error
	diff.Line, err = p.parseStartSearch()
	if err != nil {
		return Diff{}, err
	}
	for p.current.Type == textType {
		diff.Search += p.current.Text
		p.read()
	}
	if _, err := p.expect(textSeparatorType); err != nil {
		return Diff{}, err
	}
	for p.current.Type == textType {
		diff.Replace += p.current.Text
		p.read()
	}
	if _, err := p.expect(endReplaceType); err != nil {
		return Diff{}, nil
	}
	return diff, nil
}

func Parse(input string) ([]Diff, error) {
	next, stop := iter.Pull(tokenize(input))
	defer stop()
	p := parser{next: next}
	p.read()
	var diffs []Diff
	for p.current.Type != EOF {
		diff, err := p.parseDiff()
		if err != nil {
			return nil, err
		}
		diffs = append(diffs, diff)
	}
	return diffs, nil
}
