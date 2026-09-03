package main

import (
	"math/big"
	"strconv"
	"strings"
)

type token struct {
	text, kind string
}
type statement struct {
	tokens []token
	raw    []byte
}
type sqlType struct {
	kind    string
	p, s, n int
}
type effect struct {
	class, key, schema string
	concurrent         bool
	minRisk            int
	stmtSHA            string
}
type upState struct {
	created map[string]map[string]sqlType
	added   map[string]sqlType
}
type parser struct {
	t []token
	i int
}

func (p *parser) peek(s string) bool { return p.i < len(p.t) && strings.EqualFold(p.t[p.i].text, s) }
func (p *parser) take(s string) bool {
	if p.peek(s) {
		p.i++
		return true
	}
	return false
}
func (p *parser) exact(s string) bool {
	if p.i < len(p.t) && p.t[p.i].text == s {
		p.i++
		return true
	}
	return false
}
func (p *parser) ident() (string, bool) {
	if p.i >= len(p.t) {
		return "", false
	}
	s := p.t[p.i].text
	if !validIdentifier(s) {
		return "", false
	}
	p.i++
	return s, true
}
func (p *parser) qname() (string, string, bool) {
	a, ok := p.ident()
	if !ok || !p.exact(".") {
		return "", "", false
	}
	b, ok := p.ident()
	return a, b, ok
}
func (p *parser) end() bool { return p.exact(";") && p.i == len(p.t) }
func scanSQL(b []byte) ([]statement, error) {
	if !validUTF8NoBOMNUL(b) {
		return nil, verr("MIG011_SQL_LEXICAL", "R027", "SQL must be UTF-8 without BOM/NUL")
	}
	var out []statement
	var ts []token
	first := -1
	for i := 0; i < len(b); {
		c := b[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f' {
			i++
			continue
		}
		if i+1 < len(b) && c == '-' && b[i+1] == '-' {
			i += 2
			for i < len(b) && b[i] != '\n' {
				i++
			}
			continue
		}
		if i+1 < len(b) && c == '/' && b[i+1] == '*' {
			d := 1
			i += 2
			for i < len(b) && d > 0 {
				if i+1 < len(b) && b[i] == '/' && b[i+1] == '*' {
					d++
					i += 2
				} else if i+1 < len(b) && b[i] == '*' && b[i+1] == '/' {
					d--
					i += 2
				} else {
					i++
				}
			}
			if d != 0 {
				return nil, verr("MIG011_SQL_LEXICAL", "R027", "unterminated block comment")
			}
			continue
		}
		start := i
		if c == '\\' || c == ':' {
			return nil, verr("MIG011_SQL_LEXICAL", "R027", "psql client surface forbidden")
		}
		if first < 0 {
			first = start
		}
		if c == '\'' {
			i++
			for {
				if i >= len(b) {
					return nil, verr("MIG011_SQL_LEXICAL", "R027", "unterminated string")
				}
				if b[i] == '\\' {
					return nil, verr("MIG011_SQL_LEXICAL", "R040", "backslash string escape forbidden")
				}
				if b[i] == '\'' {
					if i+1 < len(b) && b[i+1] == '\'' {
						i += 2
						continue
					}
					i++
					break
				}
				if b[i] < ' ' || b[i] > 0x7e {
					return nil, verr("MIG011_SQL_LEXICAL", "R040", "string payload must be printable ASCII")
				}
				i++
			}
			ts = append(ts, token{string(b[start:i]), "string"})
			continue
		}
		if c == '"' {
			i++
			for i < len(b) {
				if b[i] == '"' {
					if i+1 < len(b) && b[i+1] == '"' {
						i += 2
						continue
					}
					i++
					break
				}
				i++
			}
			if i > len(b) || b[i-1] != '"' {
				return nil, verr("MIG011_SQL_LEXICAL", "R027", "unterminated quoted identifier")
			}
			ts = append(ts, token{string(b[start:i]), "quoted"})
			continue
		}
		if c == '$' {
			j := i + 1
			for j < len(b) && (b[j] >= 'A' && b[j] <= 'Z' || b[j] >= 'a' && b[j] <= 'z' || b[j] >= '0' && b[j] <= '9' || b[j] == '_') {
				j++
			}
			if j < len(b) && b[j] == '$' {
				tag := string(b[i : j+1])
				k := strings.Index(string(b[j+1:]), tag)
				if k < 0 {
					return nil, verr("MIG011_SQL_LEXICAL", "R027", "unterminated dollar quote")
				}
				i = j + 1 + k + len(tag)
				ts = append(ts, token{string(b[start:i]), "dollar"})
				continue
			}
		}
		if isAlpha(c) || c == '_' {
			i++
			for i < len(b) && (isAlpha(b[i]) || isDigit(b[i]) || b[i] == '_' || b[i] == '$') {
				i++
			}
			ts = append(ts, token{string(b[start:i]), "word"})
			continue
		}
		if isDigit(c) || (c == '-' && i+1 < len(b) && isDigit(b[i+1])) || (c == '+' && i+1 < len(b) && isDigit(b[i+1])) {
			i++
			for i < len(b) && (isDigit(b[i]) || b[i] == '.' || b[i] == '_' || b[i] == 'e' || b[i] == 'E' || b[i] == '+' || b[i] == '-') {
				i++
			}
			ts = append(ts, token{string(b[start:i]), "number"})
			continue
		}
		if i+1 < len(b) && (string(b[i:i+2]) == "<=" || string(b[i:i+2]) == ">=" || string(b[i:i+2]) == "<>") {
			i += 2
			ts = append(ts, token{string(b[start:i]), "op"})
			continue
		}
		i++
		ts = append(ts, token{string(b[start:i]), "punct"})
		if c == ';' {
			out = append(out, statement{append([]token(nil), ts...), append([]byte(nil), b[first:i]...)})
			ts = nil
			first = -1
		}
	}
	if len(ts) > 0 {
		return nil, verr("MIG011_SQL_LEXICAL", "R027", "statement missing semicolon")
	}
	return out, nil
}
func isAlpha(c byte) bool { return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' }
func isDigit(c byte) bool { return c >= '0' && c <= '9' }
func parseType(p *parser) (sqlType, string, bool) {
	if p.take("BOOLEAN") {
		return sqlType{kind: "boolean"}, "", true
	}
	if p.take("SMALLINT") {
		return sqlType{kind: "smallint"}, "", true
	}
	if p.take("INTEGER") {
		return sqlType{kind: "integer"}, "", true
	}
	if p.take("BIGINT") {
		return sqlType{kind: "bigint"}, "", true
	}
	if p.take("TEXT") {
		return sqlType{kind: "text"}, "", true
	}
	if p.take("UUID") {
		return sqlType{kind: "uuid"}, "", true
	}
	if p.take("DATE") {
		return sqlType{kind: "date"}, "", true
	}
	if p.take("TIMESTAMP") {
		if p.take("WITH") && p.take("TIME") && p.take("ZONE") {
			return sqlType{kind: "timestamp"}, "", true
		}
		return sqlType{}, "R051", false
	}
	if p.take("NUMERIC") {
		if !p.exact("(") || p.i >= len(p.t) || !canonPositiveRE.MatchString(p.t[p.i].text) {
			return sqlType{}, "R051", false
		}
		a, _ := strconv.Atoi(p.t[p.i].text)
		p.i++
		if !p.exact(",") || p.i >= len(p.t) || !canonUnsignedRE.MatchString(p.t[p.i].text) {
			return sqlType{}, "R051", false
		}
		s, _ := strconv.Atoi(p.t[p.i].text)
		p.i++
		if !p.exact(")") || a < 1 || a > 38 || s > a {
			return sqlType{}, "R051", false
		}
		return sqlType{kind: "numeric", p: a, s: s}, "", true
	}
	if p.take("VARCHAR") {
		if !p.exact("(") || p.i >= len(p.t) || !canonPositiveRE.MatchString(p.t[p.i].text) {
			return sqlType{}, "R051", false
		}
		n64, e := strconv.ParseInt(p.t[p.i].text, 10, 64)
		p.i++
		if e != nil || !p.exact(")") || n64 < 1 || n64 > 10485760 {
			return sqlType{}, "R051", false
		}
		return sqlType{kind: "varchar", n: int(n64)}, "", true
	}
	return sqlType{}, "R008", false
}
func parseCol(p *parser) (string, sqlType, bool, string, bool) {
	name, ok := p.ident()
	if !ok {
		return "", sqlType{}, false, "R048", false
	}
	typ, rule, ok := parseType(p)
	if !ok {
		return "", sqlType{}, false, rule, false
	}
	p.take("NULL")
	hasDefault := false
	if p.take("DEFAULT") {
		if p.i >= len(p.t) || !validLiteral(p.t[p.i], typ) {
			return "", sqlType{}, false, "R040", false
		}
		p.i++
		hasDefault = true
	}
	return name, typ, hasDefault, "", true
}
func validLiteral(t token, typ sqlType) bool {
	s := t.text
	switch typ.kind {
	case "boolean":
		return s == "TRUE" || s == "FALSE"
	case "smallint", "integer", "bigint":
		if t.kind != "number" || !canonIntegerRE.MatchString(s) {
			return false
		}
		z := new(big.Int)
		if _, ok := z.SetString(s, 10); !ok {
			return false
		}
		lo, hi := "-9223372036854775808", "9223372036854775807"
		if typ.kind == "smallint" {
			lo, hi = "-32768", "32767"
		}
		if typ.kind == "integer" {
			lo, hi = "-2147483648", "2147483647"
		}
		l, _ := new(big.Int).SetString(lo, 10)
		h, _ := new(big.Int).SetString(hi, 10)
		return z.Cmp(l) >= 0 && z.Cmp(h) <= 0
	case "numeric":
		if t.kind != "number" || strings.HasPrefix(s, "+") || strings.ContainsAny(s, "eE_") {
			return false
		}
		neg := strings.HasPrefix(s, "-")
		u := strings.TrimPrefix(s, "-")
		parts := strings.Split(u, ".")
		if len(parts) > 2 || !canonUnsignedRE.MatchString(parts[0]) {
			return false
		}
		frac := ""
		if len(parts) == 2 {
			frac = parts[1]
			if frac == "" || !regexpDigits(frac) {
				return false
			}
		}
		if neg && strings.Trim(u, "0.") == "" {
			return false
		}
		if typ.s == 0 && frac != "" || len(frac) > typ.s {
			return false
		}
		intDigits := 0
		if parts[0] != "0" {
			intDigits = len(parts[0])
		}
		return intDigits <= typ.p-typ.s
	case "text", "varchar":
		if t.kind != "string" {
			return false
		}
		payload, ok := decodeString(s)
		if !ok {
			return false
		}
		return typ.kind != "varchar" || len(payload) <= typ.n
	default:
		return false
	}
}
func regexpDigits(s string) bool {
	for i := range s {
		if !isDigit(s[i]) {
			return false
		}
	}
	return s != ""
}
func decodeString(s string) (string, bool) {
	if len(s) < 2 || s[0] != '\'' || s[len(s)-1] != '\'' {
		return "", false
	}
	var b strings.Builder
	for i := 1; i < len(s)-1; i++ {
		c := s[i]
		if c == '\\' || c < ' ' || c > 0x7e {
			return "", false
		}
		if c == '\'' {
			if i+1 >= len(s)-1 || s[i+1] != '\'' {
				return "", false
			}
			i++
		}
		b.WriteByte(c)
	}
	return b.String(), true
}
func parseUpDDL(s statement, st *upState) (effect, error) {
	p := &parser{t: s.tokens}
	eff := effect{stmtSHA: hashBytes(s.raw), minRisk: 1}
	if p.i < len(p.t) && in(makeSet([]string{"INSERT", "UPDATE", "DELETE", "MERGE", "COPY", "SELECT", "REFRESH", "DO", "CALL", "PREPARE", "EXECUTE"}), strings.ToUpper(p.t[p.i].text)) {
		return eff, badRule("R010")
	}
	if p.i+1 < len(p.t) && p.peek("CREATE") && in(makeSet([]string{"FUNCTION", "PROCEDURE", "MATERIALIZED"}), strings.ToUpper(p.t[p.i+1].text)) {
		return eff, badRule("R010")
	}
	if p.i < len(p.t) && in(makeSet([]string{"SAVEPOINT", "ROLLBACK", "COMMIT", "BEGIN", "SET", "RESET"}), strings.ToUpper(p.t[p.i].text)) {
		return eff, verr("MIG014_TRANSACTION", "R011", "unsupported transaction/session control")
	}
	if p.take("CREATE") {
		unique := p.take("UNIQUE")
		if p.take("TABLE") {
			if unique {
				return eff, badDDL()
			}
			sc, t, ok := p.qname()
			if !ok || !p.exact("(") {
				return eff, badRule("R048")
			}
			cols := map[string]sqlType{}
			for {
				n, ty, _, rule, ok := parseCol(p)
				if !ok {
					return eff, badRule(rule)
				}
				if _, dup := cols[n]; dup {
					return eff, badRule("R048")
				}
				cols[n] = ty
				if p.exact(")") {
					break
				}
				if !p.exact(",") {
					return eff, badRule("R048")
				}
				if len(cols) >= 64 {
					return eff, badRule("R048")
				}
			}
			if !p.end() {
				return eff, badRule("R048")
			}
			key := sc + "." + t
			if _, dup := st.created[key]; dup {
				return eff, badRule("R048")
			}
			st.created[key] = cols
			eff.class = "create_table"
			eff.key = "table:" + key
			eff.schema = sc
			return eff, nil
		}
		if !p.take("INDEX") {
			return eff, badRule("R041")
		}
		con := p.take("CONCURRENTLY")
		idx, ok := p.ident()
		if !ok || !p.take("ON") {
			return eff, badRule("R041")
		}
		sc, t, ok := p.qname()
		if !ok || !p.exact("(") {
			return eff, badRule("R041")
		}
		cols := []string{}
		seen := map[string]bool{}
		for {
			c, ok := p.ident()
			if !ok || seen[c] || len(cols) >= 32 {
				return eff, badRule("R041")
			}
			seen[c] = true
			cols = append(cols, c)
			if p.exact(")") {
				break
			}
			if !p.exact(",") {
				return eff, badRule("R041")
			}
		}
		if !p.end() {
			return eff, badRule("R041")
		}
		table := sc + "." + t
		if known, ok := st.created[table]; ok {
			for _, c := range cols {
				if _, ok := known[c]; !ok {
					return eff, badRule("R041")
				}
			}
		} else if !con {
			return eff, verr("MIG013_SQL_SAFETY", "R029", "pre-existing-table index must be CONCURRENTLY")
		}
		eff.schema = sc
		eff.concurrent = con
		eff.minRisk = 2
		eff.key = "index:" + sc + "." + idx
		if unique {
			eff.class = "create_unique_index"
		} else {
			eff.class = "create_index"
		}
		if con {
			eff.class += "_concurrently"
		}
		return eff, nil
	}
	if p.take("ALTER") && p.take("TABLE") {
		sc, t, ok := p.qname()
		if !ok || !p.take("ADD") {
			return eff, badRule("R033")
		}
		table := sc + "." + t
		eff.schema = sc
		if p.take("COLUMN") {
			n, ty, def, rule, ok := parseCol(p)
			if !ok {
				return eff, badRule(rule)
			}
			if !p.end() {
				return eff, badRule("R048")
			}
			key := table + "." + n
			if _, dup := st.added[key]; dup {
				return eff, badRule("R048")
			}
			if known := st.created[table]; known != nil {
				if _, dup := known[n]; dup {
					return eff, badRule("R048")
				}
				known[n] = ty
			}
			st.added[key] = ty
			eff.class = "add_column"
			eff.key = "column:" + table + "." + n
			if def {
				eff.minRisk = 2
			}
			return eff, nil
		}
		if !p.take("CONSTRAINT") || st.created[table] != nil {
			return eff, badRule("R033")
		}
		cn, ok := p.ident()
		if !ok {
			return eff, badRule("R033")
		}
		eff.minRisk = 2
		eff.key = "constraint:" + table + "." + cn
		if p.take("CHECK") {
			if !p.exact("(") {
				return eff, badRule("R052")
			}
			if p.i < len(p.t) && p.t[p.i].text == "(" {
				return eff, badRule("R052")
			}
			col, ok := p.ident()
			if !ok {
				return eff, badRule("R040")
			}
			ty, ok := st.added[table+"."+col]
			if !ok {
				return eff, badRule("R040")
			}
			if p.take("IS") {
				p.take("NOT")
				if !p.take("NULL") {
					return eff, badRule("R040")
				}
			} else {
				if p.i >= len(p.t) || !in(makeSet([]string{"=", "<>", "<", "<=", ">", ">="}), p.t[p.i].text) {
					return eff, badRule("R040")
				}
				op := p.t[p.i].text
				p.i++
				if (ty.kind == "boolean" || ty.kind == "text" || ty.kind == "varchar") && op != "=" && op != "<>" {
					return eff, badRule("R040")
				}
				if ty.kind == "uuid" || ty.kind == "date" || ty.kind == "timestamp" || p.i >= len(p.t) || !validLiteral(p.t[p.i], ty) {
					return eff, badRule("R040")
				}
				p.i++
			}
			if p.peek("AND") || p.peek("OR") {
				return eff, badRule("R040")
			}
			if !p.exact(")") || !p.take("NOT") || !p.take("VALID") || !p.end() {
				return eff, badRule("R052")
			}
			eff.class = "add_check_constraint"
			return eff, nil
		}
		if p.take("FOREIGN") {
			if !p.take("KEY") || !p.exact("(") {
				return eff, badRule("R033")
			}
			local, ok := idList(p, 32)
			if !ok || !p.exact(")") || !p.take("REFERENCES") {
				return eff, badRule("R033")
			}
			_, _, ok = p.qname()
			if !ok || !p.exact("(") {
				return eff, badRule("R033")
			}
			ref, ok := idList(p, 32)
			if !ok || !p.exact(")") || len(local) != len(ref) || !p.take("NOT") || !p.take("VALID") || !p.end() {
				return eff, badRule("R033")
			}
			eff.class = "add_foreign_key"
			return eff, nil
		}
		return eff, badRule("R033")
	}
	return eff, badDDL()
}
func idList(p *parser, max int) ([]string, bool) {
	var a []string
	seen := map[string]bool{}
	for {
		n, ok := p.ident()
		if !ok || seen[n] || len(a) >= max {
			return nil, false
		}
		seen[n] = true
		a = append(a, n)
		if !p.exact(",") {
			return a, true
		}
	}
}
func badRule(rule string) error {
	return verr("MIG012_STATEMENT_CLASS", rule, "statement outside paired-SQL v1 grammar")
}
func badDDL() error { return badRule("R008") }

func parseDownDDL(s statement) (effect, error) {
	p := &parser{t: s.tokens}
	e := effect{stmtSHA: hashBytes(s.raw), minRisk: 1}
	if p.take("DROP") {
		if p.take("TABLE") {
			sc, t, ok := p.qname()
			if ok && p.end() {
				e.class = "drop_table"
				e.schema = sc
				e.key = "table:" + sc + "." + t
				return e, nil
			}
		}
		if p.take("INDEX") {
			con := p.take("CONCURRENTLY")
			sc, i, ok := p.qname()
			if ok && p.end() {
				e.schema = sc
				e.concurrent = con
				e.key = "index:" + sc + "." + i
				e.class = "drop_index"
				if con {
					e.class += "_concurrently"
				}
				return e, nil
			}
		}
	}
	if p.take("ALTER") && p.take("TABLE") {
		sc, t, ok := p.qname()
		if !ok || !p.take("DROP") {
			return e, badDown()
		}
		if p.take("COLUMN") {
			n, ok := p.ident()
			if ok && p.end() {
				e.class = "drop_column"
				e.schema = sc
				e.key = "column:" + sc + "." + t + "." + n
				return e, nil
			}
		}
		if p.take("CONSTRAINT") {
			n, ok := p.ident()
			if ok && p.end() {
				e.class = "drop_constraint"
				e.schema = sc
				e.key = "constraint:" + sc + "." + t + "." + n
				return e, nil
			}
		}
	}
	return e, badDown()
}
func badDown() error {
	return verr("MIG022_DOWN_INVERSE", "R053", "DOWN outside exact inverse grammar")
}
