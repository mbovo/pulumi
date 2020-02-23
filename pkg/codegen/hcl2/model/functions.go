package model

import "github.com/hashicorp/hcl/v2"

type Parameter struct {
	Name string
	Type Type
}

type FunctionSignature struct {
	Parameters       []Parameter
	VarargsParameter *Parameter
	ReturnType       Type
}

type functionDefinition func(arguments []Expression) (FunctionSignature, hcl.Diagnostics)

func (f functionDefinition) signature(arguments []Expression) (FunctionSignature, hcl.Diagnostics) {
	return f(arguments)
}

func getFunctionDefinition(name string, nameRange hcl.Range) (functionDefinition, hcl.Diagnostics) {
	switch name {
	case "fileAsset":
		return func(arguments []Expression) (FunctionSignature, hcl.Diagnostics) {
			return FunctionSignature{
				Parameters: []Parameter{{
					Name: "path",
					Type: StringType,
				}},
				ReturnType: AssetType,
			}, nil
		}, nil
	case "mimeType":
		return func(arguments []Expression) (FunctionSignature, hcl.Diagnostics) {
			return FunctionSignature{
				Parameters: []Parameter{{
					Name: "path",
					Type: StringType,
				}},
				ReturnType: StringType,
			}, nil
		}, nil
	case "toJSON":
		return func(arguments []Expression) (FunctionSignature, hcl.Diagnostics) {
			return FunctionSignature{
				Parameters: []Parameter{{
					Name: "value",
					Type: AnyType,
				}},
				ReturnType: StringType,
			}, nil
		}, nil
	default:
		return nil, hcl.Diagnostics{unknownFunction(name, nameRange)}
	}
}
