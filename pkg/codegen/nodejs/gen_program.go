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

package nodejs

type programGenerator struct {
	outputDirectory string
	diagnostics     hcl.Diagnostics
}

func GenerateProgram(program *model.Program, outputDirectory string) (hcl.Diagnostics, error) {
	// Linearize the nodes into an order appropriate for procedural code generation.
	nodes := program.Linearize()

	g := &programGenerator{
		outputDirectory: outputDirectory,
	}

	if err := g.generatePreamble(program); err != nil {
		return g.diagnostics, err
	}

	for _, n := range nodes {
		if err := g.generateNode(n); err != nil {
			return g.diagnostics, err
		}
	}

	return g.diagnostics, nil
}

func tsName(pulumiName string, isObjectKey bool) string {
	if !isLegalIdentifier(pulumiName) {
		if isObjectKey {
			return fmt.Sprintf("%q", pulumiName)
		}
		return cleanName(pulumiName)
	}
	return pulumiName
}

// GeneratePreamble generates appropriate import statements based on the providers referenced by the set of modules.
func (g *generator) GeneratePreamble(modules []*il.Graph) error {
	// Find the root module and stash its path.
	for _, m := range modules {
		if m.IsRoot {
			g.rootPath = m.Path
			break
		}
	}
	if g.rootPath == "" {
		return errors.New("could not determine root module path")
	}

	// Print the @pulumi/pulumi import at the top.
	g.Println(`import * as pulumi from "@pulumi/pulumi";`)

	// Accumulate other imports for the various providers. Don't emit them yet, as we need to sort them later on.
	var imports []string
	providers := make(map[string]bool)
	for _, m := range modules {
		for _, p := range m.Providers {
			name := p.PluginName
			if !providers[name] {
				providers[name] = true
				switch name {
				case "archive":
					// Nothing to do
				case "http":
					imports = append(imports,
						`import rpn = require("request-promise-native");`)
					g.importNames["rpn"] = true
				default:
					importName := cleanName(name)
					imports = append(imports,
						fmt.Sprintf(`import * as %s from "@pulumi/%s";`, importName, name))
					g.importNames[importName] = true
				}
			}
		}
	}

	// Now sort the imports, so we emit them deterministically, and emit them.
	sort.Strings(imports)
	for _, line := range imports {
		g.Println(line)
	}
	g.Printf("\n")

	return nil
}

// BeginModule saves the indicated module in the generator and emits an appropriate function declaration if the module
// is a child module.
func (g *generator) BeginModule(m *il.Graph) error {
	g.module = m
	if !g.isRoot() {
		g.Printf("const new_mod_%s = function(mod_name: string, mod_args: pulumi.Inputs) {\n",
			cleanName(m.Name))
		g.Indent += "    "

		// Discover the set of input variables that may have unknown values. This is the complete set of inputs minus
		// the set of variables used in count interpolations, as Terraform requires that the latter are known at graph
		// generation time (and thus at Pulumi run time).
		knownInputs := make(map[*il.VariableNode]struct{})
		for _, n := range m.Resources {
			if n.Count != nil {
				_, err := il.VisitBoundNode(n.Count, il.IdentityVisitor, func(n il.BoundNode) (il.BoundNode, error) {
					if n, ok := n.(*il.BoundVariableAccess); ok {
						if v, ok := n.ILNode.(*il.VariableNode); ok {
							knownInputs[v] = struct{}{}
						}
					}
					return n, nil
				})
				contract.Assert(err == nil)
			}
		}
		g.unknownInputs = make(map[*il.VariableNode]struct{})
		for _, v := range m.Variables {
			if _, ok := knownInputs[v]; !ok {
				g.unknownInputs[v] = struct{}{}
			}
		}

		// Retype any possibly-unknown module inputs as the appropriate output type.
		err := il.VisitAllProperties(m, il.IdentityVisitor, func(n il.BoundNode) (il.BoundNode, error) {
			if n, ok := n.(*il.BoundVariableAccess); ok {
				if v, ok := n.ILNode.(*il.VariableNode); ok {
					if _, ok = g.unknownInputs[v]; ok {
						n.ExprType = n.ExprType.OutputOf()
					}
				}
			}
			return n, nil
		})
		contract.Assert(err == nil)
	}

	// Find all prompt datasources if possible.
	if g.usePromptDataSources {
		g.promptDataSources = il.MarkPromptDataSources(m)
	}

	// Find all conditional resources.
	g.conditionalResources = il.MarkConditionalResources(m)

	// Compute unambiguous names for this module's top-level nodes.
	g.nameTable = assignNames(m, g.importNames, g.isRoot())
	return nil
}

