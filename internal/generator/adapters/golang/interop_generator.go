package golang

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

const (
	headerFileName = "interop_array.h"
)

type interopGenerator struct {
	Type any
}

type generatorState struct {
	outDir       string
	types        map[reflect.Type]*typeInfo
	orderedTypes []*typeInfo
	usedNames    map[string]int
}

type typeInfo struct {
	Elem        reflect.Type
	Ptr         reflect.Type
	ExportName  string
	FileName    string
	Aliases     map[string]string
	Methods     []*methodInfo
	NeedsUnsafe bool
}

type methodInfo struct {
	Receiver                *typeInfo
	Method                  reflect.Method
	WrapperName             string
	Params                  []paramSpec
	Returns                 []returnSpec
	IsVariadic              bool
	ReturnCGoType           string
	HeaderReturnType        string
	CSharpExternReturnType  string
	CSharpManagedReturnType string
	MultiReturnName         string
}

type paramSpec struct {
	Name              string
	GoType            reflect.Type
	CGoType           string
	HeaderType        string
	CSharpExternType  string
	CSharpManagedType string
	Kind              csharpKind
	StructType        *typeInfo
	GoArgExpr         string
}

type returnSpec struct {
	GoType              reflect.Type
	CGoType             string
	HeaderType          string
	CSharpExternType    string
	CSharpManagedType   string
	Kind                csharpKind
	StructType          *typeInfo
	ArrayElemCGoType    string
	ArrayElemCSharpType string
}

type csharpKind int

const (
	csKindPlain csharpKind = iota
	csKindBool
	csKindString
	csKindHandle
	csKindStructHandle
	csKindArray
)

type numericABI struct {
	CGoType    string
	HeaderType string
	CSharpType string
	Builtin    string
}

func GenerateCLibrary(root any, outDir string) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			switch v := recovered.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("%v", recovered)
			}
		}
	}()

	g := interopGenerator{Type: root}
	g.generate(outDir)
	return nil
}

func (g interopGenerator) generate(outDir string) {
	root := reflect.TypeOf(g.Type)
	if root == nil {
		panic("generator: Generator.Type cannot be nil")
	}

	if root.Kind() != reflect.Pointer {
		root = reflect.PointerTo(root)
	}
	if root.Kind() != reflect.Pointer || root.Elem().Kind() != reflect.Struct {
		panic("generator: Generator.Type must be a struct or pointer to struct")
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		panic(err)
	}

	if err := cleanGeneratedFiles(outDir); err != nil {
		panic(err)
	}

	state := &generatorState{
		outDir:    outDir,
		types:     map[reflect.Type]*typeInfo{},
		usedNames: map[string]int{},
	}

	state.walkType(root)
	state.sortTypes()
	state.prepareMethods()

	headerPath := filepath.Join(outDir, headerFileName)
	if err := os.WriteFile(headerPath, []byte(state.renderHeader()), 0o644); err != nil {
		panic(err)
	}

	runtimePath := filepath.Join(outDir, "runtime.go")
	if err := writeFormattedGo(runtimePath, state.renderRuntimeGo()); err != nil {
		panic(err)
	}

	for _, ti := range state.orderedTypes {
		path := filepath.Join(outDir, ti.FileName)
		if err := writeFormattedGo(path, state.renderTypeGo(ti)); err != nil {
			panic(err)
		}
	}
}

