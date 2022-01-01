// Code generated by re2c, DO NOT EDIT.
// Copyright 2011 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nin

import (
	"fmt"
	"strings"
)

type Token int32

const (
	ERROR Token = iota
	BUILD
	COLON
	DEFAULT
	EQUALS
	IDENT
	INCLUDE
	INDENT
	NEWLINE
	PIPE
	PIPE2
	PIPEAT
	POOL
	RULE
	SUBNINJA
	TEOF
)

type Lexer struct {
	filename_ string
	input_    string
	// In the original C++ code, these two are char pointers and are used to do
	// pointer arithmetics. Go doesn't allow pointer arithmetics so they are
	// indexes. ofs_ starts at 0. last_token_ is initially -1 to mark that it is
	// not yet set.
	ofs_        int
	last_token_ int
}

// Read a path (complete with $escapes).
// Returns false only on error, returned path may be empty if a delimiter
// (space, newline) is hit.
func (l *Lexer) ReadPath(path *EvalString, err *string) bool {
	return l.ReadEvalString(path, true, err)
}

// Read the value side of a var = value line (complete with $escapes).
// Returns false only on error.
func (l *Lexer) ReadVarValue(value *EvalString, err *string) bool {
	return l.ReadEvalString(value, false, err)
}

// Construct an error message with context.
func (l *Lexer) Error(message string, err *string) bool {
	// Compute line/column.
	line := 1
	line_start := 0
	for p := 0; p < l.last_token_; p++ {
		if l.input_[p] == '\n' {
			line++
			line_start = p + 1
		}
	}
	col := 0
	if l.last_token_ != -1 {
		col = l.last_token_ - line_start
	}

	*err = fmt.Sprintf("%s:%d: ", l.filename_, line)
	*err += message + "\n"
	// Add some context to the message.
	const kTruncateColumn = 72
	if col > 0 && col < kTruncateColumn {
		truncated := true
		length := 0
		for ; length < kTruncateColumn; length++ {
			if l.input_[line_start+length] == 0 || l.input_[line_start+length] == '\n' {
				truncated = false
				break
			}
		}
		*err += l.input_[line_start : line_start+length]
		if truncated {
			*err += "..."
		}
		*err += "\n"
		*err += strings.Repeat(" ", col)
		*err += "^ near here"
	}
	return false
}

// NewLexer is only used in tests.
func NewLexer(input string) Lexer {
	l := Lexer{}
	l.Start("input", input+"\x00")
	return l
}

// Start parsing some input.
func (l *Lexer) Start(filename, input string) {
	l.filename_ = filename
	if !strings.HasSuffix(input, "\x00") {
		panic("Requires hack with a trailing 0 byte")
	}
	l.input_ = input
	l.ofs_ = 0
	l.last_token_ = -1
}

// Return a human-readable form of a token, used in error messages.
func TokenName(t Token) string {
	switch t {
	case ERROR:
		return "lexing error"
	case BUILD:
		return "'build'"
	case COLON:
		return "':'"
	case DEFAULT:
		return "'default'"
	case EQUALS:
		return "'='"
	case IDENT:
		return "identifier"
	case INCLUDE:
		return "'include'"
	case INDENT:
		return "indent"
	case NEWLINE:
		return "newline"
	case PIPE2:
		return "'||'"
	case PIPE:
		return "'|'"
	case PIPEAT:
		return "'|@'"
	case POOL:
		return "'pool'"
	case RULE:
		return "'rule'"
	case SUBNINJA:
		return "'subninja'"
	case TEOF:
		return "eof"
	}
	return "" // not reached
}

// Return a human-readable token hint, used in error messages.
func TokenErrorHint(expected Token) string {
	switch expected {
	case COLON:
		return " ($ also escapes ':')"
	default:
		return ""
	}
}

// If the last token read was an ERROR token, provide more info
// or the empty string.
func (l *Lexer) DescribeLastError() string {
	if l.last_token_ != -1 {
		switch l.input_[l.last_token_] {
		case '\t':
			return "tabs are not allowed, use spaces"
		}
	}
	return "lexing error"
}

// Rewind to the last read Token.
func (l *Lexer) UnreadToken() {
	l.ofs_ = l.last_token_
}

