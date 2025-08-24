package config

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenIdentifier TokenType = iota
	TokenNumber
	TokenString
	TokenSymbol
	TokenEOF
)

type Token struct {
	Type  TokenType
	Value string
	Line  int
}

type Lexer struct {
	input string
	pos   int
	line  int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input, pos: 0, line: 1}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()
	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF, Line: l.line}
	}

	ch := l.input[l.pos]

	// Symbols
	if strings.ContainsRune("{};", rune(ch)) {
		l.pos++
		return Token{Type: TokenSymbol, Value: string(ch), Line: l.line}
	}

	// Numbers
	if unicode.IsDigit(rune(ch)) {
		start := l.pos
		for l.pos < len(l.input) && unicode.IsDigit(rune(l.input[l.pos])) {
			l.pos++
		}
		return Token{Type: TokenNumber, Value: l.input[start:l.pos], Line: l.line}
	}

	// Identifiers / strings
	start := l.pos
	for l.pos < len(l.input) && !unicode.IsSpace(rune(l.input[l.pos])) && !strings.ContainsRune("{};", rune(l.input[l.pos])) {
		l.pos++
	}
	return Token{Type: TokenIdentifier, Value: l.input[start:l.pos], Line: l.line}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '\n' {
			l.line++
		}
		if !unicode.IsSpace(rune(ch)) {
			break
		}
		l.pos++
	}
}

type Parser struct {
	tokens []Token
	pos    int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) consume() Token {
	tok := p.peek()
	p.pos++
	return tok
}

func (p *Parser) expectSymbol(val string) {
	tok := p.consume()
	if tok.Type != TokenSymbol || tok.Value != val {
		panic(fmt.Sprintf("expected symbol %s at line %d, got %s", val, tok.Line, tok.Value))
	}
}

func (p *Parser) ParseConfig() Config {
	cfg := Config{}

	// Expect "servers" identifier
	serversTok := p.consume()
	if serversTok.Type != TokenIdentifier || serversTok.Value != "servers" {
		panic(fmt.Sprintf("expected 'servers' at line %d, got %s", serversTok.Line, serversTok.Value))
	}

	p.expectSymbol("{")

	// Parse server blocks
	for p.peek().Type != TokenSymbol || p.peek().Value != "}" {
		server := p.parseServer()
		cfg.Servers = append(cfg.Servers, server)
	}

	p.expectSymbol("}")
	return cfg
}

func (p *Parser) parseServer() Server {
	server := Server{}

	// Expect "server" identifier
	serverTok := p.consume()
	if serverTok.Type != TokenIdentifier || serverTok.Value != "server" {
		panic(fmt.Sprintf("expected 'server' at line %d, got %s", serverTok.Line, serverTok.Value))
	}

	p.expectSymbol("{")

	for p.peek().Type != TokenSymbol || p.peek().Value != "}" {
		tok := p.peek()
		if tok.Type == TokenIdentifier && tok.Value == "location" {
			loc := p.parseLocation()
			server.Locations = append(server.Locations, loc)
		} else {
			key, value := p.parseDirective()
			p.applyDirective(&server, key, value)
		}
	}

	p.expectSymbol("}")
	return server
}

func (p *Parser) parseLocation() Location {
	loc := Location{}
	p.consume() // consume 'location'
	pathTok := p.consume()
	if pathTok.Type != TokenIdentifier {
		panic(fmt.Sprintf("expected location path at line %d", pathTok.Line))
	}
	loc.Path = pathTok.Value
	p.expectSymbol("{")

	for p.peek().Type != TokenSymbol || p.peek().Value != "}" {
		key, value := p.parseDirective()
		switch key {
		case "root":
			loc.Root = value
		case "proxy_pass":
			loc.ProxyPass = value
		default:
			panic(fmt.Sprintf("unknown location directive %s at line %d", key, p.peek().Line))
		}
	}

	p.expectSymbol("}")
	return loc
}

func (p *Parser) parseDirective() (string, string) {
	keyTok := p.consume()
	if keyTok.Type != TokenIdentifier {
		panic(fmt.Sprintf("expected directive at line %d", keyTok.Line))
	}

	if p.peek().Type == TokenSymbol && p.peek().Value == ";" {
		// directive without value
		p.consume()
		return keyTok.Value, ""
	}

	valTok := p.consume()
	if valTok.Type != TokenIdentifier && valTok.Type != TokenNumber && valTok.Type != TokenString {
		panic(fmt.Sprintf("expected directive value at line %d", valTok.Line))
	}

	// optionally consume trailing semicolon
	if p.peek().Type == TokenSymbol && p.peek().Value == ";" {
		p.consume()
	}

	return keyTok.Value, valTok.Value
}

func (p *Parser) applyDirective(s *Server, key, value string) {
	switch key {
	case "name":
		s.Name = value
	case "listen":
		port, _ := strconv.Atoi(value)
		if s.Listen.Port == 0 {
			s.Listen = Listen{Port: port, SSL: false}
		} else {
			s.Listen.Port = port
		}
	case "ssl":
		s.Listen.SSL = value == "true" || value == "yes"
	case "hosts":
		s.Hosts = strings.Split(value, ",")
	case "access_log":
		s.AccessLog = value
	case "ssl_certificate":
		if s.SSL == nil {
			s.SSL = &SSLConfig{}
		}
		s.SSL.Certificate = value
	case "ssl_certificate_key":
		if s.SSL == nil {
			s.SSL = &SSLConfig{}
		}
		s.SSL.CertificateKey = value
	default:
		panic(fmt.Sprintf("unknown server directive %s", key))
	}
}