func cleanGeneratedFiles(outDir string) error {
	entries, err := os.ReadDir(outDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		path := filepath.Join(outDir, name)

		if name == headerFileName {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
			continue
		}

		ext := filepath.Ext(name)
		if ext != ".go" {
			continue
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if bytes.Contains(content, []byte(generatedBanner)) {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}

	return nil
}

func writeFormattedGo(path, source string) error {
	formatted, err := format.Source([]byte(source))
	if err != nil {
		return fmt.Errorf("format %s: %w\n%s", path, err, source)
	}
	if err := os.WriteFile(path, formatted, 0o644); err != nil {
		return err
	}
	return nil
}

func (s *generatorState) walkType(t reflect.Type) {
	if t == nil {
		return
	}

	switch t.Kind() {
	case reflect.Pointer:
		if isDiscoverableStruct(t.Elem()) {
			s.addStruct(t.Elem())
			return
		}
		s.walkType(t.Elem())
	case reflect.Struct:
		if isDiscoverableStruct(t) {
			s.addStruct(t)
		}
	case reflect.Slice, reflect.Array, reflect.Chan:
		s.walkType(t.Elem())
	case reflect.Map:
		s.walkType(t.Key())
		s.walkType(t.Elem())
	case reflect.Func:
		for i := 0; i < t.NumIn(); i++ {
			s.walkType(t.In(i))
		}
		for i := 0; i < t.NumOut(); i++ {
			s.walkType(t.Out(i))
		}
	}
}

func (s *generatorState) addStruct(st reflect.Type) {
	if !isDiscoverableStruct(st) {
		return
	}
	if _, ok := s.types[st]; ok {
		return
	}

	exportName := s.uniqueTypeName(st)
	ti := &typeInfo{
		Elem:       st,
		Ptr:        reflect.PointerTo(st),
		ExportName: exportName,
		FileName:   strings.ToLower(exportName) + ".go",
	}
	s.types[st] = ti
	s.orderedTypes = append(s.orderedTypes, ti)

	ptr := reflect.PointerTo(st)
	for i := 0; i < ptr.NumMethod(); i++ {
		m := ptr.Method(i)
		if m.PkgPath != "" {
			continue
		}
		if !methodParamsReferenceable(m.Type) {
			continue
		}
		for in := 1; in < m.Type.NumIn(); in++ {
			s.walkType(m.Type.In(in))
		}
		for out := 0; out < m.Type.NumOut(); out++ {
			s.walkType(m.Type.Out(out))
		}
	}
}

func (s *generatorState) uniqueTypeName(st reflect.Type) string {
	typePart := sanitizeExportIdentifier(st.Name())
	if typePart == "" {
		typePart = "Type"
	}
	pkgPart := sanitizeExportIdentifier(lastPathPart(st.PkgPath()))
	if pkgPart == "" {
		pkgPart = "Pkg"
	}
	base := pkgPart + "_" + typePart
	if s.usedNames[base] == 0 {
		s.usedNames[base] = 1
		return base
	}

	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s_%d", base, i)
		if s.usedNames[candidate] == 0 {
			s.usedNames[candidate] = 1
			return candidate
		}
	}
}

func (s *generatorState) sortTypes() {
	sort.Slice(s.orderedTypes, func(i, j int) bool {
		return s.orderedTypes[i].ExportName < s.orderedTypes[j].ExportName
	})
}

func (s *generatorState) prepareMethods() {
	for _, ti := range s.orderedTypes {
		ti.Aliases = buildAliasMap(ti)
		ptr := ti.Ptr
		for i := 0; i < ptr.NumMethod(); i++ {
			m := ptr.Method(i)
			if m.PkgPath != "" {
				continue
			}
			if !methodParamsReferenceable(m.Type) {
				continue
			}
			mi := s.buildMethodInfo(ti, m)
			if methodUsesUnsafe(mi) {
				ti.NeedsUnsafe = true
			}
			ti.Methods = append(ti.Methods, mi)
		}
	}
}

func methodUsesUnsafe(mi *methodInfo) bool {
	for _, out := range mi.Returns {
		if out.Kind == csKindArray {
			return true
		}
	}
	return false
}

func buildAliasMap(ti *typeInfo) map[string]string {
	pkgs := map[string]struct{}{}
	if ti.Elem.PkgPath() != "" {
		pkgs[ti.Elem.PkgPath()] = struct{}{}
	}

	ptr := ti.Ptr
	for i := 0; i < ptr.NumMethod(); i++ {
		m := ptr.Method(i)
		if m.PkgPath != "" {
			continue
		}
		if !methodParamsReferenceable(m.Type) {
			continue
		}
		for in := 1; in < m.Type.NumIn(); in++ {
			collectReferencePackages(m.Type.In(in), pkgs)
		}
	}

	pkgList := make([]string, 0, len(pkgs))
	for pkg := range pkgs {
		pkgList = append(pkgList, pkg)
	}
	sort.Strings(pkgList)

	used := map[string]struct{}{}
	aliases := map[string]string{}
	for _, pkg := range pkgList {
		alias := sanitizeLowerIdentifier(lastPathPart(pkg))
		if alias == "" {
			alias = "pkg"
		}
		if isGoKeyword(alias) {
			alias += "pkg"
		}
		base := alias
		idx := 2
		for {
			if _, exists := used[alias]; !exists {
				break
			}
			alias = fmt.Sprintf("%s%d", base, idx)
			idx++
		}
		used[alias] = struct{}{}
		aliases[pkg] = alias
	}

	return aliases
}

func collectReferencePackages(t reflect.Type, out map[string]struct{}) {
	if t == nil {
		return
	}
	if t.Name() != "" && t.PkgPath() != "" {
		out[t.PkgPath()] = struct{}{}
	}

	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array, reflect.Chan:
		collectReferencePackages(t.Elem(), out)
	case reflect.Map:
		collectReferencePackages(t.Key(), out)
		collectReferencePackages(t.Elem(), out)
	case reflect.Func:
		for i := 0; i < t.NumIn(); i++ {
			collectReferencePackages(t.In(i), out)
		}
		for i := 0; i < t.NumOut(); i++ {
			collectReferencePackages(t.Out(i), out)
		}
	}
}

