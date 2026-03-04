package golang

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/bedrock-gophers/plugin/internal/generator/ports"
)

func generateBridge(cfg ports.BridgeConfig) error {
	if err := generateBridgeSDK(cfg); err != nil {
		return err
	}
	if err := generateBridgeRuntime(cfg); err != nil {
		return err
	}
	return nil
}

func generateBridgeSDK(cfg ports.BridgeConfig) error {
	if err := validateBridgeBase(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.GuestHostOutput) == "" {
		return fmt.Errorf("generator: GuestHostOutput is required")
	}
	if strings.TrimSpace(cfg.RefOutput) == "" {
		return fmt.Errorf("generator: RefOutput is required")
	}
	if strings.TrimSpace(cfg.GuestHostPackage) == "" {
		return fmt.Errorf("generator: GuestHostPackage is required")
	}
	if strings.TrimSpace(cfg.RefPackage) == "" {
		return fmt.Errorf("generator: RefPackage is required")
	}
	if err := writeGeneratedGoFile(cfg.GuestHostOutput, renderGuestHost(cfg)); err != nil {
		return err
	}
	if err := writeGeneratedGoFile(cfg.RefOutput, renderRef(cfg)); err != nil {
		return err
	}
	return nil
}

func generateBridgeRuntime(cfg ports.BridgeConfig) error {
	if err := validateBridgeBase(cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.ManagerOutput) == "" {
		return fmt.Errorf("generator: ManagerOutput is required")
	}
	if strings.TrimSpace(cfg.ManagerPackage) == "" {
		return fmt.Errorf("generator: ManagerPackage is required")
	}
	if err := writeGeneratedGoFile(cfg.ManagerOutput, renderManager(cfg)); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.RuntimeOutput) == "" {
		return nil
	}
	pkg := strings.TrimSpace(cfg.RuntimePackage)
	if pkg == "" {
		return fmt.Errorf("generator: RuntimePackage is required when RuntimeOutput is set")
	}
	if !goTypeSupportedInRuntime(primaryArgType(cfg)) {
		return fmt.Errorf("generator: runtime bridge does not support PrimaryArg.Type %q", primaryArgType(cfg))
	}
	if err := writeGeneratedGoFile(cfg.RuntimeOutput, renderRuntimeBridge(cfg)); err != nil {
		return err
	}
	return nil
}

func validateBridgeBase(cfg ports.BridgeConfig) error {
	if len(cfg.Ops) == 0 {
		return fmt.Errorf("generator: bridge ops are required")
	}
	if strings.TrimSpace(primaryArgName(cfg)) == "" {
		return fmt.Errorf("generator: PrimaryArg.Name is required")
	}
	if strings.TrimSpace(primaryArgType(cfg)) == "" {
		return fmt.Errorf("generator: PrimaryArg.Type is required")
	}
	return validateBridgeOps(cfg.Ops)
}

