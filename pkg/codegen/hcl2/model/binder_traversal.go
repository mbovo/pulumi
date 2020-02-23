package model

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/pulumi/pulumi/pkg/util/contract"
	"github.com/zclconf/go-cty/cty"
)

func (b *binder) bindTraversalTypes(receiver Type, traversal hcl.Traversal) ([]Type, hcl.Diagnostics) {
	types := make([]Type, len(traversal)+1)
	types[0] = receiver

	var diagnostics hcl.Diagnostics
	for i, part := range traversal {
		var index cty.Value
		switch part := part.(type) {
		case hcl.TraverseAttr:
			index = cty.StringVal(part.Name)
		case hcl.TraverseIndex:
			index = part.Key
		default:
			contract.Failf("unexpected traversal part of type %T (%v)", part, part.SourceRange())
		}

		nextReceiver, indexDiags := b.bindIndexType(receiver, ctyTypeToType(index.Type(), false), index, part.SourceRange())
		types[i+1], receiver, diagnostics = nextReceiver, nextReceiver, append(diagnostics, indexDiags...)
	}

	return types, diagnostics
}

func (b *binder) bindIndexType(receiver Type, indexType Type, indexVal cty.Value, indexRange hcl.Range) (Type, hcl.Diagnostics) {
	switch receiver := receiver.(type) {
	case *OptionalType:
		elementType, diagnostics := b.bindIndexType(receiver.ElementType, indexType, indexVal, indexRange)
		return NewOptionalType(elementType), diagnostics
	case *OutputType:
		elementType, diagnostics := b.bindIndexType(receiver.ElementType, indexType, indexVal, indexRange)
		return NewOutputType(elementType), diagnostics
	case *PromiseType:
		elementType, diagnostics := b.bindIndexType(receiver.ElementType, indexType, indexVal, indexRange)
		return NewPromiseType(elementType), diagnostics
	case *MapType:
		var diagnostics hcl.Diagnostics
		if !inputType(StringType).AssignableFrom(indexType) {
			diagnostics = hcl.Diagnostics{unsupportedMapKey(indexRange)}
		}
		return receiver.ElementType, diagnostics
	case *ArrayType:
		var diagnostics hcl.Diagnostics
		if !inputType(NumberType).AssignableFrom(indexType) {
			diagnostics = hcl.Diagnostics{unsupportedArrayIndex(indexRange)}
		}
		return receiver.ElementType, diagnostics
	case *ObjectType:
		if !inputType(StringType).AssignableFrom(indexType) {
			return AnyType, hcl.Diagnostics{unsupportedObjectProperty(indexRange)}
		}

		if indexVal == cty.DynamicVal {
			return AnyType, nil
		}

		propertyName := indexVal.AsString()
		propertyType, hasProperty := receiver.Properties[propertyName]
		if !hasProperty {
			return AnyType, hcl.Diagnostics{unknownObjectProperty(propertyName, indexRange)}
		}
		return propertyType, nil
	default:
		if receiver == AnyType {
			return AnyType, nil
		}
		return AnyType, hcl.Diagnostics{unsupportedReceiverType(receiver, indexRange)}
	}
}
