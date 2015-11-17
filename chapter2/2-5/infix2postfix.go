package main
import (
	"fmt"
	"unicode"
	"errors"
	"log"
	"reflect"
	"bytes"
	"io"
)

/****************************************************parser********************************************************/
type Parser struct {
	lookahead interface{}
	lexer *Lexer
}

func NewParser() *Parser {
	parser := &Parser{lexer:NewLexer()}
	parser.lookahead = parser.lexer.Scan()
	if parser.lookahead == nil {
		log.Fatalln("NewParser(): no valid input, parser.lookahead == nil")
	}
	return parser
}

func (parser *Parser) Expr() {
	parser.term()
	parser.rest()
	fmt.Print("\n")
}

func (parser *Parser) term() {
	if parser.lookahead == nil {
		log.Fatalln("term(): parser.lookahead == nil")
	}
	v := reflect.ValueOf(parser.lookahead)
	if  v.Field(0).Interface() != NUM {
		log.Fatalln(errors.New("term(): syntax error:"), "lookahead is", parser.lookahead)
	}
	parser.match(parser.lookahead)
	fmt.Print(v.Field(1), " ")
}

func (parser *Parser) rest() {
	if parser.lookahead == nil {
		return
	}
	tok := reflect.ValueOf(parser.lookahead)
	tag := tok.Field(0).Interface().(Tag)
	if tag == '+' {
		parser.match(NewToken('+'))
		parser.term()
		fmt.Printf("%c ", '+')
		parser.rest()
	} else if tag == '-' {
		parser.match(NewToken('-'))
		parser.term()
		fmt.Printf("%c ", '-')
		parser.rest()
	} else {
		// do nothing
	}
}

func (parser *Parser) match(c interface{}) {
	if parser.lookahead == c {
		parser.lookahead = parser.lexer.Scan()
		if parser.lookahead == nil {
			return
		}
	} else {
		log.Fatalln("match(): syntax error,", "lookahead is", parser.lookahead, "and c is", c)
	}
}

/**************************************************lexer***************************************************/
type Tag int

const (
	NUM Tag = 256
	ID Tag = 257
	TRUE Tag = 258
	FALSE Tag = 259
)

type Token struct {
	TAG Tag
}

func NewToken(tag Tag) Token {
	return Token{tag}
}

type Num struct {
	TAG Tag
	Value int
}

func NewNum(tag Tag, value int) Num {
	return Num{tag, value}
}

type Word struct {
	TAG Tag
	Lexeme string
}

func NewWord(tag Tag, lexeme string) Word {
	return Word{tag, lexeme}
}

type Lexer struct {
	Words map[string]interface{}
	Line int
	peek byte
}

func NewLexer() *Lexer {
	return &Lexer{
		Words:map[string]interface{}{
			"true": NewWord(TRUE, "true"),
			"false": NewWord(FALSE, "false"),
		},
		Line:0,
		peek:' ',
	}
}

func (lexer *Lexer) Scan() interface{} {
	for {
		_, err := fmt.Scanf("%c", &lexer.peek)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalln("Scan():", err)
		}

		// omit the blank symbol
		if lexer.peek == ' ' || lexer.peek == '\t' {
			continue
		} else if lexer.peek == '\n' {
			lexer.Line++
			continue
		}

		// process digits
		if unicode.IsDigit(rune(lexer.peek)) {
			v := 0
			for unicode.IsDigit(rune(lexer.peek)) {
				v = v * 10 + int(lexer.peek - '0')
				_, err := fmt.Scanf("%c", &lexer.peek)
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Fatalln("Scan() process digits:", err)
				}
			}
			return NewNum(NUM, v)
		}

		// process identifier
		var w bytes.Buffer
		if unicode.IsLetter(rune(lexer.peek)) {
			w.WriteByte(lexer.peek)
			for {
				_, err := fmt.Scanf("%c", &lexer.peek)
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Fatalln("Scan() process identifier: ", err)
				}
				if unicode.IsDigit(rune(lexer.peek)) || unicode.IsLetter(rune(lexer.peek)) {
					w.WriteByte(lexer.peek)
				} else {
					break
				}
			}
			if word, ok := lexer.Words[w.String()]; ok {
				return word
			}
			word := NewWord(ID, w.String())
			lexer.Words[w.String()] = word
			return word
		}

		// process other symbols
		tok := NewToken(Tag(lexer.peek))
		lexer.peek = ' '
		return tok
	}
}

func main() {
	fmt.Println("please input the infix expression:")
	parser := NewParser()
	parser.Expr()
}