func (l *Lexer) ReadToken() Token {
	p := l.ofs_
	q := 0
	start := 0
	var token Token
	for {
		start = p

		{
			var yych byte
			yyaccept := 0
			yych = l.input_[p]
			if yych <= 'Z' {
				if yych <= '#' {
					if yych <= '\f' {
						if yych <= 0x00 {
							goto yy2
						}
						if yych == '\n' {
							goto yy6
						}
						goto yy4
					} else {
						if yych <= 0x1F {
							if yych <= '\r' {
								goto yy8
							}
							goto yy4
						} else {
							if yych <= ' ' {
								goto yy9
							}
							if yych <= '"' {
								goto yy4
							}
							goto yy12
						}
					}
				} else {
					if yych <= '9' {
						if yych <= ',' {
							goto yy4
						}
						if yych == '/' {
							goto yy4
						}
						goto yy13
					} else {
						if yych <= '<' {
							if yych <= ':' {
								goto yy16
							}
							goto yy4
						} else {
							if yych <= '=' {
								goto yy18
							}
							if yych <= '@' {
								goto yy4
							}
							goto yy13
						}
					}
				}
			} else {
				if yych <= 'i' {
					if yych <= 'a' {
						if yych == '_' {
							goto yy13
						}
						if yych <= '`' {
							goto yy4
						}
						goto yy13
					} else {
						if yych <= 'c' {
							if yych <= 'b' {
								goto yy20
							}
							goto yy13
						} else {
							if yych <= 'd' {
								goto yy21
							}
							if yych <= 'h' {
								goto yy13
							}
							goto yy22
						}
					}
				} else {
					if yych <= 'r' {
						if yych == 'p' {
							goto yy23
						}
						if yych <= 'q' {
							goto yy13
						}
						goto yy24
					} else {
						if yych <= 'z' {
							if yych <= 's' {
								goto yy25
							}
							goto yy13
						} else {
							if yych == '|' {
								goto yy26
							}
							goto yy4
						}
					}
				}
			}
		yy2:
			p++
			{
				token = TEOF
				break
			}
		yy4:
			p++
		yy5:
			{
				token = ERROR
				break
			}
		yy6:
			p++
			{
				token = NEWLINE
				break
			}
		yy8:
			p++
			yych = l.input_[p]
			if yych == '\n' {
				goto yy28
			}
			goto yy5
		yy9:
			yyaccept = 0
			p++
			q = p
			yych = l.input_[p]
			if yych <= '\r' {
				if yych == '\n' {
					goto yy6
				}
				if yych >= '\r' {
					goto yy30
				}
			} else {
				if yych <= ' ' {
					if yych >= ' ' {
						goto yy9
					}
				} else {
					if yych == '#' {
						goto yy32
					}
				}
			}
		yy11:
			{
				token = INDENT
				break
			}
		yy12:
			yyaccept = 1
			p++
			q = p
			yych = l.input_[p]
			if yych <= 0x00 {
				goto yy5
			}
			goto yy33
		yy13:
			p++
			yych = l.input_[p]
		yy14:
			if yych <= '@' {
				if yych <= '.' {
					if yych >= '-' {
						goto yy13
					}
				} else {
					if yych <= '/' {
						goto yy15
					}
					if yych <= '9' {
						goto yy13
					}
				}
			} else {
				if yych <= '_' {
					if yych <= 'Z' {
						goto yy13
					}
					if yych >= '_' {
						goto yy13
					}
				} else {
					if yych <= '`' {
						goto yy15
					}
					if yych <= 'z' {
						goto yy13
					}
				}
			}
		yy15:
			{
				token = IDENT
				break
			}
		yy16:
			p++
			{
				token = COLON
				break
			}
		yy18:
			p++
			{
				token = EQUALS
				break
			}
		yy20:
			p++
			yych = l.input_[p]
			if yych == 'u' {
				goto yy36
			}
			goto yy14
		yy21:
			p++
			yych = l.input_[p]
			if yych == 'e' {
				goto yy37
			}
			goto yy14
		yy22:
			p++
			yych = l.input_[p]
			if yych == 'n' {
				goto yy38
			}
			goto yy14
		yy23:
			p++
			yych = l.input_[p]
			if yych == 'o' {
				goto yy39
			}
			goto yy14
		yy24:
			p++
			yych = l.input_[p]
			if yych == 'u' {
				goto yy40
			}
			goto yy14
		yy25:
			p++
			yych = l.input_[p]
			if yych == 'u' {
				goto yy41
			}
			goto yy14
		yy26:
			p++
			yych = l.input_[p]
			if yych == '@' {
				goto yy42
			}
			if yych == '|' {
				goto yy44
			}
			{
				token = PIPE
				break
			}
		yy28:
			p++
			{
				token = NEWLINE
				break
			}
		yy30:
			p++
			yych = l.input_[p]
			if yych == '\n' {
				goto yy28
			}
		yy31:
			p = q
			if yyaccept == 0 {
				goto yy11
			} else {
				goto yy5
			}
		yy32:
			p++
			yych = l.input_[p]
		yy33:
			if yych <= 0x00 {
				goto yy31
			}
			if yych != '\n' {
				goto yy32
			}
			p++
			{
				continue
			}
		yy36:
			p++
			yych = l.input_[p]
			if yych == 'i' {
				goto yy46
			}
			goto yy14
		yy37:
			p++
			yych = l.input_[p]
			if yych == 'f' {
				goto yy47
			}
			goto yy14
		yy38:
			p++
			yych = l.input_[p]
			if yych == 'c' {
				goto yy48
			}
			goto yy14
		yy39:
			p++
			yych = l.input_[p]
			if yych == 'o' {
				goto yy49
			}
			goto yy14
		yy40:
			p++
			yych = l.input_[p]
			if yych == 'l' {
				goto yy50
			}
			goto yy14
		yy41:
			p++
			yych = l.input_[p]
			if yych == 'b' {
				goto yy51
			}
			goto yy14
		yy42:
			p++
			{
				token = PIPEAT
				break
			}
		yy44:
			p++
			{
				token = PIPE2
				break
			}
		yy46:
			p++
			yych = l.input_[p]
			if yych == 'l' {
				goto yy52
			}
			goto yy14
		yy47:
			p++
			yych = l.input_[p]
			if yych == 'a' {
				goto yy53
			}
			goto yy14
		yy48:
			p++
			yych = l.input_[p]
			if yych == 'l' {
				goto yy54
			}
			goto yy14
		yy49:
			p++
			yych = l.input_[p]
			if yych == 'l' {
				goto yy55
			}
			goto yy14
		yy50:
			p++
			yych = l.input_[p]
			if yych == 'e' {
				goto yy57
			}
			goto yy14
		yy51:
			p++
			yych = l.input_[p]
			if yych == 'n' {
				goto yy59
			}
			goto yy14
		yy52:
			p++
			yych = l.input_[p]
			if yych == 'd' {
				goto yy60
			}
			goto yy14
		yy53:
			p++
			yych = l.input_[p]
			if yych == 'u' {
				goto yy62
			}
			goto yy14
		yy54:
			p++
			yych = l.input_[p]
			if yych == 'u' {
				goto yy63
			}
			goto yy14
		yy55:
			p++
			yych = l.input_[p]
			if yych <= '@' {
				if yych <= '.' {
					if yych >= '-' {
						goto yy13
					}
				} else {
					if yych <= '/' {
						goto yy56
					}
					if yych <= '9' {
						goto yy13
					}
				}
			} else {
				if yych <= '_' {
					if yych <= 'Z' {
						goto yy13
					}
					if yych >= '_' {
						goto yy13
					}
				} else {
					if yych <= '`' {
						goto yy56
					}
					if yych <= 'z' {
						goto yy13
					}
				}
			}
		yy56:
			{
				token = POOL
				break
			}
		yy57:
			p++
			yych = l.input_[p]
			if yych <= '@' {
				if yych <= '.' {
					if yych >= '-' {
						goto yy13
					}
				} else {
					if yych <= '/' {
						goto yy58
					}
					if yych <= '9' {
						goto yy13
					}
				}
			} else {
				if yych <= '_' {
					if yych <= 'Z' {
						goto yy13
					}
					if yych >= '_' {
						goto yy13
					}
				} else {
					if yych <= '`' {
						goto yy58
					}
					if yych <= 'z' {
						goto yy13
					}
				}
			}
		yy58:
			{
				token = RULE
				break
			}
		yy59:
			p++
			yych = l.input_[p]
			if yych == 'i' {
				goto yy64
			}
			goto yy14
		yy60:
			p++
			yych = l.input_[p]
			if yych <= '@' {
				if yych <= '.' {
					if yych >= '-' {
						goto yy13
					}
				} else {
					if yych <= '/' {
						goto yy61
					}
					if yych <= '9' {
						goto yy13
					}
				}
			} else {
				if yych <= '_' {
					if yych <= 'Z' {
						goto yy13
					}
					if yych >= '_' {
						goto yy13
					}
				} else {
					if yych <= '`' {
						goto yy61
					}
					if yych <= 'z' {
						goto yy13
					}
				}
			}
		yy61:
			{
				token = BUILD
				break
			}
		yy62:
			p++
			yych = l.input_[p]
			if yych == 'l' {
				goto yy65
			}
			goto yy14
		yy63:
			p++
			yych = l.input_[p]
			if yych == 'd' {
				goto yy66
			}
			goto yy14
		yy64:
			p++
			yych = l.input_[p]
			if yych == 'n' {
				goto yy67
			}
			goto yy14
		yy65:
			p++
			yych = l.input_[p]
			if yych == 't' {
				goto yy68
			}
			goto yy14
		yy66:
			p++
			yych = l.input_[p]
			if yych == 'e' {
				goto yy70
			}
			goto yy14
		yy67:
			p++
			yych = l.input_[p]
			if yych == 'j' {
				goto yy72
			}
			goto yy14
		yy68:
			p++
			yych = l.input_[p]
			if yych <= '@' {
				if yych <= '.' {
					if yych >= '-' {
						goto yy13
					}
				} else {
					if yych <= '/' {
						goto yy69
					}
					if yych <= '9' {
						goto yy13
					}
				}
			} else {
				if yych <= '_' {
					if yych <= 'Z' {
						goto yy13
					}
					if yych >= '_' {
						goto yy13
					}
				} else {
					if yych <= '`' {
						goto yy69
					}
					if yych <= 'z' {
						goto yy13
					}
				}
			}
		yy69:
			{
				token = DEFAULT
				break
			}
		yy70:
			p++
			yych = l.input_[p]
			if yych <= '@' {
				if yych <= '.' {
					if yych >= '-' {
						goto yy13
					}
				} else {
					if yych <= '/' {
						goto yy71
					}
					if yych <= '9' {
						goto yy13
					}
				}
			} else {
				if yych <= '_' {
					if yych <= 'Z' {
						goto yy13
					}
					if yych >= '_' {
						goto yy13
					}
				} else {
					if yych <= '`' {
						goto yy71
					}
					if yych <= 'z' {
						goto yy13
					}
				}
			}
		yy71:
			{
				token = INCLUDE
				break
			}
		yy72:
			p++
			yych = l.input_[p]
			if yych != 'a' {
				goto yy14
			}
			p++
			yych = l.input_[p]
			if yych <= '@' {
				if yych <= '.' {
					if yych >= '-' {
						goto yy13
					}
				} else {
					if yych <= '/' {
						goto yy74
					}
					if yych <= '9' {
						goto yy13
					}
				}
			} else {
				if yych <= '_' {
					if yych <= 'Z' {
						goto yy13
					}
					if yych >= '_' {
						goto yy13
					}
				} else {
					if yych <= '`' {
						goto yy74
					}
					if yych <= 'z' {
						goto yy13
					}
				}
			}
		yy74:
			{
				token = SUBNINJA
				break
			}
		}

	}

	l.last_token_ = start
	l.ofs_ = p
	if token != NEWLINE && token != TEOF {
		l.EatWhitespace()
	}
	return token
}