// EndModule closes the current module definition if the module is a child module and clears the generator's module
// field.
func (g *generator) EndModule(m *il.Graph) error {
	if !g.isRoot() {
		g.Indent = g.Indent[:len(g.Indent)-4]
		g.Printf("};\n")
	}
	g.module = nil
	return nil
}

// GenerateVariables generates definitions for the set of user variables in the context of the current module.
func (g *generator) GenerateVariables(vs []*il.VariableNode) error {
	// If there are no variables, we're done.
	if len(vs) == 0 {
		return nil
	}

	// Otherwise, what we do depends on whether or not we're generating the root module. If we are, then we generate
	// a config object and appropriate get/require calls; if we are not, we generate references into the module args.
	isRoot := g.isRoot()
	if isRoot {
		g.Printf("const config = new pulumi.Config();\n")
	}
	for _, v := range vs {
		configName := tsName(v.Name, nil, nil, false)
		_, isUnknown := g.unknownInputs[v]

		g.genLeadingComment(g, v.Comments)

		g.Printf("%sconst %s = ", g.Indent, g.nodeName(v))
		if v.DefaultValue == nil {
			if isRoot {
				g.Printf("config.require(\"%s\")", configName)
			} else {
				f := "mod_args[\"%s\"]"
				if isUnknown {
					f = "pulumi.output(" + f + ")"
				}
				g.Printf(f, configName)
			}
		} else {
			def, _, err := g.computeProperty(v.DefaultValue, false, "")
			if err != nil {
				return err
			}

			if isRoot {
				get := "get"
				switch v.DefaultValue.Type() {
				case il.TypeBool:
					get = "getBoolean"
				case il.TypeNumber:
					get = "getNumber"
				}
				g.Printf("config.%v(\"%s\") || %s", get, configName, def)
			} else {
				f := "mod_args[\"%s\"] || %s"
				if isUnknown {
					f = "pulumi.output(" + f + ")"
				}
				g.Printf(f, configName, def)
			}
		}
		g.Printf(";")

		g.genTrailingComment(g, v.Comments)
		g.Printf("\n")
	}
	g.Printf("\n")

	return nil
}

// GenerateLocal generates a single local value. These values are generated as local variable definitions.
func (g *generator) GenerateLocal(l *il.LocalNode) error {
	value, _, err := g.computeProperty(l.Value, false, "")
	if err != nil {
		return err
	}

	g.genLeadingComment(g, l.Comments)
	g.Printf("%sconst %s = %s;", g.Indent, g.nodeName(l), value)
	g.genTrailingComment(g, l.Comments)
	g.Print("\n")

	return nil
}

// GenerateModule generates a single module instantiation. A module instantiation is generated as a call to the
// appropriate module factory function; the result is assigned to a local variable.
func (g *generator) GenerateModule(m *il.ModuleNode) error {
	// generate a call to the module constructor
	args, _, err := g.computeProperty(m.Properties, false, "")
	if err != nil {
		return err
	}

	instanceName, modName := g.nodeName(m), cleanName(m.Name)
	g.genLeadingComment(g, m.Comments)
	g.Printf("%sconst %s = new_mod_%s(\"%s\", %s);", g.Indent, instanceName, modName, instanceName, args)
	g.genTrailingComment(g, m.Comments)
	g.Print("\n")

	return nil
}

