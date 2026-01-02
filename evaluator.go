package work

import (
	"fmt"
	"html/template"
	"reflect"
	"strconv"
	"strings"
)

/* =======================
   TOKEN
======================= */

type TokenKind int

const (
	TokenEOF TokenKind = iota
	TokenIdent
	TokenNumber
	TokenString

	TokenPlus
	TokenMinus
	TokenStar
	TokenSlash

	TokenPipe
	TokenDot
	TokenLParen
	TokenRParen
)

type Token struct {
	Kind TokenKind
	Text string
}

/* =======================
   LEXER
======================= */

type Lexer struct {
	src []rune
	pos int
}

func NewLexer(s string) *Lexer {
	return &Lexer{src: []rune(s)}
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

func (l *Lexer) next() rune {
	ch := l.peek()
	l.pos++
	return ch
}

func (l *Lexer) skipSpace() {
	for l.peek() == ' ' || l.peek() == '\t' || l.peek() == '\n' {
		l.pos++
	}
}

func (l *Lexer) Next() Token {
	l.skipSpace()
	ch := l.next()

	if ch == 0 {
		return Token{Kind: TokenEOF}
	}

	if ch == '$' || isLetter(ch) {
		start := l.pos - 1
		for isLetterOrDigit(l.peek()) {
			l.pos++
		}
		return Token{Kind: TokenIdent, Text: string(l.src[start:l.pos])}
	}

	if isDigit(ch) {
		start := l.pos - 1
		for isDigit(l.peek()) {
			l.pos++
		}
		return Token{Kind: TokenNumber, Text: string(l.src[start:l.pos])}
	}

	if ch == '"' {
		start := l.pos
		for l.peek() != '"' && l.peek() != 0 {
			l.pos++
		}
		l.pos++
		return Token{Kind: TokenString, Text: string(l.src[start : l.pos-1])}
	}

	switch ch {
	case '+':
		return Token{Kind: TokenPlus}
	case '-':
		return Token{Kind: TokenMinus}
	case '*':
		return Token{Kind: TokenStar}
	case '/':
		return Token{Kind: TokenSlash}
	case '|':
		return Token{Kind: TokenPipe}
	case '.':
		return Token{Kind: TokenDot}
	case '(':
		return Token{Kind: TokenLParen}
	case ')':
		return Token{Kind: TokenRParen}
	}

	panic("invalid token")
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}
func isDigit(ch rune) bool { return ch >= '0' && ch <= '9' }
func isLetterOrDigit(ch rune) bool {
	return isLetter(ch) || isDigit(ch)
}

/* =======================
   AST
======================= */

type Expr interface{}

type NumberExpr struct{ Value string }
type StringExpr struct{ Value string }

type VarExpr struct {
	Path []string
}

type BinaryExpr struct {
	Left  Expr
	Op    TokenKind
	Right Expr
}

type PipeExpr struct {
	Base  Expr
	Pipes []PipeCall
}

type PipeCall struct {
	Name string
	Args []Expr
}

/* =======================
   PARSER
======================= */

type Parser struct {
	l   *Lexer
	cur Token
}

func NewParser(s string) *Parser {
	p := &Parser{l: NewLexer(s)}
	p.next()
	return p
}

func (p *Parser) next() {
	p.cur = p.l.Next()
}

func (p *Parser) parsePrimary() Expr {
	t := p.cur

	switch t.Kind {

	case TokenNumber:
		p.next()
		return &NumberExpr{Value: t.Text}

	case TokenString:
		p.next()
		return &StringExpr{Value: t.Text}

	case TokenIdent:
		name := strings.TrimPrefix(t.Text, "$")
		path := []string{name}
		p.next()

		for p.cur.Kind == TokenDot {
			p.next()
			if p.cur.Kind != TokenIdent {
				panic("invalid path")
			}
			path = append(path, p.cur.Text)
			p.next()
		}
		return &VarExpr{Path: path}

	case TokenLParen:
		p.next()
		e := p.parseExpr()
		p.next()
		return e
	}

	panic("invalid expression")
}

func (p *Parser) parseMul() Expr {
	left := p.parsePrimary()
	for p.cur.Kind == TokenStar || p.cur.Kind == TokenSlash {
		op := p.cur.Kind
		p.next()
		right := p.parsePrimary()
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseAdd() Expr {
	left := p.parseMul()
	for p.cur.Kind == TokenPlus || p.cur.Kind == TokenMinus {
		op := p.cur.Kind
		p.next()
		right := p.parseMul()
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parsePipe(base Expr) Expr {
	for p.cur.Kind == TokenPipe {
		p.next()
		name := p.cur.Text
		p.next()

		var args []Expr
		for p.cur.Kind != TokenPipe && p.cur.Kind != TokenEOF {
			args = append(args, p.parsePrimary())
		}

		base = &PipeExpr{
			Base:  base,
			Pipes: []PipeCall{{Name: name, Args: args}},
		}
	}
	return base
}

func (p *Parser) parseExpr() Expr {
	return p.parsePipe(p.parseAdd())
}

/* =======================
   EVALUATOR
======================= */

type Evaluator struct {
	Scope map[string]any
	Pipes template.FuncMap
}

func NewEvaluator(scope map[string]any, pipes template.FuncMap) *Evaluator {
	return &Evaluator{Scope: scope, Pipes: pipes}
}

func (e *Evaluator) Eval(s string) (any, error) {
	p := NewParser(s)
	ast := p.parseExpr()

	rv, err := e.eval(ast)
	if err != nil {
		return nil, err
	}
	return rv.Interface(), nil
}

func (e *Evaluator) eval(expr Expr) (reflect.Value, error) {
	switch n := expr.(type) {

	case *NumberExpr:
		f, err := strconv.ParseFloat(n.Value, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(f), nil

	case *StringExpr:
		return reflect.ValueOf(n.Value), nil

	case *VarExpr:
		v, err := deepGet(e.Scope, n.Path)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(v), nil

	case *BinaryExpr:
		lv, err := e.eval(n.Left)
		if err != nil {
			return reflect.Value{}, err
		}
		rv, err := e.eval(n.Right)
		if err != nil {
			return reflect.Value{}, err
		}

		a := toFloat(lv)
		b := toFloat(rv)

		switch n.Op {
		case TokenPlus:
			return reflect.ValueOf(a + b), nil
		case TokenMinus:
			return reflect.ValueOf(a - b), nil
		case TokenStar:
			return reflect.ValueOf(a * b), nil
		case TokenSlash:
			return reflect.ValueOf(a / b), nil
		}

		// =======================
	// PIPE EVALUATION (FIXED)
	// =======================

	case *PipeExpr:
		val, err := e.eval(n.Base)
		if err != nil {
			return reflect.Value{}, err
		}

		for _, call := range n.Pipes {

			fn, ok := e.Pipes[call.Name]
			if !ok {
				return reflect.Value{}, fmt.Errorf("pipe not found: %s", call.Name)
			}

			fv := reflect.ValueOf(fn)
			if fv.Kind() != reflect.Func {
				return reflect.Value{}, fmt.Errorf("pipe %s is not func", call.Name)
			}

			// ==== QUAN TRỌNG ====
			// Pipe signature:
			// func(arg1 reflect.Value, ..., value reflect.Value) reflect.Value

			args := []reflect.Value{}

			for _, a := range call.Args {
				av, err := e.eval(a)
				if err != nil {
					return reflect.Value{}, err
				}
				// BỌC reflect.Value
				args = append(args, reflect.ValueOf(av))
			}

			// value luôn ở cuối
			args = append(args, reflect.ValueOf(val))

			out := fv.Call(args)

			if len(out) == 0 || out[0].Kind() != reflect.Struct && !out[0].IsValid() {
				return reflect.Value{}, fmt.Errorf("pipe %s returned invalid value", call.Name)
			}

			// unwrap: pipe trả reflect.Value
			val = out[0].Interface().(reflect.Value)
		}

		return val, nil

	}

	return reflect.Value{}, fmt.Errorf("unknown expression")
}

/* =======================
   HELPERS
======================= */

func deepGet(m map[string]any, path []string) (any, error) {
	var cur any = m
	for _, k := range path {
		switch v := cur.(type) {
		case map[string]any:
			cur = v[k]
		default:
			return nil, fmt.Errorf("cannot access %s", k)
		}
	}
	return cur, nil
}

func toFloat(v reflect.Value) float64 {
	switch v.Kind() {
	case reflect.Int, reflect.Int64:
		return float64(v.Int())
	case reflect.Float32, reflect.Float64:
		return v.Float()
	case reflect.String:
		f, _ := strconv.ParseFloat(v.String(), 64)
		return f
	default:
		return 0
	}
}
