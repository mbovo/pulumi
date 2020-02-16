package model

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/pulumi/pulumi/pkg/util/contract"
)

func (b *binder) bindNode(node Node) hcl.Diagnostics {
	switch node.getState() {
	case binding:
		// Circular reference
		return hcl.Diagnostics{circularReference(b.stack, node.SyntaxNode())}
	case bound:
		// Already done
		return nil
	}

	node.setState(binding)
	b.stack = append(b.stack, node.SyntaxNode())
	defer func() {
		b.stack = b.stack[:len(b.stack)-1]
		node.setState(bound)
	}()

	// Bind the node's dependencies.
	var diagnostics hcl.Diagnostics
	for _, dep := range b.getDependencies(node) {
		diagnostics = append(diagnostics, b.bindNode(dep)...)
	}

	switch node := node.(type) {
	case *ConfigVariable:
		diagnostics = append(diagnostics, b.bindConfigVariable(node)...)
	case *LocalVariable:
		diagnostics = append(diagnostics, b.bindLocalVariable(node)...)
	case *Resource:
		diagnostics = append(diagnostics, b.bindResource(node)...)
	case *OutputVariable:
		diagnostics = append(diagnostics, b.bindOutputVariable(node)...)
	default:
		contract.Failf("unexpected node of type %T (%v)", node, node.SyntaxNode().Range())
	}
	return diagnostics
}

func (b *binder) getDependencies(node Node) []Node {
	depSet := nodeSet{}
	var deps []Node
	hclsyntax.VisitAll(node.SyntaxNode(), func(node hclsyntax.Node) hcl.Diagnostics {
		depName := ""
		switch node := node.(type) {
		case *hclsyntax.FunctionCallExpr:
			depName = node.Name
		case *hclsyntax.ScopeTraversalExpr:
			depName = node.Traversal.RootName()
		default:
			return nil
		}

		// Missing reference errors will be issued during expression binding.
		if referent, ok := b.nodes[depName]; ok && !depSet.has(referent) {
			depSet.add(referent)
			deps = append(deps, referent)
		}
		return nil
	})
	return sourceOrderNodes(deps)
}

func (b *binder) bindConfigVariable(node *ConfigVariable) hcl.Diagnostics {
	return notYetImplemented(node)
}

func (b *binder) bindLocalVariable(node *LocalVariable) hcl.Diagnostics {
	return notYetImplemented(node)
}

func (b *binder) bindOutputVariable(node *OutputVariable) hcl.Diagnostics {
	return notYetImplemented(node)
}
