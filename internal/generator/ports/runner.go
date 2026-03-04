package ports

import "fmt"

func Run(entries []GenerationPort, profile UsageProfile) error {
	if profile == nil {
		return fmt.Errorf("generator: usage profile is nil")
	}
	for i, entry := range entries {
		if entry == nil {
			return fmt.Errorf("generator: ports[%d] is nil", i)
		}
		if !profile.Allows(entry.Usage()) {
			continue
		}
		if err := entry.Generate(); err != nil {
			return fmt.Errorf("generator: %s [%s/%s]: %w", entry.Name(), entry.Language(), entry.Usage(), err)
		}
	}
	return nil
}

func Select(entries []GenerationPort, profile UsageProfile, language Language) []GenerationPort {
	selected := make([]GenerationPort, 0, len(entries))
	for _, entry := range entries {
		if entry == nil {
			continue
		}
		if profile != nil && !profile.Allows(entry.Usage()) {
			continue
		}
		if language != "" && entry.Language() != language {
			continue
		}
		selected = append(selected, entry)
	}
	return selected
}
