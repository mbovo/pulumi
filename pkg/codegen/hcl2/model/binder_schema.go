package model

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/pulumi/pulumi/pkg/codegen/schema"
	"github.com/pulumi/pulumi/pkg/tokens"
)

type packageSchema struct {
	schema    *schema.Package
	resources map[string]*schema.Resource
	functions map[string]*schema.Function
}

func canonicalizeToken(tok string, pkg *schema.Package) string {
	_, _, member, _ := decomposeToken(tok, hcl.Range{})
	return fmt.Sprintf("%s:%s:%s", pkg.Name, pkg.TokenToModule(tok), member)
}

func (b *binder) loadReferencedPackageSchemas(n Node) error {
	// TODO: package versions
	packageNames := stringSet{}

	if r, ok := n.(*Resource); ok {
		token, tokenRange := getResourceToken(r)
		packageName, _, _, _ := decomposeToken(token, tokenRange)
		packageNames.add(packageName)
	}

	hclsyntax.VisitAll(n.SyntaxNode(), func(node hclsyntax.Node) hcl.Diagnostics {
		call, ok := node.(*hclsyntax.FunctionCallExpr)
		if !ok {
			return nil
		}
		token, tokenRange, ok := getInvokeToken(call)
		if !ok {
			return nil
		}
		packageName, _, _, _ := decomposeToken(token, tokenRange)
		packageNames.add(packageName)
		return nil
	})

	for _, name := range packageNames.sortedValues() {
		if err := b.loadPackageSchema(name); err != nil {
			return err
		}
	}
	return nil
}

// TODO: provider versions
func (b *binder) loadPackageSchema(name string) error {
	if _, ok := b.packageSchemas[name]; ok {
		return nil
	}

	provider, err := b.host.Provider(tokens.Package(name), nil)
	if err != nil {
		return err
	}

	schemaBytes, err := provider.GetSchema(0)
	if err != nil {
		return err
	}

	var spec schema.PackageSpec
	if err := json.Unmarshal(schemaBytes, &spec); err != nil {
		return err
	}

	pkg, err := schema.ImportSpec(spec)
	if err != nil {
		return err
	}

	resources := map[string]*schema.Resource{}
	for _, r := range pkg.Resources {
		resources[canonicalizeToken(r.Token, pkg)] = r
	}
	functions := map[string]*schema.Function{}
	for _, f := range pkg.Functions {
		functions[canonicalizeToken(f.Token, pkg)] = f
	}

	b.packageSchemas[name] = &packageSchema{
		schema:    pkg,
		resources: resources,
		functions: functions,
	}
	return nil
}
