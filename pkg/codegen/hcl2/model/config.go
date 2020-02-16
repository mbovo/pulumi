package model

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type ConfigVariable struct {
	Syntax *hclsyntax.Block

	Type         Type
	DefaultValue Expression

	state bindState
}

func (cv *ConfigVariable) SyntaxNode() hclsyntax.Node {
	return cv.Syntax
}

func (cv *ConfigVariable) getState() bindState {
	return cv.state
}

func (cv *ConfigVariable) setState(s bindState) {
	cv.state = s
}

func (*ConfigVariable) isNode() {}
