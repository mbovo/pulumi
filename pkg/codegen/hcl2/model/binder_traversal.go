package model

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/pulumi/pulumi/pkg/resource"
	"github.com/zclconf/go-cty/cty"
)

func (b *binder) bindTraversalType(receiver Type, traversal hcl.Traversal) (Type, hcl.Diagnostics) {
	var diagnostics hcl.Diagnostics
	for _, part := range traversal {
		var key cty.Value
		switch part := part.(type) {
		case hcl.TraverseAttr:
			key = cty.StringVal(part.Name)
		case hcl.TraverseIndex:
			key = part.Key
		}

		var index resource.PropertyValue
		switch {
		case key.Type() == cty.Number:
			f, _ := key.AsBigFloat().Float64()
			index = resource.NewNumberProperty(f)
		case key.Type() == cty.String:
			index = resource.NewStringProperty(key.AsString())
		default:
			return AnyType, hcl.Diagnostics{unsupportedIndexKey(part)}
		}

		nextReceiver, indexDiags := b.bindIndexType(receiver, index, part.SourceRange())
		receiver, diagnostics = nextReceiver, append(diagnostics, indexDiags...)
	}

	return receiver, diagnostics
}

func (b *binder) bindIndexType(receiver Type, index resource.PropertyValue, indexRange hcl.Range) (Type, hcl.Diagnostics) {
	switch receiver := receiver.(type) {
	case *OptionalType:
		elementType, diagnostics := b.bindIndexType(receiver.ElementType, index, indexRange)
		return NewOptionalType(elementType), diagnostics
	case *OutputType:
		elementType, diagnostics := b.bindIndexType(receiver.ElementType, index, indexRange)
		return NewOutputType(elementType), diagnostics
	case *PromiseType:
		elementType, diagnostics := b.bindIndexType(receiver.ElementType, index, indexRange)
		return NewPromiseType(elementType), diagnostics
	case *MapType:
		var diagnostics hcl.Diagnostics
		if !index.IsString() {
			diagnostics = hcl.Diagnostics{unsupportedMapKey(indexRange)}
		}
		return receiver.ElementType, diagnostics
	case *ArrayType:
		var diagnostics hcl.Diagnostics
		if !index.IsNumber() {
			diagnostics = hcl.Diagnostics{unsupportedArrayIndex(indexRange)}
		}
		return receiver.ElementType, diagnostics
	case *ObjectType:
		if !index.IsString() {
			return AnyType, hcl.Diagnostics{unsupportedObjectProperty(indexRange)}
		}

		propertyName := index.StringValue()
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
