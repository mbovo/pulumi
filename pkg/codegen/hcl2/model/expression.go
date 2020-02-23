package model

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/pulumi/pulumi/pkg/resource"
)

type Expression interface {
	SyntaxNode() hclsyntax.Node
	Type() Type

	isExpression()
}

type AnonymousFunctionExpression struct {
	Signature  FunctionSignature
	Parameters []*LocalVariable

	Body Expression
}

func (x *AnonymousFunctionExpression) SyntaxNode() hclsyntax.Node {
	return x.Body.SyntaxNode()
}

func (x *AnonymousFunctionExpression) Type() Type {
	// TODO: function types
	return AnyType
}

func (*AnonymousFunctionExpression) isExpression() {}

type BinaryOpExpression struct {
	Syntax *hclsyntax.BinaryOpExpr

	LeftOperand  Expression
	RightOperand Expression

	exprType Type
}

func (x *BinaryOpExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *BinaryOpExpression) Type() Type {
	return x.exprType
}

func (*BinaryOpExpression) isExpression() {}

type ConditionalExpression struct {
	Syntax *hclsyntax.ConditionalExpr

	Condition   Expression
	TrueResult  Expression
	FalseResult Expression

	exprType Type
}

func (x *ConditionalExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *ConditionalExpression) Type() Type {
	return x.exprType
}

func (*ConditionalExpression) isExpression() {}

type ErrorExpression struct {
	Syntax hclsyntax.Node

	exprType Type
}

func (x *ErrorExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *ErrorExpression) Type() Type {
	return x.exprType
}

func (*ErrorExpression) isExpression() {}

type ForExpression struct {
	Syntax *hclsyntax.ForExpr

	Collection Expression
	Key        Expression
	Value      Expression
	Condition  Expression

	exprType Type
}

func (x *ForExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *ForExpression) Type() Type {
	return x.exprType
}

func (*ForExpression) isExpression() {}

type FunctionCallExpression struct {
	Syntax *hclsyntax.FunctionCallExpr

	Name      string
	Signature FunctionSignature
	Args      []Expression
}

func (x *FunctionCallExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *FunctionCallExpression) Type() Type {
	return x.Signature.ReturnType
}

func (*FunctionCallExpression) isExpression() {}

type IndexExpression struct {
	Syntax *hclsyntax.IndexExpr

	Collection Expression
	Key        Expression

	exprType Type
}

func (x *IndexExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *IndexExpression) Type() Type {
	return x.exprType
}

func (*IndexExpression) isExpression() {}

type LiteralValueExpression struct {
	Syntax *hclsyntax.LiteralValueExpr

	Value resource.PropertyValue

	exprType Type
}

func (x *LiteralValueExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *LiteralValueExpression) Type() Type {
	return x.exprType
}

func (*LiteralValueExpression) isExpression() {}

type ObjectConsExpression struct {
	Syntax *hclsyntax.ObjectConsExpr

	Items []ObjectConsItem

	exprType Type
}

func (x *ObjectConsExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *ObjectConsExpression) Type() Type {
	return x.exprType
}

type ObjectConsItem struct {
	Key   Expression
	Value Expression
}

func (*ObjectConsExpression) isExpression() {}

type RelativeTraversalExpression struct {
	Syntax *hclsyntax.RelativeTraversalExpr

	Source Expression
	Types  []Type
}

func (x *RelativeTraversalExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *RelativeTraversalExpression) Type() Type {
	return x.Types[len(x.Types)-1]
}

func (*RelativeTraversalExpression) isExpression() {}

type ScopeTraversalExpression struct {
	Syntax *hclsyntax.ScopeTraversalExpr

	Node  Node
	Types []Type
}

func (x *ScopeTraversalExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *ScopeTraversalExpression) Type() Type {
	return x.Types[len(x.Types)-1]
}

func (*ScopeTraversalExpression) isExpression() {}

type SplatExpression struct {
	Syntax *hclsyntax.SplatExpr

	Source Expression
	Each   Expression
	Item   *LocalVariable

	exprType Type
}

func (x *SplatExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *SplatExpression) Type() Type {
	return x.exprType
}

func (*SplatExpression) isExpression() {}

type TemplateExpression struct {
	Syntax *hclsyntax.TemplateExpr

	Parts []Expression

	exprType Type
}

func (x *TemplateExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *TemplateExpression) Type() Type {
	return x.exprType
}

func (*TemplateExpression) isExpression() {}

type TemplateJoinExpression struct {
	Syntax *hclsyntax.TemplateJoinExpr

	Tuple Expression

	exprType Type
}

func (x *TemplateJoinExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *TemplateJoinExpression) Type() Type {
	return x.exprType
}

func (*TemplateJoinExpression) isExpression() {}

type TupleConsExpression struct {
	Syntax *hclsyntax.TupleConsExpr

	Expressions []Expression

	exprType Type
}

func (x *TupleConsExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *TupleConsExpression) Type() Type {
	return x.exprType
}

func (*TupleConsExpression) isExpression() {}

type UnaryOpExpression struct {
	Syntax *hclsyntax.UnaryOpExpr

	Operand Expression

	exprType Type
}

func (x *UnaryOpExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *UnaryOpExpression) Type() Type {
	return x.exprType
}

func (*UnaryOpExpression) isExpression() {}
