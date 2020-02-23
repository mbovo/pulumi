package format

import (
	"fmt"
	"io"

	"github.com/hashicorp/hcl/v2"
	"github.com/pulumi/pulumi/pkg/codegen/hcl2/model"
	"github.com/pulumi/pulumi/pkg/util/contract"
)

type ExpressionGenerator interface {
	GenAnonymousFunctionExpression(w io.Writer, expr *model.AnonymousFunctionExpression)
	GenBinaryOpExpression(w io.Writer, expr *model.BinaryOpExpression)
	GenConditionalExpression(w io.Writer, expr *model.ConditionalExpression)
	GenForExpression(w io.Writer, expr *model.ForExpression)
	GenFunctionCallExpression(w io.Writer, expr *model.FunctionCallExpression)
	GenIndexExpression(w io.Writer, expr *model.IndexExpression)
	GenLiteralValueExpression(w io.Writer, expr *model.LiteralValueExpression)
	GenObjectConsExpression(w io.Writer, expr *model.ObjectConsExpression)
	GenRelativeTraversalExpression(w io.Writer, expr *model.RelativeTraversalExpression)
	GenScopeTraversalExpression(w io.Writer, expr *model.ScopeTraversalExpression)
	GenSplatExpression(w io.Writer, expr *model.SplatExpression)
	GenTemplateExpression(w io.Writer, expr *model.TemplateExpression)
	GenTemplateJoinExpression(w io.Writer, expr *model.TemplateJoinExpression)
	GenTupleConsExpression(w io.Writer, expr *model.TupleConsExpression)
	GenUnaryOpExpression(w io.Writer, expr *model.UnaryOpExpression)
}

// Formatter is a convenience type that implements a number of common utilities used to emit source code. It implements
// the io.Writer interface.
type Formatter struct {
	// The current indent level as a string.
	Indent string

	// The ExpressionGenerator to use in {G,Fg}en{,f}
	g ExpressionGenerator
}

// NewFormatter creates a new emitter targeting the given io.Writer that will use the given ExpressionGenerator when
// generating code.
func NewFormatter(g ExpressionGenerator) *Formatter {
	return &Formatter{g: g}
}

// indented bumps the current indentation level, invokes the given function, and then resets the indentation level to
// its prior value.
func (e *Formatter) Indented(f func()) {
	e.Indent += "    "
	f()
	e.Indent = e.Indent[:len(e.Indent)-4]
}

// Print prints one or more values to the generator's output stream.
func (e *Formatter) Fprint(w io.Writer, a ...interface{}) {
	_, err := fmt.Fprint(w, a...)
	contract.IgnoreError(err)
}

// Println prints one or more values to the generator's output stream, followed by a newline.
func (e *Formatter) Fprintln(w io.Writer, a ...interface{}) {
	e.Fprint(w, a...)
	e.Fprint(w, "\n")
}

// Printf prints a formatted message to the generator's output stream.
func (e *Formatter) Fprintf(w io.Writer, format string, a ...interface{}) {
	_, err := fmt.Fprintf(w, format, a...)
	contract.IgnoreError(err)
}

// Fgen generates code for a list of strings and expression trees. The former are written directly to the destination;
// the latter are recursively generated using the appropriate gen* functions.
func (e *Formatter) Fgen(w io.Writer, vs ...interface{}) {
	for _, v := range vs {
		switch v := v.(type) {
		case string:
			_, err := fmt.Fprint(w, v)
			contract.IgnoreError(err)
		case *model.AnonymousFunctionExpression:
			e.g.GenAnonymousFunctionExpression(w, v)
		case *model.BinaryOpExpression:
			e.g.GenBinaryOpExpression(w, v)
		case *model.ConditionalExpression:
			e.g.GenConditionalExpression(w, v)
		case *model.ForExpression:
			e.g.GenForExpression(w, v)
		case *model.FunctionCallExpression:
			e.g.GenFunctionCallExpression(w, v)
		case *model.IndexExpression:
			e.g.GenIndexExpression(w, v)
		case *model.LiteralValueExpression:
			e.g.GenLiteralValueExpression(w, v)
		case *model.ObjectConsExpression:
			e.g.GenObjectConsExpression(w, v)
		case *model.RelativeTraversalExpression:
			e.g.GenRelativeTraversalExpression(w, v)
		case *model.ScopeTraversalExpression:
			e.g.GenScopeTraversalExpression(w, v)
		case *model.SplatExpression:
			e.g.GenSplatExpression(w, v)
		case *model.TemplateExpression:
			e.g.GenTemplateExpression(w, v)
		case *model.TemplateJoinExpression:
			e.g.GenTemplateJoinExpression(w, v)
		case *model.TupleConsExpression:
			e.g.GenTupleConsExpression(w, v)
		case *model.UnaryOpExpression:
			e.g.GenUnaryOpExpression(w, v)
		default:
			var rng hcl.Range
			if v, isExpr := v.(model.Expression); isExpr {
				rng = v.SyntaxNode().Range()
			}
			contract.Failf("unexpected expression node of type %T (%v)", v, rng)
		}
	}
}

// Fgenf generates code using a format string and its arguments. Any arguments that are BoundNode values are wrapped in
// a FormatFunc that calls the appropriate recursive generation function. This allows for the composition of standard
// format strings with expression/property code gen (e.e. `e.genf(w, ".apply(__arg0 => %v)", then)`, where `then` is
// an expression tree).
func (e *Formatter) Fgenf(w io.Writer, format string, args ...interface{}) {
	for i := range args {
		if node, ok := args[i].(model.Expression); ok {
			args[i] = FormatFunc(func(f fmt.State, c rune) { e.Fgen(f, node) })
		}
	}
	fmt.Fprintf(w, format, args...)
}
