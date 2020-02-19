package model

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/pulumi/pulumi/pkg/codegen/schema"
	"github.com/zclconf/go-cty/cty"
)

func getResourceToken(node *Resource) (string, hcl.Range) {
	return node.Syntax.Labels[1], node.Syntax.LabelRanges[1]
}

func (b *binder) bindResource(node *Resource) hcl.Diagnostics {
	return b.bindResourceTypes(node)
}

// bindResourceTypes binds the input and output types for a resource.
func (b *binder) bindResourceTypes(node *Resource) hcl.Diagnostics {
	// Set the input and output types to Any by default.
	node.InputType, node.OutputType = AnyType, AnyType

	// Find the resource's schema.
	token, tokenRange := getResourceToken(node)
	pkg, _, _, diagnostics := decomposeToken(token, tokenRange)
	if diagnostics.HasErrors() {
		return diagnostics
	}

	pkgSchema, ok := b.packageSchemas[pkg]
	if !ok {
		return hcl.Diagnostics{unknownPackage(pkg, tokenRange)}
	}
	res, ok := pkgSchema.resources[token]
	if !ok {
		return hcl.Diagnostics{unknownResourceType(token, tokenRange)}
	}

	// Create input and output types for the schema.
	inputType := schemaTypeToType(&schema.ObjectType{Properties: res.InputProperties}, true)

	outputProperties := map[string]Type{
		"id":  NewOutputType(StringType),
		"urn": NewOutputType(StringType),
	}
	for _, prop := range res.Properties {
		outputProperties[prop.Name] = NewOutputType(schemaTypeToType(prop.Type, false))
	}
	outputType := NewObjectType(outputProperties)

	node.InputType, node.OutputType = inputType, outputType

	// Bind the resource's body.
	bodyItems := make([]hclsyntax.ObjectConsItem, len(node.Syntax.Body.Attributes))
	for i, attr := range sourceOrderAttributes(node.Syntax.Body.Attributes) {
		bodyItems[i] = hclsyntax.ObjectConsItem{
			KeyExpr:   &hclsyntax.LiteralValueExpr{Val: cty.StringVal(attr.Name), SrcRange: attr.NameRange},
			ValueExpr: attr.Expr,
		}
	}
	bodyObject := &hclsyntax.ObjectConsExpr{
		Items:     bodyItems,
		SrcRange:  hcl.RangeBetween(node.Syntax.OpenBraceRange, node.Syntax.CloseBraceRange),
		OpenRange: node.Syntax.OpenBraceRange,
	}
	bodyExpr, bodyDiags := b.bindExpression(bodyObject)
	diagnostics = append(diagnostics, bodyDiags...)

	// TODO(pdg): return diagnostics from AssignableFrom
	if !node.InputType.AssignableFrom(bodyExpr.Type()) {
		diagnostics = append(diagnostics, exprNotAssignable(node.InputType, bodyExpr))
	}

	// TODO(pdg): resource options

	return diagnostics
}
