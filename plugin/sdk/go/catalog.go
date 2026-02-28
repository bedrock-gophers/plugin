package guest

// BlockNames returns all host-registered block names.
func BlockNames() []string {
	return hostValue([]string(nil), func(h Host) []string {
		values := h.BlockNames()
		return append([]string(nil), values...)
	})
}

// ItemNames returns all host-registered item names.
func ItemNames() []string {
	return hostValue([]string(nil), func(h Host) []string {
		values := h.ItemNames()
		return append([]string(nil), values...)
	})
}

// WorldNames returns all world names known by the host.
func WorldNames() []string {
	return hostValue([]string(nil), func(h Host) []string {
		values := h.WorldNames()
		return append([]string(nil), values...)
	})
}

// Block is a built-in command enum for block names.
type Block string

func (Block) Type() string {
	return "block"
}

func (Block) Options(_ CommandSource) []string {
	return BlockNames()
}

// Item is a built-in command enum for item names.
type Item string

func (Item) Type() string {
	return "item"
}

func (Item) Options(_ CommandSource) []string {
	return ItemNames()
}

// World is a built-in command enum for world names.
type World string

func (World) Type() string {
	return "world"
}

func (World) Options(_ CommandSource) []string {
	return WorldNames()
}
