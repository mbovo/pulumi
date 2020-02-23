// Copyright 2016-2020, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	getDependencies() []Node
	setDependencies(nodes []Node)

	isNode()
}

type Program struct {
	Nodes []Node

	files []*syntax.File

	binder *binder
}

func (p *Program) NewDiagnosticWriter(w io.Writer, width uint, color bool) hcl.DiagnosticWriter {
	return syntax.NewDiagnosticWriter(w, p.files, width, color)
}

func (p *Program) BindExpression(node hclsyntax.Node) (Expression, hcl.Diagnostics) {
	return p.binder.bindExpression(node)
}