func (s *generatorState) buildMethodInfo(ti *typeInfo, m reflect.Method) *methodInfo {
	mi := &methodInfo{
		Receiver:    ti,
		Method:      m,
		WrapperName: ti.ExportName + "_" + m.Name,
		IsVariadic:  m.Type.IsVariadic(),
	}

	for i := 1; i < m.Type.NumIn(); i++ {
		mi.Params = append(mi.Params, s.buildParamSpec(ti, m.Type.In(i), i))
	}
	for i := 0; i < m.Type.NumOut(); i++ {
		mi.Returns = append(mi.Returns, s.buildReturnSpec(ti, m.Type.Out(i)))
	}

	switch len(mi.Returns) {
	case 0:
		mi.ReturnCGoType = ""
		mi.HeaderReturnType = "void"
		mi.CSharpExternReturnType = "void"
		mi.CSharpManagedReturnType = "void"
	case 1:
		r := mi.Returns[0]
		mi.ReturnCGoType = r.CGoType
		mi.HeaderReturnType = r.HeaderType
		mi.CSharpExternReturnType = r.CSharpExternType
		mi.CSharpManagedReturnType = r.CSharpManagedType
	default:
		mi.MultiReturnName = mi.WrapperName + "_Result"
		mi.ReturnCGoType = "C." + mi.MultiReturnName
		mi.HeaderReturnType = mi.MultiReturnName
		mi.CSharpExternReturnType = mi.MultiReturnName
		mi.CSharpManagedReturnType = mi.MultiReturnName
	}

	return mi
}

func (s *generatorState) buildParamSpec(ti *typeInfo, t reflect.Type, idx int) paramSpec {
	name := fmt.Sprintf("p%d", idx)
	ps := paramSpec{
		Name:   name,
		GoType: t,
	}

	switch {
	case isStringType(t):
		ps.CGoType = "*C.char"
		ps.HeaderType = "char*"
		ps.CSharpExternType = "IntPtr"
		ps.CSharpManagedType = "string"
		ps.Kind = csKindString
		ps.GoArgExpr = fmt.Sprintf("C.GoString(%s)", name)
		return ps
	case isBoolType(t):
		ps.CGoType = "C.bool"
		ps.HeaderType = "bool"
		ps.CSharpExternType = "byte"
		ps.CSharpManagedType = "bool"
		ps.Kind = csKindBool
		base := fmt.Sprintf("fromCBool(%s)", name)
		if isBuiltInType(t, "bool") {
			ps.GoArgExpr = base
		} else {
			ps.GoArgExpr = fmt.Sprintf("%s(%s)", goTypeName(t, ti.Aliases), base)
		}
		return ps
	}

	if num, ok := numericTypeFor(t.Kind()); ok {
		ps.CGoType = num.CGoType
		ps.HeaderType = num.HeaderType
		ps.CSharpExternType = num.CSharpType
		ps.CSharpManagedType = num.CSharpType
		ps.Kind = csKindPlain
		ps.GoArgExpr = numericFromCExpr(ti, t, name)
		return ps
	}

	ps.CGoType = "C.uint64_t"
	ps.HeaderType = "uint64_t"
	ps.CSharpExternType = "ulong"
	ps.Kind = csKindHandle

	if st := s.structInfoForType(t); st != nil {
		ps.CSharpManagedType = st.ExportName
		ps.StructType = st
		ps.Kind = csKindStructHandle
	} else {
		ps.CSharpManagedType = "ulong"
	}

	ps.GoArgExpr = handleParamExpr(ti, t, name)
	return ps
}

func (s *generatorState) buildReturnSpec(ti *typeInfo, t reflect.Type) returnSpec {
	rs := returnSpec{GoType: t}

	switch {
	case isStringType(t):
		rs.CGoType = "*C.char"
		rs.HeaderType = "char*"
		rs.CSharpExternType = "IntPtr"
		rs.CSharpManagedType = "string"
		rs.Kind = csKindString
		return rs
	case isBoolType(t):
		rs.CGoType = "C.bool"
		rs.HeaderType = "bool"
		rs.CSharpExternType = "byte"
		rs.CSharpManagedType = "bool"
		rs.Kind = csKindBool
		return rs
	}

	if num, ok := numericTypeFor(t.Kind()); ok {
		rs.CGoType = num.CGoType
		rs.HeaderType = num.HeaderType
		rs.CSharpExternType = num.CSharpType
		rs.CSharpManagedType = num.CSharpType
		rs.Kind = csKindPlain
		return rs
	}

	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		elem := t.Elem()
		rs.CGoType = "C.GoArray"
		rs.HeaderType = "GoArray"
		rs.CSharpExternType = "GoArrayNative"
		rs.CSharpManagedType = fmt.Sprintf("GoArray<%s>", csharpArrayElementType(elem))
		rs.Kind = csKindArray
		rs.ArrayElemCGoType = cgoArrayElementType(elem)
		rs.ArrayElemCSharpType = csharpArrayElementType(elem)
		return rs
	}

	rs.CGoType = "C.uint64_t"
	rs.HeaderType = "uint64_t"
	rs.CSharpExternType = "ulong"
	rs.Kind = csKindHandle
	if st := s.structInfoForType(t); st != nil {
		rs.CSharpManagedType = st.ExportName
		rs.StructType = st
		rs.Kind = csKindStructHandle
	} else {
		rs.CSharpManagedType = "ulong"
	}

	return rs
}

