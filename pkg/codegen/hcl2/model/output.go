package model

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type OutputVariable struct {
	Syntax *hclsyntax.Block

	Type  Type
	Value Expression

	state bindState
}

func (ov *OutputVariable) SyntaxNode() hclsyntax.Node {
	return ov.Syntax
}

func (ov *OutputVariable) getState() bindState {
	return ov.state
}

func (ov *OutputVariable) setState(s bindState) {
	ov.state = s
}

func (*OutputVariable) isNode() {}
