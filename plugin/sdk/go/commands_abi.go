package guest

import (
	"fmt"
	"go/ast"
	"reflect"
	"strconv"
	"strings"

	"github.com/sandertv/gophertunnel/minecraft/text"
)

// Runnable represents a Go command runnable.
// Runnables are structs with exported fields tagged with `cmd:"name"` arguments.
// The Run callback receives a lightweight command context.
type Runnable interface {
	Run(ctx Context)
}

// Allower may be implemented by a runnable to restrict command execution per source.
type Allower interface {
	Allow(source CommandSource) bool
}

// Enum may be implemented by string-backed argument types to expose a fixed set of options.
type Enum interface {
	Type() string
	Options(source CommandSource) []string
}

// SubCommand is a literal token argument that must match the parameter name.
type SubCommand struct{}

// Varargs captures all remaining command input into a single text argument.
type Varargs string

// Optional wraps an argument so it may be omitted.
// Optional fields must be at the end of the runnable.
type Optional[T any] struct {
	val T
	set bool
}

// Context provides command execution context for Go runnables.
// It intentionally does not expose tx/output from Dragonfly.
type Context struct {
	source CommandSource
	raw    []string
}

// Source returns the command source.
func (c Context) Source() CommandSource {
	return c.source
}

// Player returns the source player if the command was executed by a player.
func (c Context) Player() (PlayerRef, bool) {
	return c.source.Player()
}

// Message sends a message to the command source.
func (c Context) Message(message string) {
	c.source.Message(message)
}

// Messagef sends a formatted message to the command source.
// The format is parsed by text.Colourf.
func (c Context) Messagef(format string, a ...any) {
	c.source.Message(text.Colourf(format, a...))
}

// RawArgs returns the raw command arguments.
func (c Context) RawArgs() []string {
	return append([]string(nil), c.raw...)
}

// Load returns the parsed value and whether it was provided by the caller.
func (o Optional[T]) Load() (T, bool) {
	return o.val, o.set
}

// LoadOr returns the parsed value or a fallback when omitted.
func (o Optional[T]) LoadOr(or T) T {
	if o.set {
		return o.val
	}
	return or
}

func (o Optional[T]) with(val any) any {
	return Optional[T]{val: val.(T), set: true}
}

type optionalT interface {
	with(val any) any
}

var (
	runnableT     = reflect.TypeOf((*Runnable)(nil)).Elem()
	allowerT      = reflect.TypeOf((*Allower)(nil)).Elem()
	enumT         = reflect.TypeOf((*Enum)(nil)).Elem()
	optionalTImpl = reflect.TypeOf((*optionalT)(nil)).Elem()

	varargsT    = reflect.TypeOf(Varargs(""))
	subcommandT = reflect.TypeOf(SubCommand{})
)

type goCommandDefinition struct {
	name      string
	runnables []goRunnableDefinition
}

type goRunnableDefinition struct {
	typ      reflect.Type
	template reflect.Value
	params   []goParamDefinition
	usage    string
}

type goParamDefinition struct {
	name       string
	index      []int
	typ        reflect.Type
	optional   bool
	greedy     bool
	enum       bool
	subcommand bool

	staticEnumOptions []string
}

// RegisterCommand registers a command using the Go runnable ABI.
// It is similar to Dragonfly cmd runnables, but callbacks only receive Context.
func (baseEvents) RegisterCommand(name, description string, aliases []string, runnables ...any) {
	compiled, err := compileGoCommandDefinition(name, runnables)
	if err != nil {
		panic(fmt.Sprintf("guest.Base.RegisterCommand: %v", err))
	}
	registerCommandHandler(name, description, aliases, compiled.overloads(), compiled.dispatch)
}

func compileGoCommandDefinition(name string, runnables []any) (*goCommandDefinition, error) {
	name = normalizeCommandToken(name)
	if name == "" {
		return nil, fmt.Errorf("command name must be non-empty and contain no spaces")
	}
	if len(runnables) == 0 {
		return nil, fmt.Errorf("at least one runnable is required")
	}

	compiled := make([]goRunnableDefinition, 0, len(runnables))
	for i, r := range runnables {
		spec, err := compileGoRunnable(name, r)
		if err != nil {
			return nil, fmt.Errorf("runnable %d: %w", i, err)
		}
		compiled = append(compiled, spec)
	}

	return &goCommandDefinition{name: name, runnables: compiled}, nil
}

