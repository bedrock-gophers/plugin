package guest

import "github.com/bedrock-gophers/plugin/plugin/abi"

func ListPlugins() ([]string, error) {
	return managePlugins(abi.PluginManageList, "")
}

func LoadPlugins(target string) ([]string, error) {
	return managePlugins(abi.PluginManageLoad, target)
}

func UnloadPlugins(target string) ([]string, error) {
	return managePlugins(abi.PluginManageUnload, target)
}

func ReloadPlugins(target string) ([]string, error) {
	return managePlugins(abi.PluginManageReload, target)
}
