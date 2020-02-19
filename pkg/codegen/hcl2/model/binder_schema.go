package model

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/pulumi/pulumi/pkg/codegen/schema"
	"github.com/pulumi/pulumi/pkg/tokens"
	"github.com/pulumi/pulumi/pkg/util/contract"
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

func schemaTypeToType(src schema.Type, wrapInput bool) Type {
	var dst Type
	switch src := src.(type) {
	case *schema.ArrayType:
		dst = NewArrayType(schemaTypeToType(src.ElementType, wrapInput))
	case *schema.MapType:
		dst = NewMapType(schemaTypeToType(src.ElementType, wrapInput))
	case *schema.ObjectType:
		properties := map[string]Type{}
		for _, prop := range src.Properties {
			t := schemaTypeToType(prop.Type, wrapInput)
			if !prop.IsRequired {
				t = NewOptionalType(t)
			}
			properties[prop.Name] = t
		}
		dst = NewObjectType(properties)
	case *schema.TokenType:
		t, ok := GetTokenType(src.Token)
		if !ok {
			tt, err := NewTokenType(src.Token)
			contract.IgnoreError(err)
			t = tt
		}
		dst = t

		if wrapInput {
			dst = NewUnionType(dst, NewOutputType(dst))
		}
		if src.UnderlyingType == nil {
			return dst
		}

		underlyingType := schemaTypeToType(src.UnderlyingType, wrapInput)
		return NewUnionType(dst, underlyingType)
	case *schema.UnionType:
		switch len(src.ElementTypes) {
		case 0:
			return nil
		case 1:
			return schemaTypeToType(src.ElementTypes[0], wrapInput)
		default:
			types := make([]Type, len(src.ElementTypes))
			for i, src := range src.ElementTypes {
				types[i] = schemaTypeToType(src, wrapInput)
			}
			dst = NewUnionType(types[0], types[1], types[2:]...)
		}
	default:
		switch src {
		case schema.BoolType:
			dst = BoolType
		case schema.IntType:
			dst = IntType
		case schema.NumberType:
			dst = NumberType
		case schema.StringType:
			dst = StringType
		case schema.ArchiveType:
			dst = ArchiveType
		case schema.AssetType:
			dst = AssetType
		case schema.AnyType:
			return AnyType
		}
	}

	if wrapInput {
		dst = NewUnionType(dst, NewOutputType(dst))
	}
	return dst
}