// GenerateProvider generates a single provider instantiation. Each provider instantiation is generated as a call to
// the appropriate provider constructor that is assigned to a local variable.
func (g *generator) GenerateProvider(p *il.ProviderNode) error {
	// If this provider has no alias, ignore it.
	if p.Alias == "" {
		return nil
	}

	g.genLeadingComment(g, p.Comments)

	name := g.nodeName(p)
	qualifiedMemberName := p.PluginName + ".Provider"

	inputs, _, err := g.computeProperty(il.BoundNode(p.Properties), false, "")
	if err != nil {
		return err
	}

	var resName string
	if g.isRoot() {
		resName = fmt.Sprintf("\"%s\"", p.Alias)
	} else {
		resName = fmt.Sprintf("`${mod_name}_%s`", p.Alias)
	}

	g.Printf("%sconst %s = new %s(%s, %s);", g.Indent, name, qualifiedMemberName, resName, inputs)
	g.genTrailingComment(g, p.Comments)
	g.Print("\n")
	return nil
}

// resourceTypeName computes the NodeJS package, module, and type name for the given resource.
func resourceTypeName(r *il.ResourceNode) (string, string, string, error) {
	// Compute the resource type from the Terraform type.
	underscore := strings.IndexRune(r.Type, '_')
	if underscore == -1 {
		return "", "", "", errors.New("NYI: single-resource providers")
	}
	provider, resourceType := cleanName(r.Provider.PluginName), r.Type[underscore+1:]

	// Convert the TF resource type into its Pulumi name.
	memberName := tfbridge.TerraformToPulumiName(resourceType, nil, true)

	// Compute the module in which the Pulumi type definition lives.
	module := ""
	if tok, ok := r.Tok(); ok {
		components := strings.Split(tok, ":")
		if len(components) != 3 {
			return "", "", "", errors.Errorf("unexpected resource token format %s", tok)
		}

		mod, typ := components[1], components[2]

		slash := strings.IndexRune(mod, '/')
		if slash == -1 {
			slash = len(mod)
		}

		module, memberName = mod[:slash], typ
		if module == "index" {
			module = ""
		}
	}

	return provider, module, memberName, nil
}

// makeResourceName returns the expression that should be emitted for a resource's "name" parameter given its base name
// and the count variable name, if any.
func (g *generator) makeResourceName(baseName, count string) string {
	if g.isRoot() {
		if count == "" {
			return fmt.Sprintf(`"%s"`, baseName)
		}
		return fmt.Sprintf("`%s-${%s}`", baseName, count)
	}
	baseName = fmt.Sprintf("${mod_name}_%s", baseName)
	if count == "" {
		return fmt.Sprintf("`%s`", baseName)
	}
	return fmt.Sprintf("`%s-${%s}`", baseName, count)
}

