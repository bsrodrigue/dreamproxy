package config

import (
	"os"
)

func LoadDreamFile(config_file_path string) Config {
	config_bin, err := os.ReadFile(config_file_path)

	if err != nil {
		panic(err)
	}

	lexer := NewLexer(string(config_bin))

	var tokens []Token

	for {
		token := lexer.NextToken()
		tokens = append(tokens, token)

		if token.Type == TokenEOF {
			break
		}
	}

	parser := NewParser(tokens)

	cfg := parser.ParseConfig()

	return cfg
}
