// Copyright 2016-2020, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func errorf(subject hcl.Range, f string, args ...interface{}) *hcl.Diagnostic {
	return diagf(hcl.DiagError, subject, f, args...)
}

func diagf(severity hcl.DiagnosticSeverity, subject hcl.Range, f string, args ...interface{}) *hcl.Diagnostic {
	message := fmt.Sprintf(f, args...)
	return &hcl.Diagnostic{
		Severity: severity,
		Summary:  message,
		Detail:   message,
		Subject:  &subject,
	}
}

func labelsErrorf(block *hclsyntax.Block, f string, args ...interface{}) *hcl.Diagnostic {
	startRange := block.LabelRanges[0]

	diagRange := hcl.Range{
		Filename: startRange.Filename,
		Start:    startRange.Start,
		End:      block.LabelRanges[len(block.LabelRanges)-1].End,
	}
	return errorf(diagRange, f, args...)
}

func notYetImplemented(v interface{}) hcl.Diagnostics {
	var subject hcl.Range
	switch v := v.(type) {
	case Node:
		subject = v.SyntaxNode().Range()
	case interface{ Range() hcl.Range }:
		subject = v.Range()
	}
	return hcl.Diagnostics{errorf(subject, "NYI: %v", v)}
}

func malformedToken(token string, sourceRange hcl.Range) *hcl.Diagnostic {
	return errorf(sourceRange, "malformed token '%v': expected 'pkg:module:member'", token)
}

func circularReference(stack []hclsyntax.Node, referent hclsyntax.Node) *hcl.Diagnostic {
	// TODO(pdg): stack trace
	return errorf(referent.Range(), "circular reference to node")
}

func unknownPackage(pkg string, tokenRange hcl.Range) *hcl.Diagnostic {
	return errorf(tokenRange, "unknown package '%s'", pkg)
}

func unknownResourceType(token string, tokenRange hcl.Range) *hcl.Diagnostic {
	return errorf(tokenRange, "unknown resource type '%s'", token)
}

func exprNotAssignable(destType Type, expr Expression) *hcl.Diagnostic {
	return errorf(expr.SyntaxNode().Range(), "cannot assign expression of type %v to location of type %v", expr.Type(), destType)
}

func typesNotAssignable(destType, srcType Type, srcRange hcl.Range) *hcl.Diagnostic {
	return errorf(srcRange, "cannot assign expression of type %v to location of type %v", srcType, destType)
}

func objectKeysMustBeStrings(expr Expression) *hcl.Diagnostic {
	return errorf(expr.SyntaxNode().Range(), "object keys must be strings: cannot assign expression of type %v to location of type string", expr.Type())
}

func unsupportedLiteralValue(syntax *hclsyntax.LiteralValueExpr) *hcl.Diagnostic {
	return errorf(syntax.Range(), "unsupported literal value of type %v", syntax.Val.Type())
}

func unknownFunction(name string, nameRange hcl.Range) *hcl.Diagnostic {
	return errorf(nameRange, "unknown function '%s'", name)
}

func missingRequiredArgument(param Parameter, callRange hcl.Range) *hcl.Diagnostic {
	return errorf(callRange, "missing required parameter '%s'", param.Name)
}

func extraArguments(expected, actual int, callRange hcl.Range) *hcl.Diagnostic {
	return errorf(callRange, "too many arguments to call: expected %v, got %v", expected, actual)
}

func unsupportedIndexKey(key hcl.Traverser) *hcl.Diagnostic {
	return errorf(key.SourceRange(), "keys must be strings or numbers")
}

func unsupportedMapKey(keyRange hcl.Range) *hcl.Diagnostic {
	return errorf(keyRange, "map keys must be strings")
}

func unsupportedArrayIndex(indexRange hcl.Range) *hcl.Diagnostic {
	return errorf(indexRange, "array indexes must be numbers")
}

func unsupportedObjectProperty(indexRange hcl.Range) *hcl.Diagnostic {
	return errorf(indexRange, "object properties must be strings")
}

func unknownObjectProperty(name string, indexRange hcl.Range) *hcl.Diagnostic {
	return errorf(indexRange, "unknown property '%s'", name)
}

func unsupportedReceiverType(receiver Type, indexRange hcl.Range) *hcl.Diagnostic {
	return errorf(indexRange, "cannot index value of type %v", receiver)
}

func unsupportedCollectionType(collectionType Type, iteratorRange hcl.Range) *hcl.Diagnostic {
	return errorf(iteratorRange, "cannot iterator over a value of type %v", collectionType)
}

func undefinedVariable(variableRange hcl.Range) *hcl.Diagnostic {
	return errorf(variableRange, "undefined variable")
}

func internalError(rng hcl.Range, fmt string, args ...interface{}) *hcl.Diagnostic {
	return errorf(rng, "Internal error: "+fmt, args...)
}

func nameAlreadyDefined(name string, rng hcl.Range) *hcl.Diagnostic {
	return errorf(rng, "name %v already defined", name)
}
