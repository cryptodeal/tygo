package tygo

import (
	"fmt"
	"go/ast"
	"strings"
)

type groupContext struct {
	isGroupedDeclaration bool
	doc                  *ast.CommentGroup
	groupValue           string
	groupType            string
	iotaValue            int
	iotaOffset           int
}

func (g *PackageGenerator) writeGroupDecl(s *strings.Builder, decl *ast.GenDecl) {
	// This checks whether the declaration is a group declaration like:
	// const (
	// 	  X = 3
	//    Y = "abc"
	// )
	isGroupedDeclaration := len(decl.Specs) > 1

	if !isGroupedDeclaration {
		g.writeCommentGroupIfNotNil(s, decl.Doc, 0)
	}

	// We need a bit of state to handle syntax like
	// const (
	//   X SomeType = iota
	//   _
	//   Y
	//   Foo string = "Foo"
	//   _
	//   AlsoFoo
	// )
	group := &groupContext{
		isGroupedDeclaration: len(decl.Specs) > 1,
		doc:                  decl.Doc,
		groupType:            "",
		groupValue:           "",
		iotaValue:            -1,
	}

	for i, spec := range decl.Specs {
		isLast := i == len(decl.Specs)-1
		// fmt.Println("spec:", spec)
		g.writeSpec(s, spec, group, isLast)
	}
}

func (g *PackageGenerator) writeSpec(s *strings.Builder, spec ast.Spec, group *groupContext, isLast bool) {

	// e.g. "type Foo struct {}" or "type Bar = string"
	ts, ok := spec.(*ast.TypeSpec)
	if ok && ts.Name.IsExported() {
		// fmt.Println("ts:", ts)
		// fmt.Println("group:", group)
		g.writeTypeSpec(s, ts, group)
	}

	// e.g. "const Foo = 123"
	vs, ok := spec.(*ast.ValueSpec)
	if ok {
		// fmt.Println("vs:", vs)
		// fmt.Println("group:", group)
		g.writeValueSpec(s, vs, group, isLast)

	}
}

// Writing of type specs, which are expressions like
// `type X struct { ... }`
// or
// `type Bar = string`
func (g *PackageGenerator) writeTypeSpec(s *strings.Builder, ts *ast.TypeSpec, group *groupContext) {
	if ts.Doc != nil { // The spec has its own comment, which overrules the grouped comment.
		g.writeCommentGroup(s, ts.Doc, 0)
	} else if group.isGroupedDeclaration {
		g.writeCommentGroupIfNotNil(s, group.doc, 0)
	}

	st, isStruct := ts.Type.(*ast.StructType)
	if isStruct {
		s.WriteString("export interface ")
		s.WriteString(ts.Name.Name)

		if ts.TypeParams != nil {
			g.writeTypeParamsFields(s, ts.TypeParams.List)
		}

		s.WriteString(" {\n")
		g.writeStructFields(s, st.Fields.List, 0)
		s.WriteString("}")
	}

	id, isIdent := ts.Type.(*ast.Ident)
	if isIdent {
		if group.isGroupedDeclaration {
			s.WriteString("export enum ")
			s.WriteString(ts.Name.Name)
			s.WriteString(" {\n")
		} else {
			s.WriteString("export type ")
			s.WriteString(ts.Name.Name)
			s.WriteString(" = ")
			s.WriteString(getIdent(id.Name))
			s.WriteString(";")
		}
	}

	if !isStruct && !isIdent {
		// fmt.Println("!isStruct && !isIdent")
		s.WriteString("export type ")
		s.WriteString(ts.Name.Name)
		s.WriteString(" = ")
		g.writeType(s, ts.Type, 0, true)
		s.WriteString(";")

	}

	if ts.Comment != nil {
		s.WriteString(" // " + ts.Comment.Text())
	} else {
		s.WriteString("\n")
	}
}

// Writing of value specs, which are exported const expressions like
// const SomeValue = 3
func (g *PackageGenerator) writeValueSpec(s *strings.Builder, vs *ast.ValueSpec, group *groupContext, isLast bool) {
	for i, name := range vs.Names {
		group.iotaValue = group.iotaValue + 1
		if name.Name == "_" {
			continue
		}
		if !name.IsExported() {
			continue
		}

		if vs.Doc != nil { // The spec has its own comment, which overrules the grouped comment.
			g.writeCommentGroup(s, vs.Doc, 0)
		} else if group.isGroupedDeclaration {
			g.writeCommentGroupIfNotNil(s, group.doc, 0)
		}

		hasExplicitValue := len(vs.Values) > i
		if hasExplicitValue {
			group.groupType = ""
		}

		s.WriteString(name.Name)
		if vs.Type != nil {
			s.WriteString(name.Name)
			// fmt.Println("vs.Type != nil")

			tempSB := &strings.Builder{}
			g.writeType(tempSB, vs.Type, 0, true)
			typeString := tempSB.String()

			// s.WriteString(typeString)
			group.groupType = typeString
		} else if group.groupType != "" && !hasExplicitValue {
			s.WriteString("export const ")
			s.WriteString(name.Name)
			// fmt.Println("vs.Type = nil && group.groupType != '' && !hasExplicitValue")
			s.WriteString(": ")
			s.WriteString(group.groupType)
		}

		s.WriteString(" = ")

		if hasExplicitValue {
			val := vs.Values[i]
			tempSB := &strings.Builder{}
			g.writeType(tempSB, val, 0, true)
			valueString := tempSB.String()

			if isProbablyIotaType(valueString) {
				group.iotaOffset = basicIotaOffsetValueParse(valueString)
				group.groupValue = "iota"
				valueString = fmt.Sprint(group.iotaValue + group.iotaOffset)
			} else {
				group.groupValue = valueString
			}
			s.WriteString(valueString)

		} else { // We must use the previous value or +1 in case of iota
			valueString := group.groupValue
			if group.groupValue == "iota" {
				valueString = fmt.Sprint(group.iotaValue + group.iotaOffset)
			}
			s.WriteString(valueString)
		}
		if vs.Type != nil && !isLast {
			s.WriteByte(',')
		} else if group.groupType != "" && !hasExplicitValue {
			s.WriteByte(';')
		}

		if vs.Comment != nil {
			s.WriteString(" // " + vs.Comment.Text())
		} else {
			s.WriteByte('\n')
		}

		if vs.Type != nil && isLast && group.isGroupedDeclaration {
			s.WriteString("}")
		}
	}
}
