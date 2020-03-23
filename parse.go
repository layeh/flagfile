package flagfile

// TODO: add support for escaping new line (\)

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strconv"
	"unicode"
)

// Init parses the given files, if they exist, and adds the flags after os.Args[1:].
// An error is printed and the program exits if there was an error parsing
// any of the files.
func Init(names ...string) {
	var newArgs []string
	for _, name := range names {
		args, err := ParseFile(name)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			println(err.Error())
			os.Exit(2)
		}
		newArgs = append(newArgs, args...)
	}
	if len(newArgs) > 0 {
		os.Args = append(os.Args[:1], append(newArgs, os.Args[1:]...)...)
	}
}

// ParseFile returns the input from the file into a slice of flags that can be
// passed to flag.(*FlagSet).Parse.
func ParseFile(name string) ([]string, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

type lexer struct {
	R *bufio.Reader

	Line   int
	Column int
}

type token uint8

const (
	tokenEOF token = iota + 1
	tokenComment
	tokenWhitespace
	tokenNewLine
	tokenWord
)

func (l *lexer) Next() (token, string, error) {
	for {
		r, _, err := l.R.ReadRune()
		if err != nil {
			if err == io.EOF {
				return tokenEOF, "", nil
			}
			return 0, "", err
		}
		if l.Column == 0 && r == '#' {
			// Comment line
			for {
				r, _, err := l.R.ReadRune()
				if err != nil {
					if err == io.EOF {
						return tokenComment, "", nil
					}
					return tokenComment, "", nil
				}
				if r == '\r' {
					r2, _, _ := l.R.ReadRune()
					if r2 != '\n' {
						l.R.UnreadRune()
					}
					l.Line++
					l.Column = 0
					return tokenComment, "", nil
				} else if r == '\n' {
					l.Line++
					l.Column = 0
					return tokenComment, "", nil
				} else {
					l.Column++
				}
			}
		}

		if r == '\r' {
			r2, _, _ := l.R.ReadRune()
			if r2 != '\n' {
				l.R.UnreadRune()
			}
			l.Line++
			l.Column = 0
			return tokenNewLine, "", nil
		}

		if r == '\n' {
			l.Line++
			l.Column = 0
			return tokenNewLine, "", nil
		}

		if r == '"' {
			l.R.UnreadRune()
			return l.NextQuotedWord()
		}

		if !unicode.IsSpace(r) {
			l.R.UnreadRune()
			return l.NextWord()
		}

		for {
			r, _, _ := l.R.ReadRune()
			if !unicode.IsSpace(r) || r == '\r' || r == '\n' {
				l.R.UnreadRune()
				return tokenWhitespace, "", nil
			}
			l.Column++
		}
	}
}

func (l *lexer) NextWord() (token, string, error) {
	var value bytes.Buffer

	for {
		r, _, err := l.R.ReadRune()
		if unicode.IsSpace(r) || r == '"' || err != nil {
			l.R.UnreadRune()
			return tokenWord, value.String(), nil
		}
		l.Column++
		value.WriteRune(r)
	}
}

func (l *lexer) NextQuotedWord() (token, string, error) {
	var value bytes.Buffer

	r, _, _ := l.R.ReadRune() // "
	l.Column++
	value.WriteRune(r)

	for {
		r, _, err := l.R.ReadRune()
		if err != nil {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			l.R.UnreadRune()
			return 0, "", err
		}
		if r == '\n' {
			l.Line++
			l.Column = 0
		} else {
			l.Column++
		}

		end := false
		if r == '"' {
			value.WriteRune(r)
			end = true
		} else if r == '\\' {
			r2, _, _ := l.R.ReadRune()
			if r2 == '"' {
				l.Column++
				value.WriteRune('"')
			} else {
				value.WriteRune(r)
				l.R.UnreadRune()
			}
		} else {
			value.WriteRune(r)
		}

		if end {
			final, err := strconv.Unquote(value.String())
			if err != nil {
				return 0, "", err
			}
			return tokenWord, final, nil
		}
	}
}

// Parse returns the input from r into a slice of flags that can be
// passed to flag.(*FlagSet).Parse.
func Parse(r io.Reader) ([]string, error) {
	l := lexer{
		R: bufio.NewReader(r),
	}

	var args []string
	newLine := true
	firstOp := true
	addWhitespace := false

	for {
		tkn, val, err := l.Next()
		if err != nil {
			e := &Error{
				Line:   l.Line + 1,
				Column: l.Column + 1,
				Err:    err,
			}
			if file, ok := r.(*os.File); ok {
				e.File = file.Name()
			}
			return nil, e
		}
		switch tkn {
		case tokenEOF:
			return args, nil
		case tokenComment:
			// ignore
		case tokenWhitespace:
			addWhitespace = true
		case tokenNewLine:
			newLine = true
		case tokenWord:
			switch val {
			case "-":
				e := &Error{
					Line:   l.Line + 1,
					Column: l.Column + 1 - len(val),
					Err: &InvalidTokenError{
						Token: val,
					},
				}
				if file, ok := r.(*os.File); ok {
					e.File = file.Name()
				}
				return nil, e
			}

			if newLine {
				args = append(args, "-"+val)
				newLine = false
				firstOp = true
			} else {
				if firstOp {
					args[len(args)-1] += "="
					firstOp = false
				} else if addWhitespace {
					args[len(args)-1] += " "
				}
				args[len(args)-1] += val
			}
			addWhitespace = false
		}
	}
}

// InvalidTokenError is an underlying parse error that is
// returned when an invalid token is encountered in a file.
type InvalidTokenError struct {
	Token string
}

func (e *InvalidTokenError) Error() string {
	return `invalid token "` + e.Token + `"`
}

// Error is a parsing error.
type Error struct {
	// The file in which the error occurred.
	// Empty if no file information was available to the parser.
	File string
	// The position at or near where the error occurred.
	Line   int
	Column int
	// The underlying error, if any.
	Err error
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Error() string {
	var b bytes.Buffer
	b.WriteString(`flagfile: parsing error`)
	if e.File != "" {
		b.WriteString(` in `)
		b.WriteString(e.File)
	}

	b.WriteString(" (line ")
	b.WriteString(strconv.Itoa(e.Line))
	b.WriteString(", column ")
	b.WriteString(strconv.Itoa(e.Column))
	b.WriteString(")")

	if e.Err != nil {
		b.WriteString(": ")
		b.WriteString(e.Err.Error())
	}

	return b.String()
}
