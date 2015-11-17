package main
type Attribute int
const (
	LT Attribute = 256 + iota
	LE
	EQ
	NE
	GT
	GE
)

type Keyword int
const (
	IF Keyword = 266 + iota
	THEN
	ELSE
	WHILE
	DO
	FOR
	REST
)

type Token interface {

}

type Id struct {
	keyword Keyword
	lexeme  string
}

func newId(keyword Keyword, lexeme string) *Id {
	return &Id{keyword, lexeme}
}

type Number struct {
	lexeme string
}

func newNumber(lexeme string) *Number {
	return &Number{lexeme}
}

type Relop struct {
	lexeme    string
	attribute Attribute
}

func newRelop(lexeme string, attribute Attribute) *Relop {
	return &Relop{lexeme, attribute}
}

type Ws struct {

}


