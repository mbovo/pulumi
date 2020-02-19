package model

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Resource struct {
	Syntax *hclsyntax.Block

	InputType  Type
	OutputType Type

	Inputs Expression
	Range  Expression

	state bindState

	// TODO: Resource options
}

func (r *Resource) SyntaxNode() hclsyntax.Node {
	return r.Syntax
}

func (r *Resource) Type() Type {
	return r.OutputType
}

func (r *Resource) getState() bindState {
	return r.state
}

func (r *Resource) setState(s bindState) {
	r.state = s
}

func (*Resource) isNode() {}

// bind from syntax + schema