func compileGoRunnable(commandName string, runnable any) (goRunnableDefinition, error) {
	if runnable == nil {
		return goRunnableDefinition{}, fmt.Errorf("runnable is nil")
	}
	v := reflect.ValueOf(runnable)
	t := v.Type()
	if t.Kind() == reflect.Ptr {
		if v.IsNil() {
			return goRunnableDefinition{}, fmt.Errorf("runnable pointer is nil")
		}
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() != reflect.Struct {
		return goRunnableDefinition{}, fmt.Errorf("runnable must be a struct or pointer to struct, got %s", t.Kind())
	}
	if !t.Implements(runnableT) && !reflect.PointerTo(t).Implements(runnableT) {
		return goRunnableDefinition{}, fmt.Errorf("runnable must implement guest.Runnable")
	}

	template := reflect.New(t).Elem()
	template.Set(v)

	params := make([]goParamDefinition, 0, t.NumField())
	seenNames := map[string]struct{}{}
	fields := reflect.VisibleFields(t)
	for _, field := range fields {
		if !ast.IsExported(field.Name) || field.Anonymous {
			continue
		}
		paramName := goParamName(field)
		if paramName == "-" {
			continue
		}
		if paramName == "" {
			paramName = field.Name
		}
		paramName = normalizeArgumentName(paramName)
		if paramName == "" {
			return goRunnableDefinition{}, fmt.Errorf("invalid parameter name for field %q", field.Name)
		}
		if _, ok := seenNames[paramName]; ok {
			return goRunnableDefinition{}, fmt.Errorf("duplicate parameter name %q", paramName)
		}
		seenNames[paramName] = struct{}{}

		typ, optional := goParamType(field.Type)
		param := goParamDefinition{
			name:     paramName,
			index:    append([]int(nil), field.Index...),
			typ:      typ,
			optional: optional,
		}
		param.greedy = typ == varargsT
		param.enum = typ.Implements(enumT)
		param.subcommand = typ == subcommandT

		if param.enum && typ.Kind() != reflect.String {
			return goRunnableDefinition{}, fmt.Errorf("enum field %q must be backed by string", field.Name)
		}
		if param.greedy && param.enum {
			return goRunnableDefinition{}, fmt.Errorf("field %q cannot be both varargs and enum", field.Name)
		}
		if err := validateGoParamType(field.Name, typ); err != nil {
			return goRunnableDefinition{}, err
		}

		if param.enum {
			options := enumOptionsForType(typ, CommandSource{})
			param.staticEnumOptions = options
		}
		params = append(params, param)
	}

	optionalSeen := false
	for i, p := range params {
		if p.optional {
			optionalSeen = true
		} else if optionalSeen {
			return goRunnableDefinition{}, fmt.Errorf("required parameter %q cannot follow optional parameters", p.name)
		}
		if p.greedy && i != len(params)-1 {
			return goRunnableDefinition{}, fmt.Errorf("greedy parameter %q must be the last parameter", p.name)
		}
	}

	spec := goRunnableDefinition{
		typ:      t,
		template: template,
		params:   params,
	}
	spec.usage = goUsage(commandName, spec.params)
	return spec, nil
}

func goParamType(t reflect.Type) (reflect.Type, bool) {
	if t.Implements(optionalTImpl) && t.Kind() == reflect.Struct && t.NumField() > 0 {
		return t.Field(0).Type, true
	}
	return t, false
}

func validateGoParamType(fieldName string, t reflect.Type) error {
	switch t.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return nil
	}
	if t == varargsT || t == subcommandT || t.Implements(enumT) {
		return nil
	}
	return fmt.Errorf("unsupported parameter type %q for field %q", t.String(), fieldName)
}

func (c *goCommandDefinition) overloads() []commandOverloadSpec {
	out := make([]commandOverloadSpec, 0, len(c.runnables))
	for _, runnable := range c.runnables {
		params := make([]commandParameterSpec, 0, len(runnable.params))
		for _, p := range runnable.params {
			kind := commandParameterString
			enumOptions := []string(nil)
			switch {
			case p.subcommand:
				kind = commandParameterSubcommand
			case p.enum:
				if len(p.staticEnumOptions) > 0 {
					kind = commandParameterEnum
					enumOptions = append([]string(nil), p.staticEnumOptions...)
				}
			case p.greedy:
				kind = commandParameterText
			}
			params = append(params, commandParameterSpec{
				name:        p.name,
				kind:        kind,
				optional:    p.optional,
				enumOptions: enumOptions,
			})
		}
		out = append(out, commandOverloadSpec{parameters: params})
	}
	return out
}

func (c *goCommandDefinition) dispatch(source CommandSource, args []string) {
	ctx := Context{
		source: source,
		raw:    append([]string(nil), args...),
	}

	var (
		bestErr      error
		bestUsage    string
		bestConsumed int
		allowedAny   bool
	)
	bestConsumed = -1

	for _, runnable := range c.runnables {
		instance := reflect.New(runnable.typ).Elem()
		instance.Set(runnable.template)

		if !goAllow(instance, source) {
			continue
		}
		allowedAny = true

		consumed, err := runnable.parse(instance, source, args)
		if err != nil {
			if consumed >= bestConsumed {
				bestConsumed = consumed
				bestErr = err
				bestUsage = runnable.usage
			}
			continue
		}

		goRun(instance, ctx)
		return
	}

	if bestErr != nil {
		ctx.Message(bestErr.Error())
		if bestUsage != "" {
			ctx.Message("usage: " + bestUsage)
		}
		return
	}
	if !allowedAny {
		ctx.Message("you are not allowed to use this command")
		return
	}
	ctx.Message("unknown command usage")
}

