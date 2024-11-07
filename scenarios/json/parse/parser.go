package scenjsonparse

import (
	ei "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/expression/interpreter"
	fr "github.com/kalyan3104/k-chain-vm-v1_3-go/scenarios/fileresolver"
)

// Parser performs parsing of both json tests (older) and scenarios (new).
type Parser struct {
	ExprInterpreter ei.ExprInterpreter
}

// NewParser provides a new Parser instance.
func NewParser(fileResolver fr.FileResolver) Parser {
	return Parser{
		ExprInterpreter: ei.ExprInterpreter{
			FileResolver: fileResolver,
		},
	}
}
