package guest

func managePlugins(action uint32, target string) ([]string, error) {
	host := currentHost()
	if host == nil {
		return nil, errHostUnavailable
	}
	return host.ManagePlugins(action, target)
}