func (r goRunnableDefinition) parse(instance reflect.Value, source CommandSource, args []string) (int, error) {
	index := 0
	for _, param := range r.params {
		if param.greedy {
			if index >= len(args) {
				if param.optional {
					continue
				}
				return index, fmt.Errorf("missing required argument %q", param.name)
			}
			raw := strings.Join(args[index:], " ")
			index = len(args)
			if err := setGoParamValue(instance, param, source, raw); err != nil {
				return index, err
			}
			continue
		}

		if index >= len(args) {
			if param.optional {
				continue
			}
			return index, fmt.Errorf("missing required argument %q", param.name)
		}

		raw := args[index]
		index++
		if err := setGoParamValue(instance, param, source, raw); err != nil {
			return index, err
		}
	}
	if index != len(args) {
		return index, fmt.Errorf("unexpected arguments: %s", strings.Join(args[index:], " "))
	}
	return index, nil
}

func setGoParamValue(instance reflect.Value, param goParamDefinition, source CommandSource, raw string) error {
	field := instance.FieldByIndex(param.index)
	dst := field
	if param.optional {
		dst = reflect.New(param.typ).Elem()
	}

	if err := parseGoParamRaw(param, source, raw, dst); err != nil {
		return err
	}

	if param.optional {
		field.Set(reflect.ValueOf(field.Interface().(optionalT).with(dst.Interface())))
	}
	return nil
}

func parseGoParamRaw(param goParamDefinition, source CommandSource, raw string, dst reflect.Value) error {
	if param.subcommand {
		if strings.EqualFold(param.name, raw) {
			return nil
		}
		return fmt.Errorf("invalid value for %q: %q (expected: %s)", param.name, raw, param.name)
	}

	if param.enum {
		matched, options := matchEnumOption(param.typ, source, raw)
		if matched == "" {
			return fmt.Errorf("invalid value for %q: %q (expected: %s)", param.name, raw, strings.Join(options, ", "))
		}
		dst.SetString(matched)
		return nil
	}

	switch dst.Kind() {
	case reflect.String:
		dst.SetString(raw)
		return nil
	case reflect.Bool:
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("invalid boolean value for %q: %q", param.name, raw)
		}
		dst.SetBool(v)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(raw, 10, dst.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid number value for %q: %q", param.name, raw)
		}
		dst.SetInt(v)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(raw, 10, dst.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid number value for %q: %q", param.name, raw)
		}
		dst.SetUint(v)
		return nil
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(raw, dst.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid number value for %q: %q", param.name, raw)
		}
		dst.SetFloat(v)
		return nil
	default:
		return fmt.Errorf("unsupported parameter type %q for %q", dst.Type().String(), param.name)
	}
}

func enumOptionsForType(t reflect.Type, source CommandSource) []string {
	zero := reflect.New(t).Elem()
	enum := zero.Interface().(Enum)
	raw := enum.Options(source)
	if len(raw) == 0 {
		return nil
	}
	out := make([]string, 0, len(raw))
	seen := map[string]struct{}{}
	for _, option := range raw {
		option = strings.TrimSpace(option)
		token := normalizeCommandToken(option)
		if token == "" || option == "" {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		out = append(out, option)
	}
	return out
}

func matchEnumOption(t reflect.Type, source CommandSource, raw string) (string, []string) {
	options := enumOptionsForType(t, source)
	for _, option := range options {
		if strings.EqualFold(option, raw) {
			return option, options
		}
	}
	return "", options
}

func goUsage(commandName string, params []goParamDefinition) string {
	parts := make([]string, 0, len(params)+1)
	parts = append(parts, "/"+commandName)
	for _, param := range params {
		if param.subcommand {
			parts = append(parts, param.name)
			continue
		}
		typ := goTypeName(param)
		label := param.name + ": " + typ
		if param.optional {
			parts = append(parts, "["+label+"]")
		} else {
			parts = append(parts, "<"+label+">")
		}
	}
	return strings.Join(parts, " ")
}

func goTypeName(param goParamDefinition) string {
	if param.greedy {
		return "text"
	}
	if param.enum {
		if len(param.staticEnumOptions) > 0 {
			return strings.Join(param.staticEnumOptions, "|")
		}
		zero := reflect.New(param.typ).Elem()
		enum := zero.Interface().(Enum)
		name := strings.TrimSpace(enum.Type())
		if name != "" {
			return name
		}
	}

	switch param.typ.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "uint"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.String:
		return "string"
	default:
		return "value"
	}
}

func goAllow(instance reflect.Value, source CommandSource) bool {
	if instance.Type().Implements(allowerT) {
		return instance.Interface().(Allower).Allow(source)
	}
	if reflect.PointerTo(instance.Type()).Implements(allowerT) {
		return instance.Addr().Interface().(Allower).Allow(source)
	}
	return true
}

func goRun(instance reflect.Value, ctx Context) {
	if instance.Type().Implements(runnableT) {
		instance.Interface().(Runnable).Run(ctx)
		return
	}
	instance.Addr().Interface().(Runnable).Run(ctx)
}

func goParamName(field reflect.StructField) string {
	tag, _ := field.Tag.Lookup("cmd")
	name, _, _ := strings.Cut(tag, ",")
	return name
}

func normalizeArgumentName(name string) string {
	return normalizeCommandToken(name)
}
