package model

import "github.com/hashicorp/hcl/v2"

type parameter struct {
	name string
	typ  Type
}

type functionSignature struct {
	parameters       []parameter
	varargsParameter *parameter
	returnType       Type
}

type functionDefinition func(arguments []Expression) (functionSignature, hcl.Diagnostics)

func (f functionDefinition) signature(arguments []Expression) (functionSignature, hcl.Diagnostics) {
	return f(arguments)
}

func getFunctionDefinition(name string, nameRange hcl.Range) (functionDefinition, hcl.Diagnostics) {
	switch name {
	case "fileAsset":
		return func(arguments []Expression) (functionSignature, hcl.Diagnostics) {
			return functionSignature{
				parameters: []parameter{{
					name: "path",
					typ:  StringType,
				}},
				returnType: AssetType,
			}, nil
		}, nil
	case "mimeType":
		return func(arguments []Expression) (functionSignature, hcl.Diagnostics) {
			return functionSignature{
				parameters: []parameter{{
					name: "path",
					typ:  StringType,
				}},
				returnType: StringType,
			}, nil
		}, nil
	case "toJSON":
		return func(arguments []Expression) (functionSignature, hcl.Diagnostics) {
			return functionSignature{
				parameters: []parameter{{
					name: "value",
					typ:  AnyType,
				}},
				returnType: StringType,
			}, nil
		}, nil
	default:
		return nil, hcl.Diagnostics{unknownFunction(name, nameRange)}
	}
}
