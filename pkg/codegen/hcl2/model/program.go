package model

import (
	"io"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/pulumi/pulumi/pkg/codegen/hcl2/syntax"
)

type bindState int

const (
	unbound = 0
	binding = 1
	bound   = 2
)

type Node interface {
	SyntaxNode() hclsyntax.Node
	Type() Type

	getState() bindState
	setState(s bindState)

	isNode()
}

type Program struct {
	Nodes []Node

	files []*syntax.File
}

func (p *Program) NewDiagnosticWriter(w io.Writer, width uint, color bool) hcl.DiagnosticWriter {
	return syntax.NewDiagnosticWriter(w, p.files, width, color)
}
