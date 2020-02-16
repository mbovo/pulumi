package model

import "github.com/hashicorp/hcl/v2"

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
	_, ok = pkgSchema.resources[token]
	if !ok {
		return hcl.Diagnostics{unknownResourceType(token, tokenRange)}
	}

	// Create input and output types for the schema.
	return notYetImplemented(node)
}