// If the next token is \a token, read it and return true.
func (l *Lexer) PeekToken(token Token) bool {
	t := l.ReadToken()
	if t == token {
		return true
	}
	l.UnreadToken()
	return false
}

// Skip past whitespace (called after each read token/ident/etc.).
func (l *Lexer) EatWhitespace() {
	p := l.ofs_
	q := 0
	for {
		l.ofs_ = p

		{
			var yych byte
			yych = l.input_[p]
			if yych <= ' ' {
				if yych <= 0x00 {
					goto yy77
				}
				if yych <= 0x1F {
					goto yy79
				}
				goto yy81
			} else {
				if yych == '$' {
					goto yy84
				}
				goto yy79
			}
		yy77:
			p++
			{
				break
			}
		yy79:
			p++
		yy80:
			{
				break
			}
		yy81:
			p++
			yych = l.input_[p]
			if yych == ' ' {
				goto yy81
			}
			{
				continue
			}
		yy84:
			p++
			q = p
			yych = l.input_[p]
			if yych == '\n' {
				goto yy85
			}
			if yych == '\r' {
				goto yy87
			}
			goto yy80
		yy85:
			p++
			{
				continue
			}
		yy87:
			p++
			yych = l.input_[p]
			if yych == '\n' {
				goto yy89
			}
			p = q
			goto yy80
		yy89:
			p++
			{
				continue
			}
		}

	}
}