func (s *generatorState) structInfoForType(t reflect.Type) *typeInfo {
	switch t.Kind() {
	case reflect.Pointer:
		elem := t.Elem()
		if isNamedStruct(elem) {
			return s.types[elem]
		}
	case reflect.Struct:
		if isNamedStruct(t) {
			return s.types[t]
		}
	}
	return nil
}

func (s *generatorState) renderRuntimeGo() string {
	var out strings.Builder
	out.WriteString(generatedBanner)
	out.WriteString("\n")
	out.WriteString("package output\n\n")
	out.WriteString("/*\n")
	out.WriteString("#include \"interop_array.h\"\n")
	out.WriteString("*/\n")
	out.WriteString("import \"C\"\n\n")
	out.WriteString("import (\n")
	out.WriteString("\t\"fmt\"\n")
	out.WriteString("\t\"sync\"\n")
	out.WriteString("\t\"sync/atomic\"\n")
	out.WriteString("\t\"unsafe\"\n")
	out.WriteString(")\n\n")
	out.WriteString("var (\n")
	out.WriteString("\thandleCounter atomic.Uint64\n")
	out.WriteString("\thandleStore   sync.Map\n")
	out.WriteString(")\n\n")
	out.WriteString("func registerObject(value any) C.uint64_t {\n")
	out.WriteString("\tif value == nil {\n")
	out.WriteString("\t\treturn 0\n")
	out.WriteString("\t}\n")
	out.WriteString("\thandle := handleCounter.Add(1)\n")
	out.WriteString("\thandleStore.Store(handle, value)\n")
	out.WriteString("\treturn C.uint64_t(handle)\n")
	out.WriteString("}\n\n")
	out.WriteString("func registerStructValue[T any](value T) C.uint64_t {\n")
	out.WriteString("\tcopyValue := value\n")
	out.WriteString("\treturn registerObject(&copyValue)\n")
	out.WriteString("}\n\n")
	out.WriteString("func mustLoadHandle[T any](handle C.uint64_t) T {\n")
	out.WriteString("\tvar zero T\n")
	out.WriteString("\tif handle == 0 {\n")
	out.WriteString("\t\treturn zero\n")
	out.WriteString("\t}\n")
	out.WriteString("\traw, ok := handleStore.Load(uint64(handle))\n")
	out.WriteString("\tif !ok {\n")
	out.WriteString("\t\tpanic(fmt.Sprintf(\"generator: unknown handle %d\", uint64(handle)))\n")
	out.WriteString("\t}\n")
	out.WriteString("\tresolved, ok := raw.(T)\n")
	out.WriteString("\tif !ok {\n")
	out.WriteString("\t\tpanic(fmt.Sprintf(\"generator: handle %d has incompatible type\", uint64(handle)))\n")
	out.WriteString("\t}\n")
	out.WriteString("\treturn resolved\n")
	out.WriteString("}\n\n")
	out.WriteString("func mustLoadStructValue[T any](handle C.uint64_t) T {\n")
	out.WriteString("\tvar zero T\n")
	out.WriteString("\tif handle == 0 {\n")
	out.WriteString("\t\treturn zero\n")
	out.WriteString("\t}\n")
	out.WriteString("\tptr := mustLoadHandle[*T](handle)\n")
	out.WriteString("\tif ptr == nil {\n")
	out.WriteString("\t\treturn zero\n")
	out.WriteString("\t}\n")
	out.WriteString("\treturn *ptr\n")
	out.WriteString("}\n\n")
	out.WriteString("func toCBool(value bool) C.bool {\n")
	out.WriteString("\treturn C.bool(value)\n")
	out.WriteString("}\n\n")
	out.WriteString("func fromCBool(value C.bool) bool {\n")
	out.WriteString("\treturn bool(value)\n")
	out.WriteString("}\n\n")
	out.WriteString("//export Go_ReleaseHandle\n")
	out.WriteString("func Go_ReleaseHandle(handle C.uint64_t) {\n")
	out.WriteString("\tif handle == 0 {\n")
	out.WriteString("\t\treturn\n")
	out.WriteString("\t}\n")
	out.WriteString("\thandleStore.Delete(uint64(handle))\n")
	out.WriteString("}\n\n")
	out.WriteString("//export C_FreeString\n")
	out.WriteString("func C_FreeString(value *C.char) {\n")
	out.WriteString("\tif value == nil {\n")
	out.WriteString("\t\treturn\n")
	out.WriteString("\t}\n")
	out.WriteString("\tC.free(unsafe.Pointer(value))\n")
	out.WriteString("}\n\n")
	out.WriteString("//export C_FreeMemory\n")
	out.WriteString("func C_FreeMemory(value unsafe.Pointer) {\n")
	out.WriteString("\tif value == nil {\n")
	out.WriteString("\t\treturn\n")
	out.WriteString("\t}\n")
	out.WriteString("\tC.free(value)\n")
	out.WriteString("}\n")

	return out.String()
}

