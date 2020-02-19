package model

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/pulumi/pulumi/pkg/resource"
	"github.com/pulumi/pulumi/pkg/util/contract"
	"github.com/zclconf/go-cty/cty"
)

func (b *binder) bindExpression(syntax hclsyntax.Node) (Expression, hcl.Diagnostics) {
	switch syntax := syntax.(type) {
	case *hclsyntax.AnonSymbolExpr:
		return b.bindAnonSymbolExpression(syntax)
	case *hclsyntax.BinaryOpExpr:
		return b.bindBinaryOpExpression(syntax)
	case *hclsyntax.ConditionalExpr:
		return b.bindConditionalExpression(syntax)
	case *hclsyntax.ForExpr:
		return b.bindForExpression(syntax)
	case *hclsyntax.FunctionCallExpr:
		return b.bindFunctionCallExpression(syntax)
	case *hclsyntax.IndexExpr:
		return b.bindIndexExpression(syntax)
	case *hclsyntax.LiteralValueExpr:
		return b.bindLiteralValueExpression(syntax)
	case *hclsyntax.ObjectConsExpr:
		return b.bindObjectConsExpression(syntax)
	case *hclsyntax.ObjectConsKeyExpr:
		return b.bindObjectConsKeyExpr(syntax)
	case *hclsyntax.RelativeTraversalExpr:
		return b.bindRelativeTraversalExpression(syntax)
	case *hclsyntax.ScopeTraversalExpr:
		return b.bindScopeTraversalExpression(syntax)
	case *hclsyntax.SplatExpr:
		return b.bindSplatExpression(syntax)
	case *hclsyntax.TemplateExpr:
		return b.bindTemplateExpression(syntax)
	case *hclsyntax.TemplateJoinExpr:
		return b.bindTemplateJoinExpression(syntax)
	case *hclsyntax.TemplateWrapExpr:
		return b.bindTemplateWrapExpression(syntax)
	case *hclsyntax.TupleConsExpr:
		return b.bindTupleConsExpression(syntax)
	case *hclsyntax.UnaryOpExpr:
		return b.bindUnaryOpExpression(syntax)
	default:
		contract.Failf("unexpected expression node of type %T (%v)", syntax, syntax.Range())
		return nil, nil
	}
}

func (b *binder) bindAnonSymbolExpression(syntax *hclsyntax.AnonSymbolExpr) (Expression, hcl.Diagnostics) {
	return &ErrorExpression{Syntax: syntax, exprType: AnyType}, notYetImplemented(syntax)
}

func (b *binder) bindBinaryOpExpression(syntax *hclsyntax.BinaryOpExpr) (Expression, hcl.Diagnostics) {
	return &ErrorExpression{Syntax: syntax, exprType: AnyType}, notYetImplemented(syntax)
}

func (b *binder) bindConditionalExpression(syntax *hclsyntax.ConditionalExpr) (Expression, hcl.Diagnostics) {
	return &ErrorExpression{Syntax: syntax, exprType: AnyType}, notYetImplemented(syntax)
}

func (b *binder) bindForExpression(syntax *hclsyntax.ForExpr) (Expression, hcl.Diagnostics) {
	return &ErrorExpression{Syntax: syntax, exprType: AnyType}, notYetImplemented(syntax)
}

