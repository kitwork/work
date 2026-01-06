package work

import (
	"fmt"
	"reflect"
	"regexp"
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
	TokenSafeDot

	TokenLParen
	TokenRParen

	TokenQuestion
	TokenColon

	TokenTrue
	TokenFalse

	TokenEQ
	TokenNEQ
	TokenGT
	TokenLT
	TokenGTE
	TokenLTE

	TokenBang
	TokenNullCoalesce
	TokenOR
	TokenAND
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

func NewLexer(s string) *Lexer { return &Lexer{src: []rune(s)} }

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
	for ch := l.peek(); ch == ' ' || ch == '\t' || ch == '\n'; ch = l.peek() {
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
		text := string(l.src[start:l.pos])

		switch text {
		case "true":
			return Token{Kind: TokenTrue}
		case "false":
			return Token{Kind: TokenFalse}
		}

		return Token{Kind: TokenIdent, Text: text}
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
		if l.peek() == '|' {
			l.pos++
			return Token{Kind: TokenOR}
		}
		return Token{Kind: TokenPipe}
	case '&':
		if l.peek() == '&' {
			l.pos++
			return Token{Kind: TokenAND}
		}
	case '.':
		return Token{Kind: TokenDot}
	case '(':
		return Token{Kind: TokenLParen}
	case ')':
		return Token{Kind: TokenRParen}
	case '?':
		if l.peek() == '.' {
			l.pos++
			return Token{Kind: TokenSafeDot}
		}
		if l.peek() == '?' {
			l.pos++
			return Token{Kind: TokenNullCoalesce}
		}
		return Token{Kind: TokenQuestion}

	case ':':
		return Token{Kind: TokenColon}
	case '=':
		if l.peek() == '=' {
			l.pos++
			return Token{Kind: TokenEQ}
		}
	case '!':
		if l.peek() == '=' {
			l.pos++
			return Token{Kind: TokenNEQ}
		}
		return Token{Kind: TokenBang}
	case '>':
		if l.peek() == '=' {
			l.pos++
			return Token{Kind: TokenGTE}
		}
		return Token{Kind: TokenGT}
	case '<':
		if l.peek() == '=' {
			l.pos++
			return Token{Kind: TokenLTE}
		}
		return Token{Kind: TokenLT}
	}

	panic("invalid token")
}

