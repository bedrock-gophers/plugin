package plugin

import (
	"sort"
	"strings"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/df-mc/dragonfly/server/world"
)

func (m *Manager) BlockNames() []string {
	blocks := world.Blocks()
	values := make([]string, 0, len(blocks))
	for _, b := range blocks {
		if b == nil {
			continue
		}
		name, _ := b.EncodeBlock()
		values = append(values, name)
	}
	return uniqueSorted(values)
}

func (m *Manager) ItemNames() []string {
	items := world.Items()
	values := make([]string, 0, len(items))
	for _, it := range items {
		if it == nil {
			continue
		}
		name, _ := it.EncodeItem()
		values = append(values, name)
	}
	return uniqueSorted(values)
}

func (m *Manager) WorldNames() []string {
	if m == nil || m.srv == nil {
		return nil
	}
	values := []string{"overworld", "nether", "end"}
	if w := m.srv.World(); w != nil {
		values = append(values, w.Name())
	}
	if w := m.srv.Nether(); w != nil {
		values = append(values, w.Name())
	}
	if w := m.srv.End(); w != nil {
		values = append(values, w.Name())
	}
	return uniqueSorted(values)
}

func uniqueSorted(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, raw := range values {
		v := strings.TrimSpace(raw)
		if v == "" {
			continue
		}
		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func mainHandItemData(p *player.Player) ItemStackData {
	main, _ := p.HeldItems()
	return itemStackData(main)
}

func offHandItemData(p *player.Player) ItemStackData {
	_, off := p.HeldItems()
	return itemStackData(off)
}

func setMainHandItemData(p *player.Player, value ItemStackData) bool {
	main, ok := itemStackFromData(value)
	if !ok {
		return false
	}
	_, off := p.HeldItems()
	p.SetHeldItems(main, off)
	return true
}

func setOffHandItemData(p *player.Player, value ItemStackData) bool {
	off, ok := itemStackFromData(value)
	if !ok {
		return false
	}
	main, _ := p.HeldItems()
	p.SetHeldItems(main, off)
	return true
}

func inventoryItemsData(inv *inventory.Inventory) []ItemStackData {
	if inv == nil {
		return nil
	}
	slots := inv.Slots()
	out := make([]ItemStackData, 0, len(slots))
	for _, stack := range slots {
		out = append(out, itemStackData(stack))
	}
	return out
}

func armourItemsData(armour *inventory.Armour) []ItemStackData {
	if armour == nil {
		return nil
	}
	slots := armour.Slots()
	out := make([]ItemStackData, 0, len(slots))
	for _, stack := range slots {
		out = append(out, itemStackData(stack))
	}
	return out
}

func setInventoryItemsData(inv *inventory.Inventory, values []ItemStackData) bool {
	if inv == nil {
		return false
	}
	size := inv.Size()
	stacks := make([]item.Stack, size)
	limit := min(size, len(values))
	for i := 0; i < limit; i++ {
		stack, ok := itemStackFromData(values[i])
		if !ok {
			return false
		}
		stacks[i] = stack
	}
	for i := 0; i < size; i++ {
		if err := inv.SetItem(i, stacks[i]); err != nil {
			return false
		}
	}
	return true
}

func setArmourItemsData(armour *inventory.Armour, values []ItemStackData) bool {
	if armour == nil {
		return false
	}
	return setInventoryItemsData(armour.Inventory(), values)
}

func itemStackData(stack item.Stack) ItemStackData {
	value := ItemStackData{
		Count:      int32(stack.Count()),
		HasItem:    false,
		ItemName:   "",
		ItemMeta:   0,
		CustomName: stack.CustomName(),
	}
	if it := stack.Item(); it != nil {
		name, meta := it.EncodeItem()
		value.HasItem = true
		value.ItemName = name
		value.ItemMeta = int32(meta)
	}
	return value
}

func itemStackFromData(value ItemStackData) (item.Stack, bool) {
	if !value.HasItem {
		return item.Stack{}, true
	}

	name := strings.TrimSpace(value.ItemName)
	if name == "" {
		return item.Stack{}, false
	}
	if value.ItemMeta < -32768 || value.ItemMeta > 32767 {
		return item.Stack{}, false
	}
	if value.Count <= 0 {
		return item.Stack{}, true
	}

	it, ok := world.ItemByName(name, int16(value.ItemMeta))
	if !ok || it == nil {
		return item.Stack{}, false
	}

	stack := item.NewStack(it, int(value.Count))
	if value.CustomName != "" {
		stack = stack.WithCustomName(value.CustomName)
	}
	return stack, true
}

type noOpMenuSubmittable struct{}

func (noOpMenuSubmittable) Submit(form.Submitter, form.Button, *world.Tx) {}

type noOpModalSubmittable struct {
	Confirm form.Button
	Cancel  form.Button
}

func (noOpModalSubmittable) Submit(form.Submitter, form.Button, *world.Tx) {}

func sendMenuFormData(p *player.Player, value MenuFormData) bool {
	buttons := make([]string, 0, len(value.Buttons))
	for _, button := range value.Buttons {
		label := strings.TrimSpace(button)
		if label == "" {
			continue
		}
		buttons = append(buttons, label)
	}
	if len(buttons) == 0 {
		return false
	}

	menu := form.NewMenu(noOpMenuSubmittable{}, value.Title).WithBody(value.Body)
	for _, button := range buttons {
		menu = menu.AddButton(form.NewButton(button, ""))
	}
	p.SendForm(menu)
	return true
}

func sendModalFormData(p *player.Player, value ModalFormData) bool {
	confirm := strings.TrimSpace(value.Confirm)
	cancel := strings.TrimSpace(value.Cancel)
	if confirm == "" {
		confirm = "gui.yes"
	}
	if cancel == "" {
		cancel = "gui.no"
	}

	modal := form.NewModal(noOpModalSubmittable{
		Confirm: form.NewButton(confirm, ""),
		Cancel:  form.NewButton(cancel, ""),
	}, value.Title).WithBody(value.Body)
	p.SendForm(modal)
	return true
}
