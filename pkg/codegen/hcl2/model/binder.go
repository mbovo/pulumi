package model

import (
	"os"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/pulumi/pulumi/pkg/codegen/hcl2/syntax"
	"github.com/pulumi/pulumi/pkg/resource/plugin"
)

type binder struct {
	host plugin.Host

	packageSchemas map[string]*packageSchema

	stack       []hclsyntax.Node
	anonSymbols map[*hclsyntax.AnonSymbolExpr]*LocalVariable
	scopes      *scopes
	root        scope
}

func BindProgram(files []*syntax.File, host plugin.Host) (*Program, hcl.Diagnostics, error) {
	if host == nil {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, nil, err
		}
		ctx, err := plugin.NewContext(nil, nil, nil, nil, cwd, nil, nil)
		if err != nil {
			return nil, nil, err
		}
		host = ctx.Host
	}

	b := &binder{
		host:           host,
		packageSchemas: map[string]*packageSchema{},
		anonSymbols:    map[*hclsyntax.AnonSymbolExpr]*LocalVariable{},
		scopes:         &scopes{},
	}
	b.root = b.scopes.push()

	var diagnostics hcl.Diagnostics

	// Sort files in source order, then declare all top-level nodes in each.
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})
	for _, f := range files {
		diagnostics = append(diagnostics, b.declareNodes(f)...)
	}

	// Sort nodes in source order so downstream operations are deterministic.
	var nodes []Node
	for _, n := range b.root {
		nodes = append(nodes, n)
	}
	sourceOrderNodes(nodes)

	// Load referenced package schemas.
	for _, n := range nodes {
		if err := b.loadReferencedPackageSchemas(n); err != nil {
			return nil, nil, err
		}
	}

	// Now bind the nodes.
	for _, n := range nodes {
		diagnostics = append(diagnostics, b.bindNode(n)...)
	}

	return &Program{
		Nodes:  nodes,
		files:  files,
		binder: b,
	}, diagnostics, nil
}

func (b *binder) declareNodes(file *syntax.File) hcl.Diagnostics {
	var diagnostics hcl.Diagnostics

	// Declare blocks (config, resources, outputs), then attributes (locals)
	for _, block := range sourceOrderBlocks(file.Body.Blocks) {
		switch block.Type {
		case "config":
			if len(block.Labels) != 0 {
				diagnostics = append(diagnostics, labelsErrorf(block, "config blocks do not support labels"))
			}

			for _, attr := range sourceOrderAttributes(block.Body.Attributes) {
				diagnostics = append(diagnostics, errorf(attr.Range(), "unsupported attribute %q in config block", attr.Name))
			}

			for _, variable := range sourceOrderBlocks(block.Body.Blocks) {
				if len(variable.Labels) > 1 {
					diagnostics = append(diagnostics, labelsErrorf(block, "config variables must have no more than one label"))
				}

				diagnostics = append(diagnostics, b.declareNode(variable.Type, &ConfigVariable{
					Syntax: variable,
				})...)
			}
		case "resource":
			if len(block.Labels) != 2 {
				diagnostics = append(diagnostics, labelsErrorf(block, "resource variables must have exactly two labels"))
			}

			diagnostics = append(diagnostics, b.declareNode(block.Labels[0], &Resource{
				Syntax: block,
			})...)
		case "outputs":
			if len(block.Labels) != 0 {
				diagnostics = append(diagnostics, labelsErrorf(block, "outputs blocks do not support labels"))
			}

			for _, attr := range sourceOrderAttributes(block.Body.Attributes) {
				diagnostics = append(diagnostics, errorf(attr.Range(), "unsupported attribute %q in outputs block", attr.Name))
			}

			for _, variable := range sourceOrderBlocks(block.Body.Blocks) {
				if len(variable.Labels) > 1 {
					diagnostics = append(diagnostics, labelsErrorf(block, "output variables must have no more than one label"))
				}

				diagnostics = append(diagnostics, b.declareNode(variable.Type, &OutputVariable{
					Syntax: variable,
				})...)
			}
		}
	}

	for _, attr := range sourceOrderAttributes(file.Body.Attributes) {
		diagnostics = append(diagnostics, b.declareNode(attr.Name, &LocalVariable{
			Syntax: attr,
		})...)
	}

	return diagnostics
}

func (b *binder) declareNode(name string, n Node) hcl.Diagnostics {
	if !b.root.define(name, n) {
		existing, _ := b.root.bindReference(name)
		return hcl.Diagnostics{errorf(existing.SyntaxNode().Range(), "%q already declared", name)}
	}
	return nil
}
