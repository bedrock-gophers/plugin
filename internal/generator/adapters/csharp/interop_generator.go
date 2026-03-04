package csharp

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

const (
	csharpSharedFileName  = "interop_shared.g.cs"
	csharpGeneratedPrefix = "interop_"
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
	CSharpName  string
	CSharpNS    string
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

	sharedPath := filepath.Join(outDir, csharpSharedFileName)
	if err := os.WriteFile(sharedPath, []byte(state.renderCSharpShared()), 0o644); err != nil {
		panic(err)
	}
	for _, group := range state.groupTypesByPackage() {
		path := filepath.Join(outDir, interopPackageFileName(group.PackagePath))
		if err := os.WriteFile(path, []byte(state.renderCSharpPackage(group)), 0o644); err != nil {
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

		if strings.HasPrefix(name, csharpGeneratedPrefix) && strings.HasSuffix(name, ".g.cs") {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
			continue
		}
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
	csharpName := sanitizeExportIdentifier(st.Name())
	if csharpName == "" {
		csharpName = "Type"
	}
	ti := &typeInfo{
		Elem:       st,
		Ptr:        reflect.PointerTo(st),
		ExportName: exportName,
		CSharpName: csharpName,
		CSharpNS:   interopPackageNamespace(st.PkgPath()),
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
		ps.CSharpManagedType = csharpManagedTypeRef(ti, st)
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
		rs.CSharpExternType = "global::BedrockPlugin.Interop.GoArrayNative"
		rs.CSharpManagedType = fmt.Sprintf("global::BedrockPlugin.Interop.GoArray<%s>", csharpArrayElementType(elem))
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
		rs.CSharpManagedType = csharpManagedTypeRef(ti, st)
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

func csharpManagedTypeRef(current, target *typeInfo) string {
	if target == nil {
		return "ulong"
	}
	if current != nil && current.CSharpNS == target.CSharpNS {
		return target.CSharpName
	}
	return "global::" + target.CSharpNS + "." + target.CSharpName
}

type interopPackageGroup struct {
	PackagePath string
	Types       []*typeInfo
}

func (s *generatorState) groupTypesByPackage() []interopPackageGroup {
	byPkg := map[string][]*typeInfo{}
	for _, ti := range s.orderedTypes {
		pkg := ti.Elem.PkgPath()
		byPkg[pkg] = append(byPkg[pkg], ti)
	}

	pkgs := make([]string, 0, len(byPkg))
	for pkg := range byPkg {
		pkgs = append(pkgs, pkg)
	}
	sort.Strings(pkgs)

	out := make([]interopPackageGroup, 0, len(pkgs))
	for _, pkg := range pkgs {
		out = append(out, interopPackageGroup{
			PackagePath: pkg,
			Types:       byPkg[pkg],
		})
	}
	return out
}

func interopPackageFileName(pkgPath string) string {
	input := strings.ToLower(strings.TrimSpace(pkgPath))
	var b strings.Builder
	for _, r := range input {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			continue
		}
		b.WriteRune('_')
	}
	name := strings.Trim(b.String(), "_")
	if name == "" {
		name = "pkg"
	}
	return csharpGeneratedPrefix + name + ".g.cs"
}

func interopPackageNamespace(pkgPath string) string {
	const root = "BedrockPlugin.Interop"
	pkgPath = strings.TrimSpace(pkgPath)
	if pkgPath == "" {
		return root
	}

	parts := strings.Split(pkgPath, "/")
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		words := splitNonAlnumWords(part)
		if len(words) == 0 {
			continue
		}
		var segment strings.Builder
		for _, word := range words {
			id := sanitizeExportIdentifier(word)
			if id == "" {
				continue
			}
			segment.WriteString(id)
		}
		if segment.Len() == 0 {
			continue
		}
		segments = append(segments, segment.String())
	}
	if len(segments) == 0 {
		return root
	}
	return root + "." + strings.Join(segments, ".")
}

func splitNonAlnumWords(input string) []string {
	var out []string
	var current strings.Builder
	flush := func() {
		if current.Len() == 0 {
			return
		}
		out = append(out, current.String())
		current.Reset()
	}
	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	return out
}

func (s *generatorState) renderCSharpShared() string {
	var out strings.Builder
	out.WriteString("// Code generated by generator. DO NOT EDIT.\n")
	out.WriteString("using System;\n")
	out.WriteString("using System.Runtime.InteropServices;\n\n")
	out.WriteString("namespace BedrockPlugin.Interop;\n\n")
	out.WriteString("[StructLayout(LayoutKind.Sequential)]\n")
	out.WriteString("public readonly struct GoArray<T> where T : unmanaged\n")
	out.WriteString("{\n")
	out.WriteString("    public readonly IntPtr Data;\n")
	out.WriteString("    public readonly int Length;\n")
	out.WriteString("    public readonly int ElementSize;\n\n")
	out.WriteString("    public GoArray(IntPtr data, int length, int elementSize)\n")
	out.WriteString("    {\n")
	out.WriteString("        Data = data;\n")
	out.WriteString("        Length = length;\n")
	out.WriteString("        ElementSize = elementSize;\n")
	out.WriteString("    }\n\n")
	out.WriteString("    public unsafe Span<T> AsSpan()\n")
	out.WriteString("    {\n")
	out.WriteString("        if (Data == IntPtr.Zero || Length <= 0)\n")
	out.WriteString("        {\n")
	out.WriteString("            return Span<T>.Empty;\n")
	out.WriteString("        }\n")
	out.WriteString("        if (ElementSize != sizeof(T))\n")
	out.WriteString("        {\n")
	out.WriteString("            throw new InvalidOperationException($\"Element size mismatch. Native={ElementSize}, Managed={sizeof(T)}\");\n")
	out.WriteString("        }\n")
	out.WriteString("        return new Span<T>((void*)Data, Length);\n")
	out.WriteString("    }\n\n")
	out.WriteString("    public void Free() => Native.C_FreeMemory(Data);\n")
	out.WriteString("}\n\n")

	out.WriteString("[StructLayout(LayoutKind.Sequential)]\n")
	out.WriteString("public readonly struct GoArrayNative\n")
	out.WriteString("{\n")
	out.WriteString("    public readonly IntPtr Data;\n")
	out.WriteString("    public readonly int Length;\n")
	out.WriteString("    public readonly int ElementSize;\n")
	out.WriteString("}\n\n")

	out.WriteString("internal static class InteropHelpers\n")
	out.WriteString("{\n")
	out.WriteString("    internal static string ConsumeCString(IntPtr value)\n")
	out.WriteString("    {\n")
	out.WriteString("        if (value == IntPtr.Zero)\n")
	out.WriteString("        {\n")
	out.WriteString("            return string.Empty;\n")
	out.WriteString("        }\n")
	out.WriteString("        try\n")
	out.WriteString("        {\n")
	out.WriteString("            return Marshal.PtrToStringUTF8(value) ?? string.Empty;\n")
	out.WriteString("        }\n")
	out.WriteString("        finally\n")
	out.WriteString("        {\n")
	out.WriteString("            Native.C_FreeString(value);\n")
	out.WriteString("        }\n")
	out.WriteString("    }\n")
	out.WriteString("}\n\n")

	out.WriteString("internal static partial class Native\n")
	out.WriteString("{\n")
	out.WriteString("    private const string LibraryName = \"__Internal\";\n\n")
	out.WriteString("    [DllImport(LibraryName, CallingConvention = CallingConvention.Cdecl)]\n")
	out.WriteString("    internal static extern void Go_ReleaseHandle(ulong handle);\n\n")
	out.WriteString("    [DllImport(LibraryName, CallingConvention = CallingConvention.Cdecl)]\n")
	out.WriteString("    internal static extern void C_FreeString(IntPtr value);\n\n")
	out.WriteString("    [DllImport(LibraryName, CallingConvention = CallingConvention.Cdecl)]\n")
	out.WriteString("    internal static extern void C_FreeMemory(IntPtr value);\n\n")
	out.WriteString("}\n")

	return out.String()
}

func (s *generatorState) renderCSharpPackage(group interopPackageGroup) string {
	types := group.Types
	packageNS := interopPackageNamespace(group.PackagePath)

	var out strings.Builder
	out.WriteString("// Code generated by generator. DO NOT EDIT.\n")
	out.WriteString("using System;\n")
	out.WriteString("using System.Runtime.InteropServices;\n\n")
	out.WriteString("namespace BedrockPlugin.Interop\n")
	out.WriteString("{\n")

	out.WriteString("internal static partial class Native\n")
	out.WriteString("{\n")
	for _, ti := range types {
		out.WriteString("    [DllImport(LibraryName, CallingConvention = CallingConvention.Cdecl)]\n")
		out.WriteString(fmt.Sprintf("    internal static extern ulong %s();\n\n", constructorName(ti)))
		for _, mi := range ti.Methods {
			params := []string{"ulong self"}
			for _, p := range mi.Params {
				params = append(params, fmt.Sprintf("%s %s", p.CSharpExternType, p.Name))
			}
			externReturnType := mi.CSharpExternReturnType
			if mi.MultiReturnName != "" {
				externReturnType = "global::" + packageNS + "." + mi.MultiReturnName
			}
			out.WriteString("    [DllImport(LibraryName, CallingConvention = CallingConvention.Cdecl)]\n")
			out.WriteString(fmt.Sprintf("    internal static extern %s %s(%s);\n\n", externReturnType, mi.WrapperName, strings.Join(params, ", ")))
		}
	}
	out.WriteString("}\n\n")
	out.WriteString("}\n\n")

	out.WriteString("namespace ")
	out.WriteString(packageNS)
	out.WriteString("\n")
	out.WriteString("{\n")

	for _, ti := range types {
		for _, mi := range ti.Methods {
			if mi.MultiReturnName == "" {
				continue
			}
			out.WriteString("[StructLayout(LayoutKind.Sequential)]\n")
			out.WriteString(fmt.Sprintf("public struct %s\n", mi.MultiReturnName))
			out.WriteString("{\n")
			for i, ret := range mi.Returns {
				out.WriteString(fmt.Sprintf("    public %s V%d;\n", ret.CSharpExternType, i))
			}
			out.WriteString("}\n\n")
		}
	}

	for _, ti := range types {
		out.WriteString(fmt.Sprintf("public sealed class %s\n", ti.CSharpName))
		out.WriteString("{\n")
		out.WriteString("    public ulong Handle { get; }\n\n")
		out.WriteString(fmt.Sprintf("    public %s(ulong handle)\n", ti.CSharpName))
		out.WriteString("    {\n")
		out.WriteString("        Handle = handle;\n")
		out.WriteString("    }\n\n")
		out.WriteString(fmt.Sprintf("    public static %s New() => new %s(global::BedrockPlugin.Interop.Native.%s());\n\n", ti.CSharpName, ti.CSharpName, constructorName(ti)))
		out.WriteString("    public void Release() => global::BedrockPlugin.Interop.Native.Go_ReleaseHandle(Handle);\n\n")

		memberNames := newCSharpMemberNameSet(ti.CSharpName)
		for _, mi := range ti.Methods {
			methodName := nextCSharpMemberName(managedIdentifier(mi.Method.Name), memberNames)
			s.renderCSharpMethod(&out, ti, mi, methodName)
		}

		out.WriteString("}\n\n")
	}

	out.WriteString("}\n")

	return out.String()
}

func newCSharpMemberNameSet(typeName string) map[string]struct{} {
	used := map[string]struct{}{
		"Handle":  {},
		"Release": {},
		"New":     {},
	}
	if typeName != "" {
		used[typeName] = struct{}{}
	}
	return used
}

func nextCSharpMemberName(base string, used map[string]struct{}) string {
	if base == "" {
		base = "Method"
	}
	if _, exists := used[base]; !exists {
		used[base] = struct{}{}
		return base
	}
	withSuffix := base + "Method"
	if _, exists := used[withSuffix]; !exists {
		used[withSuffix] = struct{}{}
		return withSuffix
	}
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%sMethod%d", base, i)
		if _, exists := used[candidate]; exists {
			continue
		}
		used[candidate] = struct{}{}
		return candidate
	}
}

func managedIdentifier(name string) string {
	if name == "" {
		return "Value"
	}
	parts := splitManagedIdentifier(name)
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}
	return strings.Join(parts, "")
}

