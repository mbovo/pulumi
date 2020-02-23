package model

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type OutputVariable struct {
	Syntax *hclsyntax.Block

	typ   Type
	Value Expression

	state bindState
	deps  []Node
}

func (ov *OutputVariable) SyntaxNode() hclsyntax.Node {
	return ov.Syntax
}

func (ov *OutputVariable) Type() Type {
	return ov.typ
}

func (ov *OutputVariable) getState() bindState {
	return ov.state
}

func (ov *OutputVariable) setState(s bindState) {
	ov.state = s
}

func (ov *OutputVariable) getDependencies() []Node {
	return ov.deps
}

func (ov *OutputVariable) setDependencies(nodes []Node) {
	ov.deps = nodes
}

func (*OutputVariable) isNode() {}