func validateBridgeOps(ops []ports.BridgeOpSpec) error {
	seenHost := map[string]struct{}{}
	seenGuest := map[string]struct{}{}
	seenRef := map[string]struct{}{}
	seenManager := map[string]struct{}{}

	for i, op := range ops {
		id := fmt.Sprintf("ops[%d]", i)
		if strings.TrimSpace(op.HostMethod) == "" {
			return fmt.Errorf("generator: %s HostMethod is required", id)
		}
		if strings.TrimSpace(op.GuestFunc) == "" {
			return fmt.Errorf("generator: %s GuestFunc is required", id)
		}
		if strings.TrimSpace(op.RefMethod) == "" {
			return fmt.Errorf("generator: %s RefMethod is required", id)
		}
		if strings.TrimSpace(op.GuestExpr) == "" {
			return fmt.Errorf("generator: %s GuestExpr is required", id)
		}
		if strings.TrimSpace(op.RefExpr) == "" {
			return fmt.Errorf("generator: %s RefExpr is required", id)
		}
		if op.HostArg != nil {
			if strings.TrimSpace(op.HostArg.Name) == "" || strings.TrimSpace(op.HostArg.Type) == "" {
				return fmt.Errorf("generator: %s HostArg requires name and type", id)
			}
		}
		if op.GuestArg != nil {
			if strings.TrimSpace(op.GuestArg.Name) == "" || strings.TrimSpace(op.GuestArg.Type) == "" {
				return fmt.Errorf("generator: %s GuestArg requires name and type", id)
			}
		}
		if op.RefArg != nil {
			if strings.TrimSpace(op.RefArg.Name) == "" || strings.TrimSpace(op.RefArg.Type) == "" {
				return fmt.Errorf("generator: %s RefArg requires name and type", id)
			}
		}
		if op.HostArg == nil && (op.GuestArg != nil || op.RefArg != nil) {
			return fmt.Errorf("generator: %s GuestArg/RefArg provided without HostArg", id)
		}
		if op.HostArg != nil && (op.GuestArg == nil || op.RefArg == nil) {
			return fmt.Errorf("generator: %s HostArg requires GuestArg and RefArg", id)
		}
		if strings.TrimSpace(op.GuestReturn) == "" && strings.TrimSpace(op.HostReturn) != "" && op.HostReturn != "bool" {
			return fmt.Errorf("generator: %s GuestReturn missing for value op", id)
		}
		if strings.TrimSpace(op.GuestReturn) != "" && op.GuestReturn != "bool" && strings.TrimSpace(op.GuestFallback) == "" {
			return fmt.Errorf("generator: %s GuestFallback missing for GuestReturn %q", id, op.GuestReturn)
		}

		if _, ok := seenHost[op.HostMethod]; ok {
			return fmt.Errorf("generator: %s duplicate HostMethod %q", id, op.HostMethod)
		}
		seenHost[op.HostMethod] = struct{}{}

		if _, ok := seenGuest[op.GuestFunc]; ok {
			return fmt.Errorf("generator: %s duplicate GuestFunc %q", id, op.GuestFunc)
		}
		seenGuest[op.GuestFunc] = struct{}{}

		if _, ok := seenRef[op.RefMethod]; ok {
			return fmt.Errorf("generator: %s duplicate RefMethod %q", id, op.RefMethod)
		}
		seenRef[op.RefMethod] = struct{}{}

		if strings.TrimSpace(op.ManagerBody) != "" {
			if _, ok := seenManager[op.HostMethod]; ok {
				return fmt.Errorf("generator: %s duplicate manager method %q", id, op.HostMethod)
			}
			seenManager[op.HostMethod] = struct{}{}
		}
	}

	return nil
}

func writeGeneratedGoFile(path, src string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	formatted, err := format.Source([]byte(src))
	if err != nil {
		return fmt.Errorf("format %s: %w\n%s", path, err, src)
	}
	if err := os.WriteFile(path, formatted, 0o644); err != nil {
		return err
	}
	return nil
}

func renderGuestHost(cfg ports.BridgeConfig) string {
	primaryArg := bridgePrimaryArg(cfg)
	var b bytes.Buffer
	b.WriteString(generatedBanner)
	b.WriteString("\n")
	b.WriteString("package ")
	b.WriteString(cfg.GuestHostPackage)
	b.WriteString("\n\n")
	b.WriteString(renderImports(cfg.GuestHostImports))

	b.WriteString("type ")
	b.WriteString(guestHostInterfaceName(cfg))
	b.WriteString(" interface {\n")
	for _, op := range cfg.Ops {
		b.WriteString("\t")
		b.WriteString(methodSignature(op.HostMethod, withPrimaryArg(primaryArg, op.HostArg), op.HostReturn))
		b.WriteString("\n")
	}
	b.WriteString("}\n\n")

	for i, op := range cfg.Ops {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("func ")
		b.WriteString(methodSignature(op.GuestFunc, withPrimaryArg(primaryArg, op.GuestArg), op.GuestReturn))
		b.WriteString(" {\n")

		switch op.GuestReturn {
		case "":
			b.WriteString("\thostDo(func(h Host) { ")
			b.WriteString(op.GuestExpr)
			b.WriteString(" })\n")
		case "bool":
			b.WriteString("\treturn hostBool(func(h Host) bool { return ")
			b.WriteString(op.GuestExpr)
			b.WriteString(" })\n")
		default:
			b.WriteString("\treturn hostValue(")
			b.WriteString(op.GuestFallback)
			b.WriteString(", func(h Host) ")
			b.WriteString(op.GuestReturn)
			b.WriteString(" { return ")
			b.WriteString(op.GuestExpr)
			b.WriteString(" })\n")
		}

		b.WriteString("}\n")
	}

	return b.String()
}