// Read a simple identifier (a rule or variable name).
// Returns false if a name can't be read.
func (l *Lexer) ReadIdent(out *string) bool {
	p := l.ofs_
	start := 0
	for {
		start = p

		{
			var yych byte
			yych = l.input_[p]
			if yych <= '@' {
				if yych <= '.' {
					if yych >= '-' {
						goto yy95
					}
				} else {
					if yych <= '/' {
						goto yy93
					}
					if yych <= '9' {
						goto yy95
					}
				}
			} else {
				if yych <= '_' {
					if yych <= 'Z' {
						goto yy95
					}
					if yych >= '_' {
						goto yy95
					}
				} else {
					if yych <= '`' {
						goto yy93
					}
					if yych <= 'z' {
						goto yy95
					}
				}
			}
		yy93:
			p++
			{
				l.last_token_ = start
				return false
			}
		yy95:
			p++
			yych = l.input_[p]
			if yych <= '@' {
				if yych <= '.' {
					if yych >= '-' {
						goto yy95
					}
				} else {
					if yych <= '/' {
						goto yy97
					}
					if yych <= '9' {
						goto yy95
					}
				}
			} else {
				if yych <= '_' {
					if yych <= 'Z' {
						goto yy95
					}
					if yych >= '_' {
						goto yy95
					}
				} else {
					if yych <= '`' {
						goto yy97
					}
					if yych <= 'z' {
						goto yy95
					}
				}
			}
		yy97:
			{
				*out = l.input_[start:p]
				break
			}
		}

	}
	l.last_token_ = start
	l.ofs_ = p
	l.EatWhitespace()
	return true
}