func (s *generatorState) renderTypeGo(ti *typeInfo) string {
	var out strings.Builder
	out.WriteString(generatedBanner)
	out.WriteString("\n")
	out.WriteString("package output\n\n")
	out.WriteString("/*\n")
	out.WriteString("#include \"interop_array.h\"\n")
	out.WriteString("*/\n")
	out.WriteString("import \"C\"\n\n")

	imports := s.typeImports(ti)
	if len(imports) > 0 {
		out.WriteString("import (\n")
		for _, imp := range imports {
			if imp.alias == "" {
				out.WriteString(fmt.Sprintf("\t%q\n", imp.path))
			} else {
				out.WriteString(fmt.Sprintf("\t%s %q\n", imp.alias, imp.path))
			}
		}
		out.WriteString(")\n\n")
	}

	elemType := goTypeName(ti.Elem, ti.Aliases)
	ctor := constructorName(ti)
	out.WriteString(fmt.Sprintf("//export %s\n", ctor))
	out.WriteString(fmt.Sprintf("func %s() C.uint64_t {\n", ctor))
	out.WriteString(fmt.Sprintf("\tvalue := &%s{}\n", elemType))
	out.WriteString("\treturn registerObject(value)\n")
	out.WriteString("}\n\n")

	for _, mi := range ti.Methods {
		s.renderMethodGo(&out, ti, mi)
	}

	return out.String()
}

type importSpec struct {
	alias string
	path  string
}

func (s *generatorState) typeImports(ti *typeInfo) []importSpec {
	imports := make([]importSpec, 0, len(ti.Aliases)+1)
	if ti.NeedsUnsafe {
		imports = append(imports, importSpec{path: "unsafe"})
	}

	pkgs := make([]string, 0, len(ti.Aliases))
	for pkg := range ti.Aliases {
		pkgs = append(pkgs, pkg)
	}
	sort.Strings(pkgs)
	for _, pkg := range pkgs {
		imports = append(imports, importSpec{alias: ti.Aliases[pkg], path: pkg})
	}

	return imports
}

func (s *generatorState) renderMethodGo(out *strings.Builder, ti *typeInfo, mi *methodInfo) {
	out.WriteString(fmt.Sprintf("//export %s\n", mi.WrapperName))
	out.WriteString(fmt.Sprintf("func %s(self C.uint64_t", mi.WrapperName))
	for _, param := range mi.Params {
		out.WriteString(fmt.Sprintf(", %s %s", param.Name, param.CGoType))
	}
	out.WriteString(")")
	if mi.ReturnCGoType != "" {
		out.WriteString(" ")
		out.WriteString(mi.ReturnCGoType)
	}
	out.WriteString(" {\n")

	recvType := goTypeName(ti.Ptr, ti.Aliases)
	out.WriteString(fmt.Sprintf("\tobj := mustLoadHandle[%s](self)\n", recvType))

	callArgs := make([]string, 0, len(mi.Params))
	for _, param := range mi.Params {
		callArgs = append(callArgs, param.GoArgExpr)
	}
	if mi.IsVariadic && len(callArgs) > 0 {
		callArgs[len(callArgs)-1] = callArgs[len(callArgs)-1] + "..."
	}
	callExpr := fmt.Sprintf("obj.%s(%s)", mi.Method.Name, strings.Join(callArgs, ", "))

	switch len(mi.Returns) {
	case 0:
		out.WriteString(fmt.Sprintf("\t%s\n", callExpr))
		out.WriteString("}\n\n")
		return
	case 1:
		out.WriteString(fmt.Sprintf("\tout0 := %s\n", callExpr))
		s.renderConvertedValue(out, ti, "ret0", mi.Returns[0], "out0", "\t")
		out.WriteString("\treturn ret0\n")
		out.WriteString("}\n\n")
		return
	default:
		outVars := make([]string, len(mi.Returns))
		for i := range mi.Returns {
			outVars[i] = fmt.Sprintf("out%d", i)
		}
		out.WriteString(fmt.Sprintf("\t%s := %s\n", strings.Join(outVars, ", "), callExpr))
		for i, ret := range mi.Returns {
			varName := fmt.Sprintf("v%d", i)
			s.renderConvertedValue(out, ti, varName, ret, fmt.Sprintf("out%d", i), "\t")
		}
		out.WriteString(fmt.Sprintf("\treturn C.%s{\n", mi.MultiReturnName))
		for i := range mi.Returns {
			out.WriteString(fmt.Sprintf("\t\tv%d: v%d,\n", i, i))
		}
		out.WriteString("\t}\n")
		out.WriteString("}\n\n")
	}
}

