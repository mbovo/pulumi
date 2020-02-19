package model

import (
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type stringSet map[string]struct{}

func (ss stringSet) add(s string) {
	ss[s] = struct{}{}
}

func (ss stringSet) has(s string) bool {
	_, ok := ss[s]
	return ok
}

func (ss stringSet) sortedValues() []string {
	values := make([]string, 0, len(ss))
	for v := range ss {
		values = append(values, v)
	}
	sort.Strings(values)
	return values
}

type nodeSet map[Node]struct{}

func (ns nodeSet) add(n Node) {
	ns[n] = struct{}{}
}

func (ns nodeSet) has(n Node) bool {
	_, ok := ns[n]
	return ok
}

type typeSet map[Type]struct{}

func (ts typeSet) add(t Type) {
	ts[t] = struct{}{}
}

func (ts typeSet) has(t Type) bool {
	_, ok := ts[t]
	return ok
}

func sourceOrderBlocks(blocks []*hclsyntax.Block) []*hclsyntax.Block {
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Range().Start.Byte < blocks[j].Range().Start.Byte
	})
	return blocks
}

func sourceOrderAttributes(attrMap map[string]*hclsyntax.Attribute) []*hclsyntax.Attribute {
	var attrs []*hclsyntax.Attribute
	for _, attr := range attrMap {
		attrs = append(attrs, attr)
	}
	sort.Slice(attrs, func(i, j int) bool {
		return attrs[i].Range().Start.Byte < attrs[j].Range().End.Byte
	})
	return attrs
}

func sourceOrderNodes(nodes []Node) []Node {
	sort.Slice(nodes, func(i, j int) bool {
		ir, jr := nodes[i].SyntaxNode().Range(), nodes[j].SyntaxNode().Range()
		return ir.Filename < jr.Filename || ir.Start.Byte < jr.Start.Byte
	})
	return nodes
}

func decomposeToken(tok string, sourceRange hcl.Range) (string, string, string, hcl.Diagnostics) {
	components := strings.Split(tok, ":")
	if len(components) != 3 {
		return "", "", "", hcl.Diagnostics{malformedToken(tok, sourceRange)}
	}
	return components[0], components[1], components[2], nil
}