// Read a $-escaped string.
func (l *Lexer) ReadEvalString(eval *EvalString, path bool, err *string) bool {
	p := l.ofs_
	q := 0
	start := 0
	for {
		start = p

		{
			var yych byte
			yych = l.input_[p]
			if yych <= ' ' {
				if yych <= '\n' {
					if yych <= 0x00 {
						goto yy100
					}
					if yych <= '\t' {
						goto yy102
					}
					goto yy105
				} else {
					if yych == '\r' {
						goto yy107
					}
					if yych <= 0x1F {
						goto yy102
					}
					goto yy105
				}
			} else {
				if yych <= '9' {
					if yych == '$' {
						goto yy109
					}
					goto yy102
				} else {
					if yych <= ':' {
						goto yy105
					}
					if yych == '|' {
						goto yy105
					}
					goto yy102
				}
			}
		yy100:
			p++
			{
				l.last_token_ = start
				return l.Error("unexpected EOF", err)
			}
		yy102:
			p++
			yych = l.input_[p]
			if yych <= ' ' {
				if yych <= '\n' {
					if yych <= 0x00 {
						goto yy104
					}
					if yych <= '\t' {
						goto yy102
					}
				} else {
					if yych == '\r' {
						goto yy104
					}
					if yych <= 0x1F {
						goto yy102
					}
				}
			} else {
				if yych <= '9' {
					if yych != '$' {
						goto yy102
					}
				} else {
					if yych <= ':' {
						goto yy104
					}
					if yych != '|' {
						goto yy102
					}
				}
			}
		yy104:
			{
				eval.AddText(l.input_[start:p])
				continue
			}
		yy105:
			p++
			{
				if path {
					p = start
					break
				} else {
					if l.input_[start] == '\n' {
						break
					}
					eval.AddText(l.input_[start : start+1])
					continue
				}
			}
		yy107:
			p++
			yych = l.input_[p]
			if yych == '\n' {
				goto yy110
			}
			{
				l.last_token_ = start
				return l.Error(l.DescribeLastError(), err)
			}
		yy109:
			p++
			yych = l.input_[p]
			if yych <= '-' {
				if yych <= 0x1F {
					if yych <= '\n' {
						if yych <= '\t' {
							goto yy112
						}
						goto yy114
					} else {
						if yych == '\r' {
							goto yy117
						}
						goto yy112
					}
				} else {
					if yych <= '#' {
						if yych <= ' ' {
							goto yy118
						}
						goto yy112
					} else {
						if yych <= '$' {
							goto yy120
						}
						if yych <= ',' {
							goto yy112
						}
						goto yy122
					}
				}
			} else {
				if yych <= 'Z' {
					if yych <= '9' {
						if yych <= '/' {
							goto yy112
						}
						goto yy122
					} else {
						if yych <= ':' {
							goto yy125
						}
						if yych <= '@' {
							goto yy112
						}
						goto yy122
					}
				} else {
					if yych <= '`' {
						if yych == '_' {
							goto yy122
						}
						goto yy112
					} else {
						if yych <= 'z' {
							goto yy122
						}
						if yych <= '{' {
							goto yy127
						}
						goto yy112
					}
				}
			}
		yy110:
			p++
			{
				if path {
					p = start
				}
				break
			}
		yy112:
			p++
		yy113:
			{
				l.last_token_ = start
				return l.Error("bad $-escape (literal $ must be written as $$)", err)
			}
		yy114:
			p++
			yych = l.input_[p]
			if yych == ' ' {
				goto yy114
			}
			{
				continue
			}
		yy117:
			p++
			yych = l.input_[p]
			if yych == '\n' {
				goto yy128
			}
			goto yy113
		yy118:
			p++
			{
				eval.AddText(" ")
				continue
			}
		yy120:
			p++
			{
				eval.AddText("$")
				continue
			}
		yy122:
			p++
			yych = l.input_[p]
			if yych <= '@' {
				if yych <= '-' {
					if yych >= '-' {
						goto yy122
					}
				} else {
					if yych <= '/' {
						goto yy124
					}
					if yych <= '9' {
						goto yy122
					}
				}
			} else {
				if yych <= '_' {
					if yych <= 'Z' {
						goto yy122
					}
					if yych >= '_' {
						goto yy122
					}
				} else {
					if yych <= '`' {
						goto yy124
					}
					if yych <= 'z' {
						goto yy122
					}
				}
			}
		yy124:
			{
				eval.AddSpecial(l.input_[start+1 : p])
				continue
			}
		yy125:
			p++
			{
				eval.AddText(":")
				continue
			}
		yy127:
			p++
			q = p
			yych = l.input_[p]
			if yych <= '@' {
				if yych <= '.' {
					if yych <= ',' {
						goto yy113
					}
					goto yy131
				} else {
					if yych <= '/' {
						goto yy113
					}
					if yych <= '9' {
						goto yy131
					}
					goto yy113
				}
			} else {
				if yych <= '_' {
					if yych <= 'Z' {
						goto yy131
					}
					if yych <= '^' {
						goto yy113
					}
					goto yy131
				} else {
					if yych <= '`' {
						goto yy113
					}
					if yych <= 'z' {
						goto yy131
					}
					goto yy113
				}
			}
		yy128:
			p++
			yych = l.input_[p]
			if yych == ' ' {
				goto yy128
			}
			{
				continue
			}
		yy131:
			p++
			yych = l.input_[p]
			if yych <= 'Z' {
				if yych <= '/' {
					if yych <= ',' {
						goto yy133
					}
					if yych <= '.' {
						goto yy131
					}
				} else {
					if yych <= '9' {
						goto yy131
					}
					if yych >= 'A' {
						goto yy131
					}
				}
			} else {
				if yych <= '`' {
					if yych == '_' {
						goto yy131
					}
				} else {
					if yych <= 'z' {
						goto yy131
					}
					if yych == '}' {
						goto yy134
					}
				}
			}
		yy133:
			p = q
			goto yy113
		yy134:
			p++
			{
				eval.AddSpecial(l.input_[start+2 : p-1])
				continue
			}
		}

	}
	l.last_token_ = start
	l.ofs_ = p
	if path {
		l.EatWhitespace()
	}
	// Non-path strings end in newlines, so there's no whitespace to eat.
	return true
}