func isLetter(ch rune) bool { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' }
func isDigit(ch rune) bool  { return ch >= '0' && ch <= '9' }
func isLetterOrDigit(ch rune) bool {
	return isLetter(ch) || isDigit(ch)
}

/* =======================
   AST
======================= */

type Expr interface{}

type UnaryExpr struct {
	Op TokenKind // TokenBang
	X  Expr
}

type NumberExpr struct{ Value string }
type StringExpr struct{ Value string }
type VarExpr struct {
	Path []string
	Safe []bool // Safe[i] = true nếu truy cập path[i] bằng ?.
}

type BinaryExpr struct {
	Left  Expr
	Op    TokenKind
	Right Expr
}

type CompareExpr struct {
	Left  Expr
	Op    TokenKind
	Right Expr
}

type TernaryExpr struct {
	Cond  Expr
	True  Expr
	False Expr
}

type LogicalOrExpr struct {
	Left  Expr
	Right Expr
}

type LogicalAndExpr struct {
	Left  Expr
	Right Expr
}

type NullCoalesceExpr struct {
	Left  Expr
	Right Expr
}

type PipeExpr struct {
	Base  Expr
	Pipes []PipeCall
}

type BoolExpr struct {
	Value bool
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

func (p *Parser) next() { p.cur = p.l.Next() }

func (p *Parser) parsePrimary() Expr {
	t := p.cur
	switch t.Kind {
	case TokenTrue:
		p.next()
		return &BoolExpr{Value: true}

	case TokenFalse:
		p.next()
		return &BoolExpr{Value: false}
	case TokenNumber:
		p.next()
		return &NumberExpr{Value: t.Text}
	case TokenString:
		p.next()
		return &StringExpr{Value: t.Text}
	case TokenIdent:
		name := strings.TrimPrefix(t.Text, "$")
		path := []string{name}
		safe := []bool{false}

		p.next()

		for p.cur.Kind == TokenDot || p.cur.Kind == TokenSafeDot {
			isSafe := p.cur.Kind == TokenSafeDot
			p.next()

			if p.cur.Kind != TokenIdent {
				panic("invalid path")
			}

			path = append(path, p.cur.Text)
			safe = append(safe, isSafe)

			p.next()
		}

		return &VarExpr{
			Path: path,
			Safe: safe,
		}

	case TokenLParen:
		p.next()
		e := p.parseExpr()
		p.next()
		return e
	}
	panic("invalid expression")
}

func (p *Parser) parseMul() Expr {
	expr := p.parseUnary()
	for p.cur.Kind == TokenStar || p.cur.Kind == TokenSlash {
		op := p.cur.Kind
		p.next()
		right := p.parseUnary()
		expr = &BinaryExpr{Left: expr, Op: op, Right: right}
	}
	return expr
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

func (p *Parser) parseCompare() Expr {
	left := p.parseAdd()
	for p.cur.Kind == TokenEQ || p.cur.Kind == TokenNEQ || p.cur.Kind == TokenGT ||
		p.cur.Kind == TokenLT || p.cur.Kind == TokenGTE || p.cur.Kind == TokenLTE {
		op := p.cur.Kind
		p.next()
		right := p.parseAdd()
		left = &CompareExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseAnd() Expr {
	left := p.parseCompare()
	for p.cur.Kind == TokenAND {
		p.next()
		right := p.parseCompare()
		left = &LogicalAndExpr{Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseOr() Expr {
	left := p.parseAnd()
	for p.cur.Kind == TokenOR {
		p.next()
		right := p.parseAnd()
		left = &LogicalOrExpr{Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseNullish() Expr {
	left := p.parseOr()
	for p.cur.Kind == TokenNullCoalesce {
		p.next()
		right := p.parseOr()
		left = &NullCoalesceExpr{Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseTernary() Expr {
	cond := p.parseNullish() // 🔥 QUAN TRỌNG
	if p.cur.Kind == TokenQuestion {
		p.next()
		t := p.parseTernary()
		p.next() // :
		f := p.parseTernary()
		return &TernaryExpr{Cond: cond, True: t, False: f}
	}
	return cond
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
	return p.parsePipe(p.parseTernary())
}

func (p *Parser) parseUnary() Expr {
	if p.cur.Kind == TokenBang {
		op := p.cur.Kind
		p.next()
		return &UnaryExpr{
			Op: op,
			X:  p.parseUnary(), // 🔥 recursion
		}
	}
	return p.parsePrimary()
}

/* =======================
   EVALUATOR
======================= */

type Evaluator struct {
	Scope map[string]any
	Pipes map[string]interface{}
}

func NewEvaluator(scope map[string]any, pipes map[string]interface{}) *Evaluator {
	return &Evaluator{Scope: scope, Pipes: pipes}
}

func (e *Evaluator) Eval(s string) (any, error) {
	p := NewParser(s)
	ast := p.parseExpr()

	rv, err := e.eval(ast)
	if err != nil {
		return nil, err
	}

	// 🔑 QUAN TRỌNG
	if !rv.IsValid() {
		return nil, nil
	}

	return rv.Interface(), nil
}

func (e *Evaluator) Render(s string) (any, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}

	// unwrap {{ }}
	if strings.HasPrefix(s, "{{") && strings.HasSuffix(s, "}}") {
		expr := strings.TrimSpace(s[2 : len(s)-2])
		return e.Eval(expr)
	}

	// expression thuần
	if looksLikeExpression(s) {
		return e.Eval(s)
	}

	// string template
	return e.Template(s)
}

func (e *Evaluator) Template(s string) (any, error) {
	// Regex tìm tất cả {{ ... }}
	re := regexp.MustCompile(`\{\{\s*(.*?)\s*\}\}`)

	result := re.ReplaceAllStringFunc(s, func(match string) string {
		// Lấy expression bên trong {{ ... }}
		inner := re.FindStringSubmatch(match)[1]

		// fmt.Println(inner)
		// Eval expression
		val, err := e.Eval(inner)
		if err != nil {
			fmt.Println(err)
			// Nếu lỗi, giữ nguyên template để debug
			return match
		}

		return fmt.Sprintf("%v", val)
	})

	return result, nil
}

func (e *Evaluator) eval(expr Expr) (reflect.Value, error) {
	switch n := expr.(type) {

	case *BoolExpr:
		return reflect.ValueOf(n.Value), nil

	case *UnaryExpr:
		v, err := e.eval(n.X)
		if err != nil {
			return reflect.Value{}, err
		}

		switch n.Op {
		case TokenBang:
			return reflect.ValueOf(!isTruthy(v)), nil
		}

		return reflect.Value{}, fmt.Errorf("unknown unary op")

	case *NumberExpr:
		f, _ := strconv.ParseFloat(n.Value, 64)
		return reflect.ValueOf(f), nil
	case *StringExpr:
		// Chuỗi có thể chứa {{ ... }} expressions
		s := n.Value
		var sb strings.Builder

		start := 0
		for {
			open := strings.Index(s[start:], "{{")
			if open == -1 {
				sb.WriteString(s[start:])
				break
			}
			open += start
			sb.WriteString(s[start:open])
			close := strings.Index(s[open:], "}}")
			if close == -1 {
				// không đóng, coi phần còn lại là literal
				sb.WriteString(s[open:])
				break
			}
			close += open

			// lấy biểu thức giữa {{ ... }}
			exprStr := strings.TrimSpace(s[open+2 : close])
			val, err := e.Eval(exprStr)
			if err != nil {
				return reflect.Value{}, err
			}
			sb.WriteString(fmt.Sprintf("%v", val))

			start = close + 2
		}
		return reflect.ValueOf(sb.String()), nil

	case *VarExpr:
		v, err := resolveWithSafe(
			reflect.ValueOf(e.Scope),
			n.Path,
			n.Safe,
		)
		if err != nil {
			return reflect.Value{}, err
		}
		return v, nil

	case *BinaryExpr:
		lv, _ := e.eval(n.Left)
		rv, _ := e.eval(n.Right)
		a, b := toFloat(lv), toFloat(rv)
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
	case *CompareExpr:
		lv, _ := e.eval(n.Left)
		rv, _ := e.eval(n.Right)
		a, b := toFloat(lv), toFloat(rv)
		var res bool
		switch n.Op {
		case TokenEQ:
			res = a == b
		case TokenNEQ:
			res = a != b
		case TokenGT:
			res = a > b
		case TokenLT:
			res = a < b
		case TokenGTE:
			res = a >= b
		case TokenLTE:
			res = a <= b
		}
		return reflect.ValueOf(res), nil

	case *LogicalOrExpr:
		lv, err := e.eval(n.Left)
		if err != nil {
			return reflect.Value{}, err
		}

		if isTruthy(lv) {
			return lv, nil // 👈 trả về LEFT
		}

		return e.eval(n.Right)

	case *LogicalAndExpr:
		lv, err := e.eval(n.Left)
		if err != nil {
			return reflect.Value{}, err
		}

		if !isTruthy(lv) {
			return lv, nil // 👈 trả về LEFT
		}

		return e.eval(n.Right) // 👈 trả về RIGHT

	case *NullCoalesceExpr:
		lv, err := e.eval(n.Left)
		if err == nil && lv.IsValid() && !(lv.Kind() == reflect.Interface && lv.IsNil()) {
			return lv, nil
		}
		return e.eval(n.Right)

	case *TernaryExpr:
		cond, err := e.eval(n.Cond)
		if err != nil || !cond.IsValid() {
			return e.eval(n.False)
		}

		if isTruthy(cond) { // 👈 dùng isTruthy, KHÔNG dùng toBool trực tiếp
			return e.eval(n.True)
		}
		return e.eval(n.False)

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
			args := []reflect.Value{}
			for _, a := range call.Args {
				av, err := e.eval(a)
				if err != nil {
					return reflect.Value{}, err
				}
				args = append(args, reflect.ValueOf(av))
			}
			args = append(args, reflect.ValueOf(val))
			out := fv.Call(args)
			if len(out) == 0 || !out[0].IsValid() {
				return reflect.Value{}, fmt.Errorf("pipe %s returned invalid value", call.Name)
			}
			val = out[0].Interface().(reflect.Value)
		}
		return val, nil

	}
	return reflect.Value{}, fmt.Errorf("unknown expr")
}

/* =======================
   HELPERS
======================= */

func deepGet(m map[string]any, path []string) (any, error) {
	var cur any = m

	for _, k := range path {
		switch v := cur.(type) {

		case map[string]any:
			val, ok := v[k]
			if !ok {
				return nil, fmt.Errorf("key %s not found", k)
			}
			cur = val

		case map[string]string:
			val, ok := v[k]
			if !ok {
				return nil, fmt.Errorf("key %s not found", k)
			}
			cur = val

		default:
			return nil, fmt.Errorf("cannot access key %s on %T", k, cur)
		}
	}

	return cur, nil
}

func isTruthy(v reflect.Value) bool {
	if !v.IsValid() {
		return false
	}

	for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Bool:
		return v.Bool()
	case reflect.String:
		return v.String() != ""
	case reflect.Int, reflect.Int64:
		return v.Int() != 0
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0
	default:
		return true
	}
}

func resolve(v reflect.Value, path []string) (reflect.Value, error) {
	for _, key := range path {

		// unwrap interface / pointer
		for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return reflect.Value{}, fmt.Errorf("nil encountered while resolving %q", key)
			}
			v = v.Elem()
		}

		switch v.Kind() {

		case reflect.Map:
			if v.Type().Key().Kind() != reflect.String {
				return reflect.Value{}, fmt.Errorf("map key is %s, not string", v.Type().Key())
			}

			val := v.MapIndex(reflect.ValueOf(key))
			if !val.IsValid() {
				return reflect.Value{}, fmt.Errorf("key %q not found", key)
			}
			v = val

		case reflect.Struct:
			field, ok := findStructField(v, key)
			if !ok {
				return reflect.Value{}, fmt.Errorf("field %q not found", key)
			}
			v = field

		default:
			return reflect.Value{}, fmt.Errorf("cannot access %q on %s", key, v.Kind())
		}
	}

	return v, nil
}

func resolveSafeExpr(e *Evaluator, expr Expr) reflect.Value {
	switch v := expr.(type) {
	case *VarExpr:
		return resolveSafe(reflect.ValueOf(e.Scope), v.Path)
	default:
		rv, err := e.eval(expr)
		if err != nil {
			return reflect.Value{}
		}
		return rv
	}
}

func resolveWithSafe(v reflect.Value, path []string, safe []bool) (reflect.Value, error) {
	for i, key := range path {

		isSafe := false
		if i < len(safe) {
			isSafe = safe[i]
		}

		for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
			if v.IsNil() {
				if isSafe {
					return reflect.Value{}, nil
				}
				return reflect.Value{}, fmt.Errorf("nil while resolving %q", key)
			}
			v = v.Elem()
		}

		switch v.Kind() {

		case reflect.Map:
			val := v.MapIndex(reflect.ValueOf(key))
			if !val.IsValid() {
				if isSafe {
					return reflect.Value{}, nil
				}
				return reflect.Value{}, fmt.Errorf("key %q not found", key)
			}
			v = val

		case reflect.Struct:
			field, ok := findStructField(v, key)
			if !ok {
				if isSafe {
					return reflect.Value{}, nil
				}
				return reflect.Value{}, fmt.Errorf("field %q not found", key)
			}
			v = field

		default:
			if isSafe {
				return reflect.Value{}, nil
			}
			return reflect.Value{}, fmt.Errorf("cannot access %q on %s", key, v.Kind())
		}
	}

	return v, nil
}

func resolveSafe(v reflect.Value, path []string) reflect.Value {
	for _, key := range path {

		for v.Kind() == reflect.Interface || v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return reflect.Value{}
			}
			v = v.Elem()
		}

		switch v.Kind() {

		case reflect.Map:
			val := v.MapIndex(reflect.ValueOf(key))
			if !val.IsValid() {
				return reflect.Value{}
			}
			v = val

		case reflect.Struct:
			field, ok := findStructField(v, key)
			if !ok {
				return reflect.Value{}
			}
			v = field

		default:
			return reflect.Value{}
		}
	}

	return v
}

func findStructField(v reflect.Value, key string) (reflect.Value, bool) {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		// bỏ field private
		if f.PkgPath != "" {
			continue
		}

		// json tag
		if tag := f.Tag.Get("json"); tag != "" {
			name := strings.Split(tag, ",")[0]
			if name == key {
				return v.Field(i), true
			}
		}

		if f.Name == key {
			return v.Field(i), true
		}
	}

	return reflect.Value{}, false
}

func toFloat(v reflect.Value) float64 {
	switch v.Kind() {
	case reflect.Int, reflect.Int64:
		return float64(v.Int())
	case reflect.Float32, reflect.Float64:
		return v.Float()
	case reflect.Bool:
		if v.Bool() {
			return 1
		}
		return 0
	case reflect.String:
		f, _ := strconv.ParseFloat(v.String(), 64)
		return f
	default:
		return 0
	}
}

func looksLikeExpression(s string) bool {
	if s == "" {
		return false
	}

	// literal
	if s == "true" || s == "false" || s == "null" {
		return true
	}

	// number
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}

	// unary
	if strings.HasPrefix(s, "!") {
		return true
	}

	// grouping
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		return true
	}

	// operators
	if strings.ContainsAny(s, "&|?:=!<>+-*/") {
		return true
	}

	// variable
	if strings.HasPrefix(s, "$") {
		return true
	}

	return false
}
