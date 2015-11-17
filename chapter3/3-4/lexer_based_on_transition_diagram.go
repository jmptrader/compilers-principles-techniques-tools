package main
import (
	"io"
	"log"
	"unicode"
	"os"
	"fmt"
	"reflect"
)


type Lexer struct {
	words map[string]Token
	df    *DoubleBuffer
}

func newLexer(bufSize int, inputSrc io.Reader) *Lexer {
	lexer := &Lexer{words:make(map[string]Token)}
	lexer.words["if"] = newId(IF, "if")
	lexer.words["then"] = newId(THEN, "then")
	lexer.words["else"] = newId(ELSE, "else")
	lexer.words["while"] = newId(WHILE, "while")
	lexer.words["do"] = newId(DO, "do")
	lexer.words["for"] = newId(FOR, "for")

	lexer.df = newDoubleBuffer(bufSize, inputSrc)
	if lexer.df == nil {
		log.Println("newLexer(): newDoubleBuffer() returned nil")
		return nil
	}
	return lexer
}

func (lexer *Lexer) nextToken() Token {
	ch, err := lexer.df.nextChar()
	if err != nil {
		return nil
	}
	switch {
	case unicode.IsDigit(rune(ch)):
		return lexer.nextNumber(ch)
	case unicode.IsLetter(rune(ch)):
		return lexer.nextId(ch)
	case ch == '<' || ch == '=' || ch == '>':
		return lexer.nextRelop(ch)
	case ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r':
		return lexer.nextWs(ch)
	default:
		log.Fatalln("Lexer::nextToken(): invalid character:", ch)
	}
	return nil
}

func (lexer *Lexer) nextId(ch byte) *Id {
	state := 9
	for {
		switch state {
		case 9:
			if unicode.IsLetter(rune(ch)) {
				state = 10
				ch, _ = lexer.df.nextChar()
			} else {
				log.Fatalln("Lexer::nextId(): invalid input", ch, "and current state is", state)
			}
		case 10:
			if unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) {
				state = 10
				ch, _ = lexer.df.nextChar()
			} else {
				state = 11
			}
		case 11:
			lexeme := lexer.df.nextLexeme()
			if id, ok := lexer.words[lexeme]; ok {
				return id.(*Id)
			}
			id := newId(REST, lexeme)
			lexer.words[lexeme] = id
			return id
		}
	}
}

func (lexer *Lexer) nextNumber(ch byte) *Number {
	state := 12
	for {
		switch state {
		case 12:
			if unicode.IsDigit(rune(ch)) {
				state = 13
				ch, _ = lexer.df.nextChar()
			} else {
				log.Fatalln("Lexer::nextNumber(): invalid input", ch, "and current state is", state)
			}
		case 13:
			if unicode.IsDigit(rune(ch)) {
				state = 13
				ch, _ = lexer.df.nextChar()
			} else if ch == '.' {
				state = 14
				ch, _ = lexer.df.nextChar()
			} else if ch == 'E' {
				state = 16
				ch, _ = lexer.df.nextChar()
			} else {
				state = 20
			}
		case 14:
			if unicode.IsDigit(rune(ch)) {
				state = 15
				ch, _ = lexer.df.nextChar()
			} else {
				log.Fatalln("Lexer::nextNumber(): invalid input", ch, "and current state is", state)
			}
		case 15:
			if unicode.IsDigit(rune(ch)) {
				state = 15
				ch, _ = lexer.df.nextChar()
			} else if ch == 'E' {
				state = 16
				ch, _ = lexer.df.nextChar()
			} else {
				state = 21
			}
		case 16:
			if ch == '+' || ch == '-' {
				state = 17
				ch, _ = lexer.df.nextChar()
			} else if unicode.IsDigit(rune(ch)) {
				state = 18
				ch, _ = lexer.df.nextChar()
			} else {
				log.Fatalln("Lexer::nextNumber(): invalid input", ch, "and current state is", state)
			}
		case 17:
			if unicode.IsDigit(rune(ch)) {
				state = 18
				ch, _ = lexer.df.nextChar()
			} else {
				log.Fatalln("Lexer::nextNumber(): invalid input", ch, "and current state is", state)
			}
		case 18:
			if unicode.IsDigit(rune(ch)) {
				state = 18
				ch, _ = lexer.df.nextChar()
			} else {
				state = 19
			}
		case 19, 20, 21:
			return newNumber(lexer.df.nextLexeme())
		}
	}
}

func (lexer *Lexer) nextRelop(ch byte) *Relop {
	state := 0
	for {
		switch state {
		case 0:
			if ch == '<' {
				state = 1
				ch, _ = lexer.df.nextChar()
			} else if ch == '=' {
				state = 5
			} else if ch == '>' {
				state = 6
				ch, _ = lexer.df.nextChar()
			} else {
				log.Fatalln("Lexer::nextRelop(): invalid input", ch, "and current state is", state)
			}
		case 1:
			if ch == '=' {
				state = 2
			} else if ch == '>' {
				state = 3
			} else {
				state = 4
			}
		case 2:
			return newRelop(lexer.df.nextLexeme(), LE)
		case 3:
			return newRelop(lexer.df.nextLexeme(), NE)
		case 4:
			return newRelop(lexer.df.nextLexeme(), LT)
		case 5:
			return newRelop(lexer.df.nextLexeme(), EQ)
		case 6:
			if ch == '=' {
				state = 7
			} else {
				state = 8
			}
		case 7:
			return newRelop(lexer.df.nextLexeme(), GE)
		case 8:
			return newRelop(lexer.df.nextLexeme(), GT)
		}
	}
}

func (lexer *Lexer) nextWs(ch byte) Ws {
	state := 22
	for {
		switch state {
		case 22:
			if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
				state = 23
				ch, _ = lexer.df.nextChar()
			} else {
				log.Fatalln("Lexer::nextWs(): invalid input", ch, "and current state is", state)
			}
		case 23:
			if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
				state = 23
				ch, _ = lexer.df.nextChar()
			} else {
				state = 24
			}
		case 24:
			return Ws{}
		}
	}
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalln("main():", err)
	}
//	fmt.Println(dir)
	file, err := os.Open(dir + "/" + "program.data")
	if err != nil {
		log.Fatalln("main():", err)
	}
	defer file.Close()
	lexer := newLexer(4096, file)
//	fmt.Printf("%s\n", string(lexer.df.buf[0]))
	for tok := lexer.nextToken(); tok != nil; tok = lexer.nextToken() {
		if _, ok := tok.(Ws); !ok {
			fmt.Println(reflect.ValueOf(tok), reflect.TypeOf(tok))
		}
	}
}
