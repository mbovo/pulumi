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

type AnonSymbolExpression struct {
	Syntax *hclsyntax.AnonSymbolExpr

	exprType Type
}

func (x *AnonSymbolExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *AnonSymbolExpression) Type() Type {
	return x.exprType
}

func (*AnonSymbolExpression) isExpression() {}

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
}

func (x *ConditionalExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *ConditionalExpression) Type() Type {
	return BoolType
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

	Args []Expression

	exprType Type
}

func (x *FunctionCallExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *FunctionCallExpression) Type() Type {
	return x.exprType
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

	exprType Type
}

func (x *RelativeTraversalExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *RelativeTraversalExpression) Type() Type {
	return x.exprType
}

func (*RelativeTraversalExpression) isExpression() {}

type ScopeTraversalExpression struct {
	Syntax *hclsyntax.ScopeTraversalExpr

	exprType Type
}

func (x *ScopeTraversalExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *ScopeTraversalExpression) Type() Type {
	return x.exprType
}

func (*ScopeTraversalExpression) isExpression() {}

type SplatExpression struct {
	Syntax *hclsyntax.SplatExpr

	Source Expression
	Each   Expression
	Item   *AnonSymbolExpression

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
}

func (x *TemplateExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *TemplateExpression) Type() Type {
	return StringType
}

func (*TemplateExpression) isExpression() {}

type TemplateJoinExpression struct {
	Syntax *hclsyntax.TemplateJoinExpr

	Tuple Expression
}

func (x *TemplateJoinExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *TemplateJoinExpression) Type() Type {
	return StringType
}

func (*TemplateJoinExpression) isExpression() {}

type TemplateWrapExpression struct {
	Syntax *hclsyntax.TemplateWrapExpr

	Wrapped Expression

	exprType Type
}

func (x *TemplateWrapExpression) SyntaxNode() hclsyntax.Node {
	return x.Syntax
}

func (x *TemplateWrapExpression) Type() Type {
	return x.exprType
}

func (*TemplateWrapExpression) isExpression() {}

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
