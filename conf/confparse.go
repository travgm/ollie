package ollie

import (
	"fmt"
	"bufio"
	"os"
	"strings"
	"unicode"
)

type Tokens int

const defaultConfFile := "~/.ollie.conf"

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
confParams := map[string]int{
	"spellcheck": TokenString,
	"dictionary": TokenString,
	"append-default": TokenString,
}

// Token holds the Token type and the value of the token found in the stream
type Token struct {
	Type Token
	Value string
	Location int  // We save the location of the token in the file for error handling
}

type Tokenizer struct {
	lines *bufio.Scanner
	currentLine string
	location int
	column int
}

type Parser struct {
	tokenizer *Tokenizer
	tokens []Token
	location int
}


func NewTokenizer(i *os.File) *Tokenizer {
	return &Tokenizer{lines: bufio.NewScanner(i)}
}

func NewParser(t *Tokenizer) *Parser {
	return &Parser{tokenizer: t, tokens []Token{}}
}

func ParseConfig() (*Settings, error) {
	conf, err := os.Open(defaultConfFile)
	if err != nil {
		return nil, fmt.Errorf("no config file found")
	}
	defer conf.Close()

	// Need to add error checking
	tokenizer := NewTokenizer(conf)
	parser := NewParser(tokenizer)

	parsedConfig, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	return parsedConfig, nil
}

// We iterate the config file lines either returning error tokens or
// passing to tokenizeLine and returning the token found from there.
func (t *Tokenizer) GetNextToken() Token {
	lineNumber := 0
	for t.lines.Scan() {
		// This will be used for error reporting where syntax errors could occur during
		// parsing
		lineNumber += 1
		t.currentLine = t.lines.Text()
		t.location = lineNumber

		if (len(strings.TrimSpace(t.currentLine)) == 0) {
			continue;
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

		switch sym {
			case '#':
				t.column += 1
				return Token{Type: TokenComment, Value: line, Location: t.column}
			case unicode.IsSpace(symbol):
				t.column += 1
			case '=':
				t.column += 1
				return Token{Type: TokenEquals, Value: sym, Location: t.column}
			case '\n':
				t.column += 1
				return Token{Type: TokenNL, Value: sym, Location: t.column}
			// If it isnt any of the other symbols for the grammar we strip the
			// key/value pairs here.
			case unicode.IsLetter(sym) || unicode.IsDigit(sym) || sym == '-':
				return findKeyOrValueInLine(line, t)
			default:
				t.column += 1
				return Token{Type: TokenError, Value: sym, Location: t.column}
		}
	}
}

func findKey(string line, t *Tokenizer) Token {
	begin := t.column
	for t.column < len(line) && 
		(unicode.IsLetter(line[t.column]) || unicode.IsDigit(line[t.column]) || line[t.column] == '-') {
		t.column++
	}

	// We have reached either something that isnt a character/digit or a valid separating symbol for keys "-"
	// test it for either being =val or key=
	val := line[begin:t.column]
	if strings.Contains(val, "=") {
		valOrKey := strings.Split(value, "=")
		if len(valOrKey) >= 2 {
			if valOrKey[0] == "=" {
				return Token{Type: TokenValue, Value: val}
			} else if valOrKey[1] == "=" {
				return Token{Type: TokenKey, Value: valOrKey[0]}
			}
		}
	}
}

func (p *Parser) getNextToken() Token {
	if p.location >= len(p.tokens) {
		token := p.tokenizer.GetNextToken()
		p.tokens = append(p.tokens, p.tokenizer.GetNextToken())
	}
	p.location += 1
	return p.tokens[p.location-1]
}

func (p *Parser) Parse() (*Settings, error) {
	settings := &Settings{settings: make(map[string]string)}

	for {
		token := p.getNextToken()
		switch token.Type {
			case TokenComment:
				continue
			case TokenEOF:
				return settings, nil
			case TokenKey:
				key := token.Value
				_, ok := confParams[key]
				if ok {
					nt := p.getNextToken()
					if nt.Type == TokenEquals {
						val := p.getNextToken()
						settings[key] = val
					}
				}
			default:
				return nil, fmt.Errorf("Invalid token found %v at %d", token, token.Location)
		}
	}
}