// generateResource handles the generation of instantiations of non-builtin resources.
func (g *generator) generateResource(r *il.ResourceNode) error {
	provider, module, memberName, err := resourceTypeName(r)
	if err != nil {
		return err
	}
	if module != "" {
		module = "." + module
	}

	var resourceOptions []string
	if r.Provider.Alias != "" {
		resourceOptions = append(resourceOptions, "provider: "+g.nodeName(r.Provider))
	}

	// Build the list of explicit deps, if any.
	if len(r.ExplicitDeps) != 0 && !r.IsDataSource {
		buf := &bytes.Buffer{}
		fmt.Fprintf(buf, "dependsOn: [")
		for i, n := range r.ExplicitDeps {
			if i > 0 {
				fmt.Fprintf(buf, ", ")
			}
			depRes := n.(*il.ResourceNode)
			if depRes.Count != nil {
				if g.isConditionalResource(depRes) {
					fmt.Fprintf(buf, "!")
				} else {
					fmt.Fprintf(buf, "...")
				}
			}
			fmt.Fprintf(buf, "%s", g.nodeName(depRes))
		}
		fmt.Fprintf(buf, "]")
		resourceOptions = append(resourceOptions, buf.String())
	}

	if r.Timeouts != nil {
		buf := &bytes.Buffer{}
		g.Fgenf(buf, "timeouts: %v", r.Timeouts)
		resourceOptions = append(resourceOptions, buf.String())
	}

	if len(r.IgnoreChanges) != 0 {
		buf := &bytes.Buffer{}
		fmt.Fprintf(buf, "ignoreChanges: [")
		for i, ic := range r.IgnoreChanges {
			if i > 0 {
				fmt.Fprintf(buf, ", ")
			}
			fmt.Fprintf(buf, "\"%s\"", ic)
		}
		fmt.Fprintf(buf, "]")
		resourceOptions = append(resourceOptions, buf.String())
	}

	if r.IsDataSource && !g.promptDataSources[r] {
		resourceOptions = append(resourceOptions, "async: true")
	}

	optionsBag := ""
	if len(resourceOptions) != 0 {
		optionsBag = fmt.Sprintf("{ %s }", strings.Join(resourceOptions, ", "))
	}

	name := g.nodeName(r)
	qualifiedMemberName := fmt.Sprintf("%s%s.%s", provider, module, memberName)

	// Because data sources are treated as normal function calls, we treat them a little bit differently by first
	// rewriting them into calls to the `__dataSource` intrinsic.
	properties := il.BoundNode(r.Properties)
	if r.IsDataSource {
		properties = newDataSourceCall(qualifiedMemberName, properties, optionsBag)
	}

	if optionsBag != "" {
		optionsBag = ", " + optionsBag
	}

	if r.Count == nil {
		// If count is nil, this is a single-instance resource.
		inputs, transformed, err := g.computeProperty(properties, false, "")
		if err != nil {
			return err
		}

		if !r.IsDataSource {
			resName := g.makeResourceName(r.Name, "")
			g.Printf("%sconst %s = new %s(%s, %s%s);", g.Indent, name, qualifiedMemberName, resName, inputs, optionsBag)
		} else {
			// TODO: explicit dependencies

			// If the input properties did not contain any outputs, then we need to wrap the result in a call to pulumi.output.
			// Otherwise, we are okay as-is: the apply rewrite perfomed by computeProperty will have ensured that the result
			// is output-typed.
			fmtstr := "%sconst %s = pulumi.output(%s);"
			if g.promptDataSources[r] || transformed {
				fmtstr = "%sconst %s = %s;"
			}

			g.Printf(fmtstr, g.Indent, name, inputs)
		}
	} else if g.isConditionalResource(r) {
		// If this is a confitional resource, we need to generate a resource that is instantiated inside an if statement.

		// If this resource's properties reference its count, we need to generate its code slightly differently:
		// a) We need to assign the value of the count to a local s.t. the properties have something to reference
		// b) We want to avoid changing the type of the count if it is not a boolean so that downstream code does not
		//    require changes.
		hasCountReference, countVariableName := false, ""
		_, err = il.VisitBoundNode(properties, il.IdentityVisitor, func(n il.BoundNode) (il.BoundNode, error) {
			if n, ok := n.(*il.BoundVariableAccess); ok {
				_, isCountVar := n.TFVar.(*config.CountVariable)
				hasCountReference = hasCountReference || isCountVar
			}
			return n, nil
		})
		contract.Assert(err == nil)

		// If the resource's properties do not reference the count, we can simplify the condition expression for
		// cleaner-looking code. We don't do this if the count is referenced because it can change the type of the
		// expression (e.g. from a number to a boolean, if the number is statically coerceable to a boolean).
		count := r.Count
		if !hasCountReference {
			count = il.SimplifyBooleanExpressions(count.(il.BoundExpr))
		}
		condition, _, err := g.computeProperty(count, false, "")
		if err != nil {
			return err
		}

		// If the resoure's properties reference the count, assign its value to a local s.t. the properties have
		// something to refer to.
		if hasCountReference {
			countVariableName = fmt.Sprintf("create%s", title(name))
			g.Printf("%sconst %s = %s;\n", g.Indent, countVariableName, condition)
			condition = countVariableName
		}

		inputs, transformed, err := g.computeProperty(properties, true, countVariableName)
		if err != nil {
			return err
		}

		g.Printf("%slet %s: %s | undefined;\n", g.Indent, name, qualifiedMemberName)
		ifFmt := "%sif (%s) {\n"
		if count.Type() != il.TypeBool {
			ifFmt = "%sif (!!(%s)) {\n"
		}
		g.Printf(ifFmt, g.Indent, condition)
		g.Indented(func() {
			if !r.IsDataSource {
				resName := g.makeResourceName(r.Name, "")
				g.Printf("%s%s = new %s(%s, %s%s);\n", g.Indent, name, qualifiedMemberName, resName, inputs, optionsBag)
			} else {
				// TODO: explicit dependencies

				// If the input properties did not contain any outputs, then we need to wrap the result in a call to pulumi.output.
				// Otherwise, we are okay as-is: the apply rewrite perfomed by computeProperty will have ensured that the result
				// is output-typed.
				fmtstr := "%s%s = pulumi.output(%s);\n"
				if g.promptDataSources[r] || transformed {
					fmtstr = "%s%s = %s;\n"
				}

				g.Printf(fmtstr, g.Indent, name, inputs)
			}
		})
		g.Printf("%s}", g.Indent)
	} else {
		// Otherwise we need to Generate multiple resources in a loop.
		count, _, err := g.computeProperty(r.Count, false, "")
		if err != nil {
			return err
		}
		inputs, transformed, err := g.computeProperty(properties, true, "i")
		if err != nil {
			return err
		}

		arrElementType := qualifiedMemberName
		if r.IsDataSource {
			fmtStr := "pulumi.Output<%s%s.%sResult>"
			if g.promptDataSources[r] {
				fmtStr = "%s%s.%sResult"
			}
			arrElementType = fmt.Sprintf(fmtStr, provider, module, strings.Title(memberName))
		}

		g.Printf("%sconst %s: %s[] = [];\n", g.Indent, name, arrElementType)
		g.Printf("%sfor (let i = 0; i < %s; i++) {\n", g.Indent, count)
		g.Indented(func() {
			if !r.IsDataSource {
				resName := g.makeResourceName(r.Name, "i")
				g.Printf("%s%s.push(new %s(%s, %s%s));\n", g.Indent, name, qualifiedMemberName, resName, inputs,
					optionsBag)
			} else {
				// TODO: explicit dependencies

				// If the input properties did not contain any outputs, then we need to wrap the result in a call to
				// pulumi.output. Otherwise, we are okay as-is: the apply rewrite perfomed by computeProperty will hav
				// ensured that the result is output-typed.
				fmtstr := "%s%s.push(pulumi.output(%s));\n"
				if g.promptDataSources[r] || transformed {
					fmtstr = "%s%s.push(%s);\n"
				}

				g.Printf(fmtstr, g.Indent, name, inputs)
			}
		})
		g.Printf("%s}", g.Indent)
	}

	return nil
}

// GenerateResource generates a single resource instantiation. Each resource instantiation is generated as a call or
// sequence of calls (in the case of a counted resource) to the approriate resource constructor or data source
// function. Single-instance resources are assigned to a local variable; counted resources are stored in an array-typed
// local.
func (g *generator) GenerateResource(r *il.ResourceNode) error {
	g.genLeadingComment(g, r.Comments)

	// If this resource's provider is one of the built-ins, perform whatever provider-specific code generation is
	// required.
	var err error
	switch r.Provider.Name {
	case "archive":
		err = g.generateArchive(r)
	case "http":
		err = g.generateHTTP(r)
	default:
		err = g.generateResource(r)
	}
	if err != nil {
		return err
	}

	g.genTrailingComment(g, r.Comments)
	g.Print("\n")
	return nil
}