func (s *generatorState) renderConvertedValue(out *strings.Builder, ti *typeInfo, target string, spec returnSpec, source string, indent string) {
	switch spec.Kind {
	case csKindString:
		out.WriteString(fmt.Sprintf("%s%s := C.CString(%s)\n", indent, target, source))
	case csKindBool:
		value := source
		if !isBuiltInType(spec.GoType, "bool") {
			value = fmt.Sprintf("bool(%s)", source)
		}
		out.WriteString(fmt.Sprintf("%s%s := toCBool(%s)\n", indent, target, value))
	case csKindPlain:
		out.WriteString(fmt.Sprintf("%s%s := %s\n", indent, target, numericToCExpr(spec.GoType, source)))
	case csKindStructHandle, csKindHandle:
		out.WriteString(fmt.Sprintf("%s%s := %s\n", indent, target, handleReturnExpr(spec.GoType, source)))
	case csKindArray:
		s.renderArrayConversion(out, ti, target, spec.GoType, source, indent)
	default:
		out.WriteString(fmt.Sprintf("%s%s := %s\n", indent, target, source))
	}
}

func (s *generatorState) renderArrayConversion(out *strings.Builder, ti *typeInfo, target string, arrType reflect.Type, source string, indent string) {
	sliceExpr := source
	if arrType.Kind() == reflect.Array {
		out.WriteString(fmt.Sprintf("%s%sSlice := %s[:]\n", indent, target, source))
		sliceExpr = target + "Slice"
	}

	elemType := arrType.Elem()
	cElemType := cgoArrayElementType(elemType)

	out.WriteString(fmt.Sprintf("%svar %sElemZero %s\n", indent, target, cElemType))
	out.WriteString(fmt.Sprintf("%s%sElemSize := C.int(unsafe.Sizeof(%sElemZero))\n", indent, target, target))
	out.WriteString(fmt.Sprintf("%s%s := C.GoArray{data: nil, length: 0, elementSize: %sElemSize}\n", indent, target, target))
	out.WriteString(fmt.Sprintf("%sif len(%s) > 0 {\n", indent, sliceExpr))
	out.WriteString(fmt.Sprintf("%s\ttotalSize := C.size_t(len(%s)) * C.size_t(%sElemSize)\n", indent, sliceExpr, target))
	out.WriteString(fmt.Sprintf("%s\tdata := C.malloc(totalSize)\n", indent))
	out.WriteString(fmt.Sprintf("%s\tif data != nil {\n", indent))
	out.WriteString(fmt.Sprintf("%s\t\tbuf := unsafe.Slice((*%s)(data), len(%s))\n", indent, cElemType, sliceExpr))
	out.WriteString(fmt.Sprintf("%s\t\tfor i := 0; i < len(%s); i++ {\n", indent, sliceExpr))
	out.WriteString(fmt.Sprintf("%s\t\t\tbuf[i] = %s\n", indent, arrayElementToCExpr(ti, elemType, fmt.Sprintf("%s[i]", sliceExpr))))
	out.WriteString(fmt.Sprintf("%s\t\t}\n", indent))
	out.WriteString(fmt.Sprintf("%s\t\t%s.data = data\n", indent, target))
	out.WriteString(fmt.Sprintf("%s\t\t%s.length = C.int(len(%s))\n", indent, target, sliceExpr))
	out.WriteString(fmt.Sprintf("%s\t}\n", indent))
	out.WriteString(fmt.Sprintf("%s}\n", indent))
}

func (s *generatorState) renderHeader() string {
	var out strings.Builder
	out.WriteString("#ifndef GULP_ARRAY_H\n")
	out.WriteString("#define GULP_ARRAY_H\n\n")
	out.WriteString("#include <stdbool.h>\n")
	out.WriteString("#include <stdint.h>\n")
	out.WriteString("#include <stddef.h>\n")
	out.WriteString("#include <stdlib.h>\n\n")
	out.WriteString("typedef struct {\n")
	out.WriteString("    void* data;\n")
	out.WriteString("    int length;\n")
	out.WriteString("    int elementSize;\n")
	out.WriteString("} GoArray;\n\n")

	for _, ti := range s.orderedTypes {
		for _, mi := range ti.Methods {
			if mi.MultiReturnName == "" {
				continue
			}
			out.WriteString("typedef struct {\n")
			for i, ret := range mi.Returns {
				out.WriteString(fmt.Sprintf("    %s v%d;\n", ret.HeaderType, i))
			}
			out.WriteString(fmt.Sprintf("} %s;\n\n", mi.MultiReturnName))
		}
	}

	out.WriteString("#ifdef __cplusplus\n")
	out.WriteString("extern \"C\" {\n")
	out.WriteString("#endif\n\n")

	out.WriteString("void Go_ReleaseHandle(uint64_t handle);\n")
	out.WriteString("void C_FreeString(char* value);\n")
	out.WriteString("void C_FreeMemory(void* value);\n\n")

	for _, ti := range s.orderedTypes {
		out.WriteString(fmt.Sprintf("uint64_t %s(void);\n", constructorName(ti)))
		for _, mi := range ti.Methods {
			params := []string{"uint64_t self"}
			for _, p := range mi.Params {
				params = append(params, fmt.Sprintf("%s %s", p.HeaderType, p.Name))
			}
			out.WriteString(fmt.Sprintf("%s %s(%s);\n", mi.HeaderReturnType, mi.WrapperName, strings.Join(params, ", ")))
		}
		out.WriteString("\n")
	}

	out.WriteString("#ifdef __cplusplus\n")
	out.WriteString("}\n")
	out.WriteString("#endif\n\n")
	out.WriteString("#endif\n")
	return out.String()
}

