package guest

func registerCommand(name, description string, aliases []string, handlerID uint32, overloads []commandOverloadSpec) bool {
	host := currentHost()
	if host == nil {
		return false
	}
	pluginName := currentPluginRegistration()
	if pluginName == "" {
		return false
	}
	return host.RegisterCommand(pluginName, name, description, aliases, handlerID, toPublicCommandOverloads(overloads))
}