func renderRef(cfg ports.BridgeConfig) string {
	receiverName := refReceiverName(cfg)
	var b bytes.Buffer
	b.WriteString(generatedBanner)
	b.WriteString("\n")
	b.WriteString("package ")
	b.WriteString(cfg.RefPackage)
	b.WriteString("\n\n")
	b.WriteString(renderImports(cfg.RefImports))

	for i, op := range cfg.Ops {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("func (")
		b.WriteString(receiverName)
		b.WriteString(" ")
		b.WriteString(refReceiverType(cfg))
		b.WriteString(") ")
		b.WriteString(methodSignature(op.RefMethod, maybeArg(op.RefArg), op.RefReturn))
		b.WriteString(" {\n")
		if op.RefReturn == "" {
			b.WriteString("\t")
			b.WriteString(op.RefExpr)
			b.WriteString("\n")
		} else {
			b.WriteString("\treturn ")
			b.WriteString(op.RefExpr)
			b.WriteString("\n")
		}
		b.WriteString("}\n")
	}

	return b.String()
}

func renderManager(cfg ports.BridgeConfig) string {
	primaryArg := bridgePrimaryArg(cfg)
	var b bytes.Buffer
	b.WriteString(generatedBanner)
	b.WriteString("\n")
	b.WriteString("package ")
	b.WriteString(cfg.ManagerPackage)
	b.WriteString("\n\n")
	b.WriteString(renderImports(cfg.ManagerImports))

	first := true
	for _, op := range cfg.Ops {
		if strings.TrimSpace(op.ManagerBody) == "" {
			continue
		}
		if !first {
			b.WriteString("\n")
		}
		first = false

		b.WriteString("func (m *Manager) ")
		b.WriteString(methodSignature(op.HostMethod, withPrimaryArg(primaryArg, op.HostArg), op.HostReturn))
		b.WriteString(" {\n")
		for _, line := range strings.Split(op.ManagerBody, "\n") {
			b.WriteString("\t")
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("}\n")
	}

	return b.String()
}

func renderImports(imports []string) string {
	imports = dedupeAndSortImports(imports)
	if len(imports) == 0 {
		return ""
	}
	if len(imports) == 1 {
		return "import " + imports[0] + "\n\n"
	}

	var b bytes.Buffer
	b.WriteString("import (\n")
	for _, imp := range imports {
		b.WriteString("\t")
		b.WriteString(imp)
		b.WriteString("\n")
	}
	b.WriteString(")\n\n")
	return b.String()
}

func dedupeAndSortImports(imports []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(imports))
	for _, imp := range imports {
		imp = strings.TrimSpace(imp)
		if imp == "" {
			continue
		}
		if _, ok := seen[imp]; ok {
			continue
		}
		seen[imp] = struct{}{}
		out = append(out, imp)
	}
	sort.Strings(out)
	return out
}