func handleParamExpr(ti *typeInfo, t reflect.Type, name string) string {
	if t.Kind() == reflect.Struct {
		return fmt.Sprintf("mustLoadStructValue[%s](%s)", goTypeName(t, ti.Aliases), name)
	}
	return fmt.Sprintf("mustLoadHandle[%s](%s)", goTypeName(t, ti.Aliases), name)
}

func handleReturnExpr(t reflect.Type, source string) string {
	if t.Kind() == reflect.Struct {
		return fmt.Sprintf("registerStructValue(%s)", source)
	}
	return fmt.Sprintf("registerObject(%s)", source)
}

func arrayElementToCExpr(ti *typeInfo, t reflect.Type, source string) string {
	switch {
	case isStringType(t):
		return fmt.Sprintf("C.CString(%s)", source)
	case isBoolType(t):
		value := source
		if !isBuiltInType(t, "bool") {
			value = fmt.Sprintf("bool(%s)", source)
		}
		return fmt.Sprintf("toCBool(%s)", value)
	}
	if _, ok := numericTypeFor(t.Kind()); ok {
		return numericToCExpr(t, source)
	}
	if t.Kind() == reflect.Struct {
		return fmt.Sprintf("registerStructValue(%s)", source)
	}
	return fmt.Sprintf("registerObject(%s)", source)
}

func cgoArrayElementType(t reflect.Type) string {
	switch {
	case isStringType(t):
		return "*C.char"
	case isBoolType(t):
		return "C.bool"
	}
	if num, ok := numericTypeFor(t.Kind()); ok {
		return num.CGoType
	}
	return "C.uint64_t"
}

func csharpArrayElementType(t reflect.Type) string {
	switch {
	case isStringType(t):
		return "IntPtr"
	case isBoolType(t):
		return "bool"
	}
	if num, ok := numericTypeFor(t.Kind()); ok {
		return num.CSharpType
	}
	return "ulong"
}

func numericTypeFor(kind reflect.Kind) (numericABI, bool) {
	switch kind {
	case reflect.Int:
		return numericABI{CGoType: "C.int", HeaderType: "int", CSharpType: "int", Builtin: "int"}, true
	case reflect.Int8:
		return numericABI{CGoType: "C.int8_t", HeaderType: "int8_t", CSharpType: "sbyte", Builtin: "int8"}, true
	case reflect.Int16:
		return numericABI{CGoType: "C.int16_t", HeaderType: "int16_t", CSharpType: "short", Builtin: "int16"}, true
	case reflect.Int32:
		return numericABI{CGoType: "C.int32_t", HeaderType: "int32_t", CSharpType: "int", Builtin: "int32"}, true
	case reflect.Int64:
		return numericABI{CGoType: "C.int64_t", HeaderType: "int64_t", CSharpType: "long", Builtin: "int64"}, true
	case reflect.Uint:
		return numericABI{CGoType: "C.uint64_t", HeaderType: "uint64_t", CSharpType: "ulong", Builtin: "uint"}, true
	case reflect.Uint8:
		return numericABI{CGoType: "C.uint8_t", HeaderType: "uint8_t", CSharpType: "byte", Builtin: "uint8"}, true
	case reflect.Uint16:
		return numericABI{CGoType: "C.uint16_t", HeaderType: "uint16_t", CSharpType: "ushort", Builtin: "uint16"}, true
	case reflect.Uint32:
		return numericABI{CGoType: "C.uint32_t", HeaderType: "uint32_t", CSharpType: "uint", Builtin: "uint32"}, true
	case reflect.Uint64:
		return numericABI{CGoType: "C.uint64_t", HeaderType: "uint64_t", CSharpType: "ulong", Builtin: "uint64"}, true
	case reflect.Uintptr:
		return numericABI{CGoType: "C.uint64_t", HeaderType: "uint64_t", CSharpType: "ulong", Builtin: "uintptr"}, true
	case reflect.Float32:
		return numericABI{CGoType: "C.float", HeaderType: "float", CSharpType: "float", Builtin: "float32"}, true
	case reflect.Float64:
		return numericABI{CGoType: "C.double", HeaderType: "double", CSharpType: "double", Builtin: "float64"}, true
	default:
		return numericABI{}, false
	}
}

func numericFromCExpr(ti *typeInfo, t reflect.Type, source string) string {
	num, ok := numericTypeFor(t.Kind())
	if !ok {
		return source
	}
	base := fmt.Sprintf("%s(%s)", num.Builtin, source)
	if isBuiltInType(t, num.Builtin) {
		return base
	}
	return fmt.Sprintf("%s(%s)", goTypeName(t, ti.Aliases), base)
}

