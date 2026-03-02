package guest

// CommandParameterKind describes a command argument kind that can be surfaced to the host.
type CommandParameterKind uint8

const (
	CommandParameterString CommandParameterKind = iota
	CommandParameterText
	CommandParameterEnum
	CommandParameterSubcommand
	CommandParameterPluginAvailable
	CommandParameterPluginLoaded
	CommandParameterTarget
)

// CommandParameterSpec describes one parameter in a command overload.
type CommandParameterSpec struct {
	Name        string
	Kind        CommandParameterKind
	Optional    bool
	EnumOptions []string
}

// CommandOverloadSpec describes one concrete command signature.
type CommandOverloadSpec struct {
	Parameters []CommandParameterSpec
}

func toPublicCommandOverloads(overloads []commandOverloadSpec) []CommandOverloadSpec {
	if len(overloads) == 0 {
		return nil
	}
	out := make([]CommandOverloadSpec, 0, len(overloads))
	for _, overload := range overloads {
		params := make([]CommandParameterSpec, 0, len(overload.parameters))
		for _, p := range overload.parameters {
			params = append(params, CommandParameterSpec{
				Name:        p.name,
				Kind:        CommandParameterKind(p.kind),
				Optional:    p.optional,
				EnumOptions: append([]string(nil), p.enumOptions...),
			})
		}
		out = append(out, CommandOverloadSpec{Parameters: params})
	}
	return out
}
