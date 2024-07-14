// MIT License
//
// # Copyright (c) 2024 Travis Montoya and the ollie contributors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// Package conf provides a basic application configuration file format.
// The configuration language consists of key-value pairs, one per line,
// separated by the "=" character.
// Lines beginning with "#" are comments and are ignored.
//
// For example:
//
//	logfile = somepath.txt
//	# override default of 420
//	maxcount = 69
package conf

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
	"unicode"
)

type Tokens int

// Default location of the config file
// an example can be found in ../examples
const DefaultConfFile = "/.ollie.conf"
const logFilename = "ollie.log"

var timeFormat = "2006-01-02 15:04:05"

// Holds the key value pairs parsed from the config file
type Settings struct {
	settings map[string]string
}

/*
 * Each Token that we can encounter for our configuration file
 * the layout of the config is simple key value pairs with "="
 * denoting what the value of the key is. We do accept comments
 * in the form of lines beginning with "#"
 */
const (
	TokenError Tokens = iota
	TokenEOF
	TokenNL
	TokenComment
	TokenKey
	TokenValue
	TokenEquals
	TokenString
	TokenInteger
)

// Valid configuration file keys and the valid types for the values
//
// This is used during configuration file validation during the parsing phase
var confParams = map[string]Tokens{
	"spellcheck":     TokenString,
	"dictionary":     TokenString,
	"append-default": TokenString,
}

// Token holds the Token type and the value of the token found in the stream
//
// We will hold a stream of these structs for each token found in the config file
// the tokenizer does NOT validate tokens.
type Token struct {
	Type     Tokens
	Value    string
	Location int // We save the location of the token in the file for error handling
}

type Tokenizer struct {
	lines       *bufio.Scanner
	currentLine string
	location    int
	column      int // Used for position in the line
}

type Parser struct {
	tokenizer *Tokenizer
	tokens    []Token
	location  int
}

func parserTime() string {
	return time.Now().Format(timeFormat)
}

func init() {
	f, err := os.OpenFile(logFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
	}
	log.SetOutput(f)
	log.SetPrefix("confparser: ")
	log.Printf("Logging started at %s\n", parserTime())
}

func NewTokenizer(r io.Reader) *Tokenizer {
	return &Tokenizer{lines: bufio.NewScanner(r)}
}

func NewParser(t *Tokenizer) *Parser {
	return &Parser{tokenizer: t, tokens: []Token{}}
}

// FromFile returns parsed settings from the named file.
func FromFile(name string) (*Settings, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseConfig(f)
}

// ParseConfig parses settings from rd.
func ParseConfig(rd io.Reader) (*Settings, error) {
	tokenizer := NewTokenizer(rd)
	parser := NewParser(tokenizer)
	return parser.Parse()
}

// We iterate the config file lines either returning error tokens or
// passing to tokenizeLine and returning the token found from there.
func (t *Tokenizer) GetNextToken() Token {
	lineNumber := 0
	for t.lines.Scan() {
		log.Printf("%s: parsing line: %s", parserTime(), t.lines.Text())

		// This will be used for error reporting where syntax errors could occur during
		// parsing
		lineNumber += 1
		t.currentLine = t.lines.Text()
		t.location = lineNumber

		if len(strings.TrimSpace(t.currentLine)) == 0 {
			continue
		}

		return t.tokenizeLine(t.currentLine)
	}

	if err := t.lines.Err(); err != nil {
		return Token{Type: TokenError, Value: err.Error(), Location: lineNumber}
	}

	return Token{Type: TokenEOF, Value: ""}
}

// tokenizeLine will parse a line into tokens for the parser
func (t *Tokenizer) tokenizeLine(line string) Token {
	t.column = 0
	for t.column < len(line) {
		sym := line[t.column]

		switch {
		case sym == '#':
			t.column += 1
			return Token{Type: TokenComment, Value: line, Location: t.column}
		case unicode.IsSpace(rune(sym)):
			t.column += 1
		case sym == '=':
			t.column += 1
			return Token{Type: TokenEquals, Value: "=", Location: t.column}
		case sym == '\n':
			t.column += 1
			return Token{Type: TokenNL, Value: "\n", Location: t.column}
		// If it isnt any of the other symbols for the grammar we strip the
		// key/value pairs here.
		case unicode.IsLetter(rune(sym)) || unicode.IsDigit(rune(sym)) || sym == '-':
			log.Println("tokenizer line should be val or key:", line)
			return findKeyOrValueInLine(line, t)
		default:
			t.column += 1
			return Token{Type: TokenError, Value: string(sym), Location: t.column}
		}
	}
	return Token{}
}

func findKeyOrValueInLine(line string, t *Tokenizer) Token {

	fl := strings.Replace(line, " ", "", -1)

	end := 0
	begin := 0
	for end < len(fl) {
		end++
		if fl[end] == '=' {
			break
		}
	}

	// We have reached either something that isnt a character/digit or a valid separating symbol for keys "-"
	// test it for either being =val or key=
	val := fl[begin : end+1]
	if strings.Contains(val, "=") {

		k, v, _ := strings.Cut(val, "=")
		key := strings.TrimSpace(k)
		value := strings.TrimSpace(v)

		if len(value) > 0 {
			return Token{Type: TokenValue, Value: value}
		}

		if len(value) > 0 {
			return Token{Type: TokenKey, Value: key}
		}
	}
	return Token{}
}

func (p *Parser) getNextToken() Token {
	return p.tokenizer.GetNextToken()
}

func (p *Parser) Parse() (*Settings, error) {
	conf := &Settings{settings: make(map[string]string)}

	for {
		token := p.getNextToken()
		if token == (Token{}) {
			continue
		}
		log.Println("current token:", token)
		switch token.Type {
		case TokenComment:
			continue
		case TokenEOF:
			return conf, nil
		case TokenKey:
			key := token.Value
			_, ok := confParams[key]
			if ok {
				nt := p.getNextToken()
				if nt.Type == TokenEquals {
					val := p.getNextToken()
					conf.settings[key] = val.Value
				}
			}
		case TokenValue:
			continue
		case TokenEquals:
		default:
			return nil, fmt.Errorf("Invalid token found %v at %d", token, token.Location)
		}
	}
}
