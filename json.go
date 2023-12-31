package main

import (
	"fmt"
	"os"
	"strconv"
)

const (
	LEFT_PAREN_TOKEN    int = 0
	RIGHT_PAREN_TOKEN       = 1
	COLON_TOKEN             = 2
	STRING_TOKEN            = 3
	COMMA_TOKEN             = 4
	NUMBER_TOKEN            = 5
	BOOL_TOKEN              = 6
	NULL_TOKEN              = 7
	LEFT_BRACKET_TOKEN      = 8
	RIGHT_BRACKET_TOKEN     = 9
	MINUS_TOKEN             = 10
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
		case '[':
			parser.tokens = append(parser.tokens, Token{(*source)[start : end+1], LEFT_BRACKET_TOKEN})
		case ']':
			parser.tokens = append(parser.tokens, Token{(*source)[start : end+1], RIGHT_BRACKET_TOKEN})
		case ':':
			parser.tokens = append(parser.tokens, Token{(*source)[start : end+1], COLON_TOKEN})
		case ',':
			parser.tokens = append(parser.tokens, Token{(*source)[start : end+1], COMMA_TOKEN})
		case '"':
			end++
			for (*source)[end] != '"' {
				// if escape \, skip 2 characters instead
				if (*source)[end] == '\\' {
					end++
				}
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

				// Handle scientific notation, 1e123
				expUnsigned := end+1 < len(*source) && isDigit((*source)[end+1])
				expSigned := (end+2 < len(*source) && ((*source)[end+1] == '+' || (*source)[end+1] == '-') && isDigit((*source)[end+2]))
				if ((*source)[end] == 'e' || (*source)[end] == 'E') && (expUnsigned || expSigned) {
					end++ // step over 'e'
					if expSigned {
						end++
					}

					for isDigit((*source)[end]) {
						end++
					}
				}

				number := (*source)[start:end]
				// Decrement end so that we don't skip over the following character. Kinda ugly.
				end--
				parser.tokens = append(parser.tokens, Token{number, NUMBER_TOKEN})
			} else if ch == '-' {
				parser.tokens = append(parser.tokens, Token{"-", MINUS_TOKEN})
			} else if ch == 't' && (*source)[end:end+4] == "true" {
				parser.tokens = append(parser.tokens, Token{"true", BOOL_TOKEN})
				end += 3
			} else if ch == 'f' && (*source)[end:end+5] == "false" {
				parser.tokens = append(parser.tokens, Token{"false", BOOL_TOKEN})
				end += 4
			} else if ch == 'n' && (*source)[end:end+4] == "null" {
				parser.tokens = append(parser.tokens, Token{"null", NULL_TOKEN})
				end += 3
			} else {
				err := fmt.Errorf("Found unexpected symbol '%c'.\n", ch)
				os.Stderr.WriteString(err.Error())
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

func parseNumber(token Token, negative bool) (interface{}, error) {
	if token.lexeme[0] == '0' && len(token.lexeme) > 1 && token.lexeme[1] != '.' {
		return nil, fmt.Errorf("Leading 0 error in '%s'.\n", token.lexeme)
	}
	if i, err := strconv.Atoi(token.lexeme); err == nil {
		if negative {
			return -i, nil
		}
		return i, nil
	}

	// Try to parse as a float64
	if f, err := strconv.ParseFloat(token.lexeme, 64); err == nil {
		if negative {
			return -f, nil
		}
		fmt.Printf("parsing float %f\n", f)
		return f, nil
	}

	return nil, fmt.Errorf("'%s' is not an integer or float.\n", token.lexeme)
}

func parseArray() ([]interface{}, error) {
	arr := []interface{}{}
	consume(LEFT_BRACKET_TOKEN, "Expected '[' at the start of an array.")

	token := peek(0)
	for token.token != RIGHT_BRACKET_TOKEN {
		val := parseValue()
		arr = append(arr, val)
		if peek(0).token != RIGHT_BRACKET_TOKEN && peek(0).token != COMMA_TOKEN {
			return nil, fmt.Errorf("Expected a ',' to separate array elements.\n")
		}

		if peek(0).token == COMMA_TOKEN {
			// skip over comma
			position++
		}
		// Update token
		token = peek(0)
	}
	consume(RIGHT_BRACKET_TOKEN, "Expected ']' to close an array.")
	return arr, nil
}

// can be string, bool, int, float, null
func parseValue() interface{} {
	switch token := peek(0); token.token {
	case STRING_TOKEN:
		position++
		return token.lexeme
	case MINUS_TOKEN:
		consume(MINUS_TOKEN, "Expected minus token.")
		token = peek(0)
		num, err := parseNumber(token, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
		position++
		return num
	case NUMBER_TOKEN:
		num, err := parseNumber(token, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
		position++
		return num
	case BOOL_TOKEN:
		var boolean bool
		switch token.lexeme {
		case "true":
			boolean = true
		case "false":
			boolean = false
		default:
			fmt.Fprintf(os.Stderr, "Unexpected boolean value '%s'\n", peek(0).lexeme)
			os.Exit(1)
		}
		position++
		return boolean
	case NULL_TOKEN:
		position++
		return nil
	case LEFT_BRACKET_TOKEN:
		// Start of an array. Can hold multiple types.
		arr, err := parseArray()
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
		return arr
	case LEFT_PAREN_TOKEN:
		nested_m := Map{}
		parseJson(&nested_m)
		return nested_m
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
			(*m)[key.lexeme] = str
		} else if num, ok := val.(int); ok {
			(*m)[key.lexeme] = num
		} else if num, ok := val.(float64); ok {
			(*m)[key.lexeme] = num
		} else if arr, ok := val.([]interface{}); ok {
			(*m)[key.lexeme] = arr
		} else if boolean, ok := val.(bool); ok {
			(*m)[key.lexeme] = boolean
		} else if val == nil {
			(*m)[key.lexeme] = val
		} else if nestedMap, ok := val.(Map); ok {
			(*m)[key.lexeme] = nestedMap
		}

		if peek(0).token == COMMA_TOKEN && position+1 < len(parser.tokens) && peek(1).token == RIGHT_PAREN_TOKEN {
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
	os.Exit(1)
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

	// Parsed json is stored in 'm'
	parse(&m)
	fmt.Println(m)
}
