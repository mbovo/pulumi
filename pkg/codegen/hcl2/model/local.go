package model

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type LocalVariable struct {
	Syntax *hclsyntax.Attribute

	Name         string
	VariableType Type
	Value        Expression

	state bindState
	deps  []Node
}

func (lv *LocalVariable) SyntaxNode() hclsyntax.Node {
	return lv.Syntax
}

func (lv *LocalVariable) Type() Type {
	return lv.VariableType
}

func (lv *LocalVariable) getState() bindState {
	return lv.state
}

func (lv *LocalVariable) setState(s bindState) {
	lv.state = s
}

func (lv *LocalVariable) getDependencies() []Node {
	return lv.deps
}

func (lv *LocalVariable) setDependencies(nodes []Node) {
	lv.deps = nodes
}

func (*LocalVariable) isNode() {}