func renderRuntimeBridge(cfg ports.BridgeConfig) string {
	primaryArg := bridgePrimaryArg(cfg)
	supported := make([]ports.BridgeOpSpec, 0, len(cfg.Ops))
	for _, op := range cfg.Ops {
		if !opSupportedInRuntime(primaryArg.Type, op) {
			continue
		}
		supported = append(supported, op)
	}

	var b bytes.Buffer
	b.WriteString(generatedBanner)
	b.WriteString("\n")
	b.WriteString("//go:build cgo\n\n")
	b.WriteString("package ")
	b.WriteString(cfg.RuntimePackage)
	b.WriteString("\n\n")
	b.WriteString("/*\n")
	b.WriteString("#include <stdint.h>\n")
	b.WriteString("*/\n")
	b.WriteString("import \"C\"\n\n")

	for i, op := range supported {
		if i > 0 {
			b.WriteString("\n")
		}
		funcName := "csharp_" + snakeCase(op.HostMethod)
		b.WriteString("//export ")
		b.WriteString(funcName)
		b.WriteString("\n")
		b.WriteString("func ")
		b.WriteString(funcName)
		b.WriteString("(ctx C.uintptr_t, ")
		b.WriteString(primaryArg.Name)
		b.WriteString(" ")
		b.WriteString(cTypeForGoType(primaryArg.Type))
		if op.HostArg != nil {
			b.WriteString(", ")
			b.WriteString(op.HostArg.Name)
			b.WriteString(" ")
			b.WriteString(cTypeForGoType(op.HostArg.Type))
		}
		b.WriteString(")")
		retType := cReturnTypeForGoType(op.HostReturn)
		if retType != "" {
			b.WriteString(" ")
			b.WriteString(retType)
		}
		b.WriteString(" {\n")
		b.WriteString("\thostCtx, ok := CSharpHostContextByID(uintptr(ctx))\n")
		b.WriteString("\tif !ok || hostCtx.manager == nil {\n")
		if retType != "" {
			b.WriteString("\t\treturn ")
			b.WriteString(cZeroValueForGoType(op.HostReturn))
			b.WriteString("\n")
		} else {
			b.WriteString("\t\treturn\n")
		}
		b.WriteString("\t}\n")

		call := "hostCtx.manager." + op.HostMethod + "(" + cArgToGoExpr(&primaryArg)
		if op.HostArg != nil {
			call += ", " + cArgToGoExpr(op.HostArg)
		}
		call += ")"

		if retType == "" {
			b.WriteString("\t")
			b.WriteString(call)
			b.WriteString("\n")
			b.WriteString("}\n")
			continue
		}

		b.WriteString("\treturn ")
		b.WriteString(goReturnToCExpr(op.HostReturn, call))
		b.WriteString("\n")
		b.WriteString("}\n")
	}

	return b.String()
}

func opSupportedInRuntime(primaryType string, op ports.BridgeOpSpec) bool {
	if !goTypeSupportedInRuntime(primaryType) {
		return false
	}
	if !goTypeSupportedInRuntime(op.HostReturn) {
		return false
	}
	if op.HostArg != nil && !goTypeSupportedInRuntime(op.HostArg.Type) {
		return false
	}
	return true
}

func goTypeSupportedInRuntime(t string) bool {
	switch t {
	case "", "bool", "int32", "int64", "float64", "string", "uint64", "time.Duration":
		return true
	default:
		return false
	}
}

func cTypeForGoType(t string) string {
	switch t {
	case "bool":
		return "C.int"
	case "int32":
		return "C.int32_t"
	case "int64", "time.Duration":
		return "C.int64_t"
	case "float64":
		return "C.double"
	case "string":
		return "*C.char"
	case "uint64":
		return "C.uint64_t"
	default:
		panic("unsupported cgo parameter type: " + t)
	}
}

func cReturnTypeForGoType(t string) string {
	switch t {
	case "":
		return ""
	case "bool":
		return "C.int"
	case "int32":
		return "C.int32_t"
	case "int64", "time.Duration":
		return "C.int64_t"
	case "float64":
		return "C.double"
	case "string":
		return "*C.char"
	case "uint64":
		return "C.uint64_t"
	default:
		panic("unsupported cgo return type: " + t)
	}
}

func cZeroValueForGoType(t string) string {
	switch t {
	case "":
		return ""
	case "string":
		return `CString("")`
	default:
		return "0"
	}
}

