package main
import (
	"log"
	"fmt"
	"io"
	"unicode"
	"bytes"
	"sync/atomic"
)

/****************************Env*******************************/
type Env struct {
	table map[string]*Symbol
	pre *Env
}

func NewEnv(pre *Env) *Env {
	return &Env{table:map[string]*Symbol{}, pre:pre}
}

func (env *Env) get(key string) *Symbol {
	for scope := env; scope != nil; scope = scope.pre {
		if symbol, ok := scope.table[key]; ok {
			return symbol
		}
	}
	return nil
}

func (env *Env) put(key string, symbol *Symbol) {
	if key == "" || symbol == nil {
		log.Fatalln("Env::put()", "key==", key, ",symbol==", symbol)
	}
	env.table[key] = symbol
}

type Symbol struct {
	Type string
}

func NewSymbol() *Symbol {
	return &Symbol{}
}
/******************************************************************Parser*********************************************/
/********Node**************/
type Node interface {

}

/*******expression*******/
type Expr interface {
	LValue() Node
	RValue() Node
}

type Op struct {
	Tok Token
	Y Expr
	Z Expr
	T Label
}

func (op *Op) LValue() Node {
	log.Fatalln("Op::LValue():", "no LValue operation.")
	return nil
}

func (op *Op) RValue() Node {
	y := op.Y.RValue()
	z := op.Z.RValue()
	op.T = newLabel()
	var ty Label
	var tz Label
	switch y.(type) {
	case Op:
		ty = y.(Op).T
	default:
		log.Fatalln("Op::RValue():", "y.(type) == " + y.(type))
}
fmt.Println("t" + op.Tok + "=t" + y. + "+t" +)
return newOp(op.Tok, y, z)
}

func newOp(token Token, y Expr, z Expr) *Op {
	return &Op{token, y, z}
}

type Stmt interface {
	Gen()
}

type Label int
var inc int32
func newLabel() Label {
	return Label(atomic.AddInt32(&inc, 1))
}
/*********************If********************/
type If struct {
	E Expr
	S Stmt
	After Label
}

func newIf(E Expr, S Stmt) *If {
	return &If(E, S, newLabel())
}
func (i *If) Gen() {
	t := i.E.RValue()
	fmt.Println("IfFalse", t, "goto", i.After)
	i.S.Gen()
	fmt.Println(i.After+":")
}
/*****************While****************/
type While struct {
	E Expr
	S Stmt
	Begain Label
	After Label
}

func newWhile(E Expr, S Stmt) *While {
	return &While{E, S, newLabel(), newLabel()}
}

func (w *While) Gen() {
	fmt.Println(w.Begain + ":")
	t := w.E.RValue()
	fmt.Println("IfFalse", t, "goto", w.After)
	w.S.Gen()
	fmt.Println("goto", w.Begain)
	fmt.Println(w.After + ":")
}
/************Do*********************/
type Do struct {
	E Expr
	S Stmt
	Begain Label
	After Label
}

func newDo(E Expr, S Stmt) *Do {
	return &Do{E, S, newLabel(), newLabel()}
}

func (d *Do) Gen() {

}

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

var top *Env
func (parser *Parser) program() {
	top = nil
	parser.block()
	fmt.Print("\n")
}

func (parser *Parser) block() {
	parser.match(NewToken('{'))
	saved := top
	top = NewEnv(top)
	fmt.Print("{ ")
	parser.decls()
	parser.stmts()
	parser.match(NewToken('}'))

	top = saved
	fmt.Print("} ")
}
//decls->decls decl| e
// <=>
//decls->declsRest
//declsRest->decl declsRest | e
func (parser *Parser) decls() {
	parser.declsRest()
}

func (parser *Parser) decl() {
	typ := parser.lookahead.(Word)
	parser.match(parser.lookahead)
	id := parser.lookahead.(Word)
	parser.match(parser.lookahead)
	parser.match(NewToken(';'))

	s := NewSymbol()
	s.Type = typ.Lexeme
	top.put(id.Lexeme, s)
	//	fmt.Println("top put:", top, top.pre, id.Lexeme, s)
}

func (parser *Parser) declsRest() {
	if t, ok := parser.lookahead.(Word); ok && t.TAG == TYPE {
		parser.decl()
		parser.declsRest()
	} else {
		// do nothing
	}
}