func splitManagedIdentifier(s string) []string {
	if s == "" {
		return nil
	}

	var parts []string
	var current bytes.Buffer
	flush := func() {
		if current.Len() == 0 {
			return
		}
		parts = append(parts, current.String())
		current.Reset()
	}

	runes := []rune(s)
	for i, r := range runes {
		if r == '_' {
			flush()
			continue
		}
		if i > 0 && isManagedIdentifierBoundary(runes[i-1], r, i+1 < len(runes) && isManagedIdentifierLower(runes[i+1])) {
			flush()
		}
		current.WriteRune(r)
	}
	flush()
	return parts
}

func isManagedIdentifierBoundary(prev, cur rune, nextIsLower bool) bool {
	if isManagedIdentifierLower(prev) && isManagedIdentifierUpper(cur) {
		return true
	}
	if isManagedIdentifierUpper(prev) && isManagedIdentifierUpper(cur) && nextIsLower {
		return true
	}
	if isManagedIdentifierDigit(prev) != isManagedIdentifierDigit(cur) {
		return true
	}
	return false
}

func isManagedIdentifierLower(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func isManagedIdentifierUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func isManagedIdentifierDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func (s *generatorState) renderCSharpMethod(out *strings.Builder, ti *typeInfo, mi *methodInfo, methodName string) {
	params := make([]string, 0, len(mi.Params))
	for _, p := range mi.Params {
		params = append(params, fmt.Sprintf("%s %s", p.CSharpManagedType, p.Name))
	}

	out.WriteString(fmt.Sprintf("    public %s %s(%s)\n", mi.CSharpManagedReturnType, methodName, strings.Join(params, ", ")))
	out.WriteString("    {\n")

	stringParams := make([]paramSpec, 0)
	for _, p := range mi.Params {
		if p.Kind == csKindString {
			stringParams = append(stringParams, p)
		}
	}

	if len(stringParams) > 0 {
		for _, sp := range stringParams {
			out.WriteString(fmt.Sprintf("        IntPtr %sNative = Marshal.StringToCoTaskMemUTF8(%s);\n", sp.Name, sp.Name))
		}
		out.WriteString("        try\n")
		out.WriteString("        {\n")
		s.renderCSharpCallBody(out, ti, mi, "            ")
		out.WriteString("        }\n")
		out.WriteString("        finally\n")
		out.WriteString("        {\n")
		for _, sp := range stringParams {
			out.WriteString(fmt.Sprintf("            if (%sNative != IntPtr.Zero)\n", sp.Name))
			out.WriteString("            {\n")
			out.WriteString(fmt.Sprintf("                Marshal.FreeCoTaskMem(%sNative);\n", sp.Name))
			out.WriteString("            }\n")
		}
		out.WriteString("        }\n")
	} else {
		s.renderCSharpCallBody(out, ti, mi, "        ")
	}

	out.WriteString("    }\n\n")
}

func (s *generatorState) renderCSharpCallBody(out *strings.Builder, ti *typeInfo, mi *methodInfo, indent string) {
	callArgs := []string{"Handle"}
	for _, p := range mi.Params {
		switch p.Kind {
		case csKindString:
			callArgs = append(callArgs, p.Name+"Native")
		case csKindBool:
			callArgs = append(callArgs, fmt.Sprintf("(byte)(%s ? 1 : 0)", p.Name))
		case csKindStructHandle:
			callArgs = append(callArgs, fmt.Sprintf("%s is null ? 0UL : %s.Handle", p.Name, p.Name))
		default:
			callArgs = append(callArgs, p.Name)
		}
	}
	callExpr := fmt.Sprintf("global::BedrockPlugin.Interop.Native.%s(%s)", mi.WrapperName, strings.Join(callArgs, ", "))

	switch len(mi.Returns) {
	case 0:
		out.WriteString(fmt.Sprintf("%s%s;\n", indent, callExpr))
		return
	case 1:
		ret := mi.Returns[0]
		out.WriteString(fmt.Sprintf("%svar result = %s;\n", indent, callExpr))
		switch ret.Kind {
		case csKindBool:
			out.WriteString(fmt.Sprintf("%sreturn result != 0;\n", indent))
		case csKindString:
			out.WriteString(fmt.Sprintf("%sreturn global::BedrockPlugin.Interop.InteropHelpers.ConsumeCString(result);\n", indent))
		case csKindStructHandle:
			out.WriteString(fmt.Sprintf("%sreturn new %s(result);\n", indent, ret.CSharpManagedType))
		case csKindArray:
			out.WriteString(fmt.Sprintf("%sreturn new global::BedrockPlugin.Interop.GoArray<%s>(result.Data, result.Length, result.ElementSize);\n", indent, ret.ArrayElemCSharpType))
		default:
			out.WriteString(fmt.Sprintf("%sreturn result;\n", indent))
		}
		return
	default:
		out.WriteString(fmt.Sprintf("%svar result = %s;\n", indent, callExpr))
		out.WriteString(fmt.Sprintf("%sreturn result;\n", indent))
	}
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