func (b *binder) bindFunctionCallExpression(syntax *hclsyntax.FunctionCallExpr) (Expression, hcl.Diagnostics) {
	definition, diagnostics := getFunctionDefinition(syntax.Name, syntax.NameRange)

	args := make([]Expression, len(syntax.Args))
	for i, syntax := range syntax.Args {
		arg, argDiagnostics := b.bindExpression(syntax)
		args[i], diagnostics = arg, append(diagnostics, argDiagnostics...)
	}

	if definition == nil {
		return &FunctionCallExpression{
			Syntax:   syntax,
			Args:     args,
			exprType: AnyType,
		}, diagnostics
	}

	signature, sigDiags := definition.signature(args)
	diagnostics = append(diagnostics, sigDiags...)

	remainingArgs := args
	for _, param := range signature.parameters {
		if len(remainingArgs) == 0 {
			if !IsOptionalType(param.typ) {
				diagnostics = append(diagnostics, missingRequiredArgument(param, syntax.Range()))
			}
		} else {
			if !param.typ.AssignableFrom(remainingArgs[0].Type()) {
				diagnostics = append(diagnostics, exprNotAssignable(param.typ, remainingArgs[0]))
			}
			remainingArgs = remainingArgs[1:]
		}
	}

	if len(remainingArgs) > 0 {
		varargs := signature.varargsParameter
		if varargs == nil {
			diagnostics = append(diagnostics, extraArguments(len(signature.parameters), len(args), syntax.Range()))
		} else {
			for _, arg := range remainingArgs {
				if !varargs.typ.AssignableFrom(arg.Type()) {
					diagnostics = append(diagnostics, exprNotAssignable(varargs.typ, arg))
				}
			}
		}
	}

	return &FunctionCallExpression{
		Syntax:   syntax,
		Args:     args,
		exprType: signature.returnType,
	}, diagnostics
}

func (b *binder) bindIndexExpression(syntax *hclsyntax.IndexExpr) (Expression, hcl.Diagnostics) {
	return &ErrorExpression{Syntax: syntax, exprType: AnyType}, notYetImplemented(syntax)
}

func (b *binder) bindLiteralValueExpression(syntax *hclsyntax.LiteralValueExpr) (Expression, hcl.Diagnostics) {
	pv, typ, diagnostics := resource.PropertyValue{}, Type(nil), hcl.Diagnostics(nil)

	v := syntax.Val
	switch {
	case v.IsNull():
		// OK
	case v.Type() == cty.Bool:
		pv, typ = resource.NewBoolProperty(v.True()), BoolType
	case v.Type() == cty.Number:
		f, _ := v.AsBigFloat().Float64()
		pv, typ = resource.NewNumberProperty(f), NumberType
	case v.Type() == cty.String:
		pv, typ = resource.NewStringProperty(v.AsString()), StringType
	default:
		typ, diagnostics = AnyType, hcl.Diagnostics{unsupportedLiteralValue(syntax)}
	}

	return &LiteralValueExpression{
		Syntax:   syntax,
		Value:    pv,
		exprType: typ,
	}, diagnostics
}

func (b *binder) bindObjectConsExpression(syntax *hclsyntax.ObjectConsExpr) (Expression, hcl.Diagnostics) {
	var diagnostics hcl.Diagnostics

	items := make([]ObjectConsItem, len(syntax.Items))
	for i, item := range syntax.Items {
		keyExpr, keyDiags := b.bindExpression(item.KeyExpr)
		diagnostics = append(diagnostics, keyDiags...)

		if !StringType.AssignableFrom(keyExpr.Type()) {
			// TODO(pdg): this does not match the default HCL2 evaluation semantics.
			diagnostics = append(diagnostics, objectKeysMustBeStrings(keyExpr))
		}

		valExpr, valDiags := b.bindExpression(item.ValueExpr)
		diagnostics = append(diagnostics, valDiags...)

		items[i] = ObjectConsItem{Key: keyExpr, Value: valExpr}
	}

	// Attempt to build an object type out of the result. If there are any attribute names that come from variables,
	// type the result as Any.
	//
	// TODO)pdg): can we refine this?
	properties, isAnyType, typ := map[string]Type{}, false, Type(nil)
	for _, item := range items {
		keyLit, ok := item.Key.(*LiteralValueExpression)
		if !ok || !keyLit.Value.IsString() {
			isAnyType, typ = true, AnyType
			break
		}
		properties[keyLit.Value.StringValue()] = item.Value.Type()
	}
	if !isAnyType {
		typ = NewObjectType(properties)
	}

	return &ObjectConsExpression{
		Syntax:   syntax,
		Items:    items,
		exprType: typ,
	}, diagnostics
}

func (b *binder) bindObjectConsKeyExpr(syntax *hclsyntax.ObjectConsKeyExpr) (Expression, hcl.Diagnostics) {
	if !syntax.ForceNonLiteral {
		if name := hcl.ExprAsKeyword(syntax); name != "" {
			return b.bindExpression(&hclsyntax.LiteralValueExpr{
				Val:      cty.StringVal(name),
				SrcRange: syntax.Range(),
			})
		}
	}
	return b.bindExpression(syntax.Wrapped)
}