// stmts->stmts stmt | e
// <=>
// stmts->stmtsRest
// stmtsRest->stmt stmtsRest | e
func (parser *Parser) stmts() {
	parser.stmtsRest()
}

func (parser *Parser) stmt() {
	if c, ok := parser.lookahead.(Token); ok && c.TAG == Tag('{') {
		parser.block()
	} else if _, ok := parser.lookahead.(Word); ok {
		parser.factor()
		parser.match(NewToken(';'))
		fmt.Print("; ")
	} else {
		log.Fatalln("Parser::stmt(), syntax error, parser.lookahead ==", parser.lookahead)
	}
}

func (parser *Parser) stmtsRest() {
	if c, ok := parser.lookahead.(Token); ok && c.TAG == Tag('{'){
		parser.stmt()
		parser.stmtsRest()
	} else if _, ok := parser.lookahead.(Word); ok {
		parser.stmt()
		parser.stmtsRest()
	} else {
		// do nothing
	}
}

func (parser *Parser) factor() {
	id := parser.lookahead.(Word)
	parser.match(id)
	s := top.get(id.Lexeme)
	if s == nil {
		log.Fatal("factor():", "top.get(\"", id.Lexeme, "\") returned nil. top == ", top, "\n")
	}
	fmt.Print(id.Lexeme, ":", s.Type)
}

func (parser *Parser) match(c interface{}) {
	if parser.lookahead == c {
		//		{
		//			if t, ok := c.(Token); ok {
		//				fmt.Printf("\n<%c> matched\n", t.TAG)
		//			} else if t, ok := c.(Word); ok {
		//				fmt.Printf("\n<%d, %s> matched\n", t.TAG, t.Lexeme)
		//			} else if t, ok := c.(Num); ok {
		//				fmt.Printf("\n<%d, %d> matched\n", t.TAG, t.Value)
		//			}
		//		}
		parser.lookahead = parser.lexer.Scan()
		if parser.lookahead == nil {
			return
		}
	} else {
		if t, ok := c.(Token); ok {
			log.Fatalf("match(): syntax error, parser.lookahead is <%v>, c is <%c>\n", parser.lookahead, t.TAG)
		} else if t, ok := c.(Word); ok {
			log.Fatalf("match(): syntax error, parser.lookahead is <%v>, c is <%d, %s>\n", parser.lookahead, t.TAG, t.Lexeme)
		} else if t, ok := c.(Num); ok {
			log.Fatalf("match(): syntax error, parser.lookahead is <%v>, c is <%d, %d>\n", parser.lookahead, t.TAG, t.Value)
		}
	}
}
/*********************************Lexer*************************/
type Tag int

const (
	NUM Tag = 256
	ID Tag = 257
	TRUE Tag = 258
	FALSE Tag = 259
	TYPE Tag = 260
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

func NewNum(value int) Num {
	return Num{NUM, value}
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
	line int
	peek byte
}

func NewLexer() *Lexer {
	return &Lexer{
		Words:map[string]interface{}{
			"true": NewWord(TRUE, "true"),
			"false": NewWord(FALSE, "false"),
			"int": NewWord(TYPE, "int"),
			"char": NewWord(TYPE, "char"),
			"bool": NewWord(TYPE, "bool"),
			"double": NewWord(TYPE, "double"),
			"float": NewWord(TYPE, "float"),
		},
		line:0,
		peek:' ',
	}
}

func (lexer *Lexer) Scan() interface{} {
	for {
		// omit the blank symbol
		if lexer.peek == ' ' || lexer.peek == '\t' {
			_, err := fmt.Scanf("%c", &lexer.peek)
			if err == io.EOF {
				return nil
			}
			if err != nil {
				log.Fatalln("Scan():", err)
			}
			continue
		} else if lexer.peek == '\n' {
			lexer.line++
			_, err := fmt.Scanf("%c", &lexer.peek)
			if err == io.EOF {
				return nil
			}
			if err != nil {
				log.Fatalln("Scan():", err)
			}
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
			return NewNum(v)
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
		//		fmt.Printf("\nscaning... <%c>\n", lexer.peek)
		tok := NewToken(Tag(lexer.peek))
		lexer.peek = ' '
		return tok
	}
}

func main() {
	parser := NewParser()
	parser.program()
}