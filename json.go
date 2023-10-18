package main

import (
	"fmt"
	"os"
	"strconv"
)

const (
	LEFT_PAREN_TOKEN  int = 0
	RIGHT_PAREN_TOKEN     = 1
	COLON_TOKEN           = 2
	STRING_TOKEN          = 3
	COMMA_TOKEN           = 4
	NUMBER_TOKEN          = 5
)

type Token struct {
	lexeme string
	token  int
}

type Value struct {
	text    string
	boolean bool
	number  int
}

type Parser struct {
	tokens []Token
}

var position int
var parser = Parser{
	tokens: []Token{},
}

func isDigit(ch byte) bool {
	if '0' <= ch && ch <= '9' {
		return true
	}
	return false
}

// Hashmap that can hold any value with string keys.
type Map map[string]interface{}

func tokenize(source *string) {
	start, end := 0, 0

	for start < len(*source) {
		switch ch := (*source)[end]; ch {
		case '{':
			parser.tokens = append(parser.tokens, Token{(*source)[start : end+1], LEFT_PAREN_TOKEN})
		case '}':
			parser.tokens = append(parser.tokens, Token{(*source)[start : end+1], RIGHT_PAREN_TOKEN})
		case ':':
			parser.tokens = append(parser.tokens, Token{(*source)[start : end+1], COLON_TOKEN})
		case ',':
			parser.tokens = append(parser.tokens, Token{(*source)[start : end+1], COMMA_TOKEN})
		case '"':
			end++
			for (*source)[end] != '"' {
				end++
			}
			// string without quotes
			str := (*source)[start+1 : end]
			parser.tokens = append(parser.tokens, Token{str, STRING_TOKEN})
		// ignore white space
		case ' ':
		case '\n':
		case '\t':
		default:
			if isDigit(ch) {
				for isDigit((*source)[end]) {
					end++
				}
				// Handle floating numbers
				if (*source)[end] == '.' && end+1 < len(*source) && isDigit((*source)[end+1]) {
					end++ // step over '.'
					for isDigit((*source)[end]) {
						end++
					}
				}
				number := (*source)[start:end]
				// Decrement end so that we don't skip over the following character. Kinda ugly.
				end--
				parser.tokens = append(parser.tokens, Token{number, NUMBER_TOKEN})
			} else {
				os.Stderr.WriteString("Found unexpected symbol.\n")
				os.Exit(1)
			}
		}
		start = end + 1
		end++
	}

	for _, token := range parser.tokens {
		fmt.Printf("%d '%s'\n", token.token, token.lexeme)
	}
}

func parseNumber(token Token) (interface{}, error) {
	if i, err := strconv.Atoi(token.lexeme); err == nil {
		return i, nil
	}

	// Try to parse as a float64
	if f, err := strconv.ParseFloat(token.lexeme, 64); err == nil {
		return f, nil
	}

	return nil, fmt.Errorf("'%s' is not an integer or float.\n", token.lexeme)
}

// can be string, bool, int, float,
func parseValue() interface{} {
	switch token := peek(0); token.token {
	case STRING_TOKEN:
		position++
		return token.lexeme
	case NUMBER_TOKEN:
		position++
		num, err := parseNumber(token)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
		return num
	}
	fmt.Fprintf(os.Stderr, "Unexpected token in parseValue: %d\n", peek(0).token)
	os.Exit(1)
	return false
}

func parseJson(m *Map) {
	consume(LEFT_PAREN_TOKEN, "Expected '{' at the start of JSON.")

	// Parsing list of key-value pairs.
	for !check(RIGHT_PAREN_TOKEN) {
		consume(STRING_TOKEN, "Expected a string key.")
		key := parser.tokens[position-1]
		consume(COLON_TOKEN, "Expected a ':' after a key.")

		val := parseValue()

		if str, ok := val.(string); ok {
			fmt.Println(key.lexeme, ": ", str)
		} else if num, ok := val.(int); ok {
			fmt.Println(key, ": ", num)
		}

		(*m)[key.lexeme] = val

		if peek(0).token == COMMA_TOKEN && peek(1).token == RIGHT_PAREN_TOKEN {
			os.Stderr.WriteString("Trailing comma.\n")
			os.Exit(1)
		}

		if peek(0).token != RIGHT_PAREN_TOKEN {
			consume(COMMA_TOKEN, "Expected ',' to separate key-value pairs.")
		}
	}
	consume(RIGHT_PAREN_TOKEN, "Expected '}' at the end of JSON.")
}

func parse(m *Map) {
	for position < len(parser.tokens) {
		parseJson(m)
	}
}

func match(token int) bool {
	if check(token) {
		position++
		return true
	}
	return false
}

func consume(token int, message string) bool {
	if check(token) {
		position++
		return true
	}
	os.Stderr.WriteString(message + "\n")
	return false
}

func check(token int) bool {
	if position >= len(parser.tokens) {
		return false
	}
	return parser.tokens[position].token == token
}

func peek(offset int) Token {
	return parser.tokens[position+offset]
}

func main() {
	if len(os.Args) != 2 {
		os.Stderr.WriteString("Usage: ./json fileName\n")
		os.Exit(2)
	}
	fileName := os.Args[1]

	dat, err := os.ReadFile(fileName)
	if err != nil {
		os.Stderr.WriteString("Failed to read file.")
		os.Exit(1)
	}

	plainJson := string(dat)
	if len(plainJson) < 2 {
		os.Stderr.WriteString("Invalid json file.\n")
		os.Exit(1)
	}

	fmt.Println("Tokenizing...")
	tokenize(&plainJson)
	m := Map{}
	fmt.Println("Parsing...")
	parse(&m)
	fmt.Println(m)
}