func (b *binder) bindRelativeTraversalExpression(syntax *hclsyntax.RelativeTraversalExpr) (Expression, hcl.Diagnostics) {
	source, diagnostics := b.bindExpression(syntax.Source)

	typ, typDiags := b.bindTraversalType(source.Type(), syntax.Traversal)
	diagnostics = append(diagnostics, typDiags...)

	return &RelativeTraversalExpression{
		Syntax:   syntax,
		Source:   source,
		exprType: typ,
	}, diagnostics
}

func (b *binder) bindScopeTraversalExpression(syntax *hclsyntax.ScopeTraversalExpr) (Expression, hcl.Diagnostics) {
	// TODO(pdg): count, range, for, etc.

	node, ok := b.nodes[syntax.Traversal.RootName()]
	if !ok {
		return &ScopeTraversalExpression{
			Syntax:   syntax,
			exprType: AnyType,
		}, hcl.Diagnostics{undefinedVariable(syntax.Traversal.SimpleSplit().Abs.SourceRange())}
	}

	typ, diagnostics := b.bindTraversalType(node.Type(), syntax.Traversal.SimpleSplit().Rel)
	return &ScopeTraversalExpression{
		Syntax:   syntax,
		exprType: typ,
	}, diagnostics
}

func (b *binder) bindSplatExpression(syntax *hclsyntax.SplatExpr) (Expression, hcl.Diagnostics) {
	return &ErrorExpression{Syntax: syntax, exprType: AnyType}, notYetImplemented(syntax)
}

func (b *binder) bindTemplateExpression(syntax *hclsyntax.TemplateExpr) (Expression, hcl.Diagnostics) {
	if syntax.IsStringLiteral() {
		return b.bindExpression(syntax.Parts[0])
	}

	var diagnostics hcl.Diagnostics
	parts := make([]Expression, len(syntax.Parts))
	for i, syntax := range syntax.Parts {
		part, partDiags := b.bindExpression(syntax)
		parts[i], diagnostics = part, append(diagnostics, partDiags...)
	}

	return &TemplateExpression{
		Syntax: syntax,
		Parts:  parts,
	}, diagnostics

	return &ErrorExpression{Syntax: syntax, exprType: AnyType}, notYetImplemented(syntax)
}

func (b *binder) bindTemplateJoinExpression(syntax *hclsyntax.TemplateJoinExpr) (Expression, hcl.Diagnostics) {
	tuple, diagnostics := b.bindExpression(syntax.Tuple)
	return &TemplateJoinExpression{
		Syntax: syntax,
		Tuple:  tuple,
	}, diagnostics
}

func (b *binder) bindTemplateWrapExpression(syntax *hclsyntax.TemplateWrapExpr) (Expression, hcl.Diagnostics) {
	return b.bindExpression(syntax.Wrapped)
}

func (b *binder) bindTupleConsExpression(syntax *hclsyntax.TupleConsExpr) (Expression, hcl.Diagnostics) {
	var diagnostics hcl.Diagnostics
	exprs := make([]Expression, len(syntax.Exprs))
	for i, syntax := range syntax.Exprs {
		expr, exprDiags := b.bindExpression(syntax)
		exprs[i], diagnostics = expr, append(diagnostics, exprDiags...)
	}

	// TODO(pdg): better typing. Need an algorithm for finding the best type.
	var typ Type
	for _, expr := range exprs {
		if typ == nil {
			typ = expr.Type()
		} else if expr.Type() != typ {
			typ = AnyType
			break
		}
	}

	return &TupleConsExpression{
		Syntax:      syntax,
		Expressions: exprs,
		exprType:    typ,
	}, diagnostics
}

func (b *binder) bindUnaryOpExpression(syntax *hclsyntax.UnaryOpExpr) (Expression, hcl.Diagnostics) {
	return &ErrorExpression{Syntax: syntax, exprType: AnyType}, notYetImplemented(syntax)
}