func numericToCExpr(t reflect.Type, source string) string {
	num, ok := numericTypeFor(t.Kind())
	if !ok {
		return source
	}
	base := fmt.Sprintf("%s(%s)", num.Builtin, source)
	return fmt.Sprintf("%s(%s)", num.CGoType, base)
}

func goTypeName(t reflect.Type, aliases map[string]string) string {
	if t.Name() != "" {
		if t.PkgPath() == "" {
			return t.Name()
		}
		alias, ok := aliases[t.PkgPath()]
		if !ok || alias == "" {
			alias = sanitizeLowerIdentifier(lastPathPart(t.PkgPath()))
			if alias == "" {
				alias = "pkg"
			}
		}
		return alias + "." + t.Name()
	}

	switch t.Kind() {
	case reflect.Pointer:
		return "*" + goTypeName(t.Elem(), aliases)
	case reflect.Slice:
		return "[]" + goTypeName(t.Elem(), aliases)
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), goTypeName(t.Elem(), aliases))
	case reflect.Map:
		return fmt.Sprintf("map[%s]%s", goTypeName(t.Key(), aliases), goTypeName(t.Elem(), aliases))
	case reflect.Chan:
		dir := "chan "
		if t.ChanDir() == reflect.RecvDir {
			dir = "<-chan "
		} else if t.ChanDir() == reflect.SendDir {
			dir = "chan<- "
		}
		return dir + goTypeName(t.Elem(), aliases)
	case reflect.Func:
		var in []string
		for i := 0; i < t.NumIn(); i++ {
			paramType := t.In(i)
			if t.IsVariadic() && i == t.NumIn()-1 {
				in = append(in, "..."+goTypeName(paramType.Elem(), aliases))
			} else {
				in = append(in, goTypeName(paramType, aliases))
			}
		}
		var out []string
		for i := 0; i < t.NumOut(); i++ {
			out = append(out, goTypeName(t.Out(i), aliases))
		}
		if len(out) == 0 {
			return "func(" + strings.Join(in, ", ") + ")"
		}
		if len(out) == 1 {
			return "func(" + strings.Join(in, ", ") + ") " + out[0]
		}
		return "func(" + strings.Join(in, ", ") + ") (" + strings.Join(out, ", ") + ")"
	case reflect.Interface:
		return t.String()
	}

	return t.String()
}

func isNamedStruct(t reflect.Type) bool {
	return t != nil && t.Kind() == reflect.Struct && t.Name() != ""
}

func isDiscoverableStruct(t reflect.Type) bool {
	return isNamedStruct(t) && isExportedName(t.Name())
}

func methodParamsReferenceable(sig reflect.Type) bool {
	for i := 1; i < sig.NumIn(); i++ {
		if !typeReferenceable(sig.In(i)) {
			return false
		}
	}
	return true
}

func typeReferenceable(t reflect.Type) bool {
	if t == nil {
		return true
	}
	if t.Name() != "" && t.PkgPath() != "" && !isExportedName(t.Name()) {
		return false
	}

	switch t.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array, reflect.Chan:
		return typeReferenceable(t.Elem())
	case reflect.Map:
		return typeReferenceable(t.Key()) && typeReferenceable(t.Elem())
	case reflect.Func:
		for i := 0; i < t.NumIn(); i++ {
			if !typeReferenceable(t.In(i)) {
				return false
			}
		}
		for i := 0; i < t.NumOut(); i++ {
			if !typeReferenceable(t.Out(i)) {
				return false
			}
		}
	}

	return true
}

func isStringType(t reflect.Type) bool {
	return t.Kind() == reflect.String
}

func isBoolType(t reflect.Type) bool {
	return t.Kind() == reflect.Bool
}

func isBuiltInType(t reflect.Type, builtinName string) bool {
	return t.PkgPath() == "" && t.Name() == builtinName
}

func lastPathPart(path string) string {
	if path == "" {
		return ""
	}
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

func sanitizeExportIdentifier(input string) string {
	if input == "" {
		return ""
	}
	var out []rune
	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		return ""
	}
	if unicode.IsDigit(out[0]) {
		out = append([]rune{'X'}, out...)
	}
	out[0] = unicode.ToUpper(out[0])
	return string(out)
}

func sanitizeLowerIdentifier(input string) string {
	if input == "" {
		return ""
	}
	var out []rune
	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			out = append(out, unicode.ToLower(r))
		}
	}
	if len(out) == 0 {
		return ""
	}
	if unicode.IsDigit(out[0]) {
		out = append([]rune{'p'}, out...)
	}
	return string(out)
}

func isGoKeyword(s string) bool {
	switch s {
	case "break", "default", "func", "interface", "select", "case", "defer", "go", "map", "struct", "chan", "else", "goto", "package", "switch", "const", "fallthrough", "if", "range", "type", "continue", "for", "import", "return", "var":
		return true
	default:
		return false
	}
}

func constructorName(ti *typeInfo) string {
	return ti.ExportName + "_Ctor"
}

func isExportedName(name string) bool {
	if name == "" {
		return false
	}
	r := rune(name[0])
	return unicode.IsUpper(r)
}