func cArgToGoExpr(arg *ports.BridgeArgSpec) string {
	switch arg.Type {
	case "bool":
		return arg.Name + " != 0"
	case "int32":
		return "int32(" + arg.Name + ")"
	case "int64":
		return "int64(" + arg.Name + ")"
	case "time.Duration":
		return "time.Duration(int64(" + arg.Name + "))"
	case "float64":
		return "float64(" + arg.Name + ")"
	case "string":
		return "GoCString(" + arg.Name + ")"
	case "uint64":
		return "uint64(" + arg.Name + ")"
	default:
		panic("unsupported cgo arg type: " + arg.Type)
	}
}

func goReturnToCExpr(goType, expr string) string {
	switch goType {
	case "bool":
		return "BoolInt(" + expr + ")"
	case "int32":
		return "C.int32_t(" + expr + ")"
	case "int64":
		return "C.int64_t(" + expr + ")"
	case "time.Duration":
		return "C.int64_t(" + expr + ".Milliseconds())"
	case "float64":
		return "C.double(" + expr + ")"
	case "string":
		return "CString(" + expr + ")"
	case "uint64":
		return "C.uint64_t(" + expr + ")"
	default:
		panic("unsupported cgo return conversion type: " + goType)
	}
}

func withPrimaryArg(primary ports.BridgeArgSpec, extra *ports.BridgeArgSpec) []ports.BridgeArgSpec {
	args := []ports.BridgeArgSpec{primary}
	if extra != nil {
		args = append(args, *extra)
	}
	return args
}

func bridgePrimaryArg(cfg ports.BridgeConfig) ports.BridgeArgSpec {
	return ports.BridgeArgSpec{Name: primaryArgName(cfg), Type: primaryArgType(cfg)}
}

func primaryArgName(cfg ports.BridgeConfig) string {
	name := strings.TrimSpace(cfg.PrimaryArg.Name)
	if name == "" {
		return "subjectID"
	}
	return name
}

func primaryArgType(cfg ports.BridgeConfig) string {
	typ := strings.TrimSpace(cfg.PrimaryArg.Type)
	if typ == "" {
		return "uint64"
	}
	return typ
}

func guestHostInterfaceName(cfg ports.BridgeConfig) string {
	name := strings.TrimSpace(cfg.GuestHostInterface)
	if name == "" {
		return "bridgeHost"
	}
	return name
}

func refReceiverType(cfg ports.BridgeConfig) string {
	name := strings.TrimSpace(cfg.RefReceiverType)
	if name == "" {
		return "BridgeRef"
	}
	return name
}

func refReceiverName(cfg ports.BridgeConfig) string {
	name := strings.TrimSpace(cfg.RefReceiverName)
	if name == "" {
		return "ref"
	}
	return name
}

func maybeArg(arg *ports.BridgeArgSpec) []ports.BridgeArgSpec {
	if arg == nil {
		return nil
	}
	return []ports.BridgeArgSpec{*arg}
}

func methodSignature(name string, args []ports.BridgeArgSpec, ret string) string {
	var b strings.Builder
	b.WriteString(name)
	b.WriteString("(")
	for i, arg := range args {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(arg.Name)
		b.WriteString(" ")
		b.WriteString(arg.Type)
	}
	b.WriteString(")")
	if strings.TrimSpace(ret) != "" {
		b.WriteString(" ")
		b.WriteString(ret)
	}
	return b.String()
}

func snakeCase(in string) string {
	if in == "" {
		return ""
	}
	var out strings.Builder
	var prev rune
	for i, r := range in {
		if unicode.IsUpper(r) {
			if i > 0 && (unicode.IsLower(prev) || unicode.IsDigit(prev)) {
				out.WriteByte('_')
			}
			out.WriteRune(unicode.ToLower(r))
		} else {
			if i > 0 && unicode.IsDigit(r) && unicode.IsLetter(prev) {
				out.WriteByte('_')
			}
			out.WriteRune(r)
		}
		prev = r
	}
	return out.String()
}
