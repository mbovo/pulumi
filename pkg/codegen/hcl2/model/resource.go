package model

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Resource struct {
	Syntax *hclsyntax.Block

	InputType  Type
	OutputType Type

	Inputs Expression
	Range  Expression

	state bindState
	deps  []Node

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

func (r *Resource) getDependencies() []Node {
	return r.deps
}

func (r *Resource) setDependencies(nodes []Node) {
	r.deps = nodes
}

func (*Resource) isNode() {}

func (r *Resource) Name() string {
	return r.Syntax.Labels[0]
}

func (r *Resource) DecomposeToken() (string, string, string, hcl.Diagnostics) {
	token, tokenRange := getResourceToken(r)
	return decomposeToken(token, tokenRange)
}
