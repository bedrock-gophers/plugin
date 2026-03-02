package plugin

import "github.com/bedrock-gophers/plugin/plugin/abi"

func (m *Manager) HostCall(op uint32, payload []byte) []byte {
	d := abi.NewDecoder(payload)
	switch op {
	case hostCallBlockNames:
		return encodeStringListPayload(m.BlockNames())
	case hostCallItemNames:
		return encodeStringListPayload(m.ItemNames())
	case hostCallWorldNames:
		return encodeStringListPayload(m.WorldNames())

	case hostCallPlayerMainHandItemGet:
		playerID, ok := decodePlayerID(d)
		if !ok {
			return nil
		}
		return encodeItemStackPayload(m.PlayerMainHandItem(playerID))
	case hostCallPlayerMainHandItemSet:
		playerID, ok := decodePlayerID(d)
		if !ok {
			return encodeBoolPayload(false)
		}
		stack, ok := decodeItemStackPayload(d)
		if !ok {
			return encodeBoolPayload(false)
		}
		return encodeBoolPayload(m.SetPlayerMainHandItem(playerID, stack))
	case hostCallPlayerOffHandItemGet:
		playerID, ok := decodePlayerID(d)
		if !ok {
			return nil
		}
		return encodeItemStackPayload(m.PlayerOffHandItem(playerID))
	case hostCallPlayerOffHandItemSet:
		playerID, ok := decodePlayerID(d)
		if !ok {
			return encodeBoolPayload(false)
		}
		stack, ok := decodeItemStackPayload(d)
		if !ok {
			return encodeBoolPayload(false)
		}
		return encodeBoolPayload(m.SetPlayerOffHandItem(playerID, stack))

	case hostCallPlayerInventoryItemsGet:
		playerID, ok := decodePlayerID(d)
		if !ok {
			return nil
		}
		return encodeItemStackListPayload(m.PlayerInventoryItems(playerID))
	case hostCallPlayerInventoryItemsSet:
		playerID, ok := decodePlayerID(d)
		if !ok {
			return encodeBoolPayload(false)
		}
		values, ok := decodeItemStackListPayload(d)
		if !ok {
			return encodeBoolPayload(false)
		}
		return encodeBoolPayload(m.SetPlayerInventoryItems(playerID, values))
	case hostCallPlayerEnderChestItemsGet:
		playerID, ok := decodePlayerID(d)
		if !ok {
			return nil
		}
		return encodeItemStackListPayload(m.PlayerEnderChestItems(playerID))
	case hostCallPlayerEnderChestItemsSet:
		playerID, ok := decodePlayerID(d)
		if !ok {
			return encodeBoolPayload(false)
		}
		values, ok := decodeItemStackListPayload(d)
		if !ok {
			return encodeBoolPayload(false)
		}
		return encodeBoolPayload(m.SetPlayerEnderChestItems(playerID, values))
	case hostCallPlayerArmourItemsGet:
		playerID, ok := decodePlayerID(d)
		if !ok {
			return nil
		}
		return encodeItemStackListPayload(m.PlayerArmourItems(playerID))
	case hostCallPlayerArmourItemsSet:
		playerID, ok := decodePlayerID(d)
		if !ok {
			return encodeBoolPayload(false)
		}
		values, ok := decodeItemStackListPayload(d)
		if !ok {
			return encodeBoolPayload(false)
		}
		return encodeBoolPayload(m.SetPlayerArmourItems(playerID, values))

	case hostCallPlayerSetHeldSlot:
		playerID, ok := decodePlayerID(d)
		if !ok {
			return encodeBoolPayload(false)
		}
		slot := d.I32()
		if !d.Ok() {
			return encodeBoolPayload(false)
		}
		return encodeBoolPayload(m.SetPlayerHeldSlot(playerID, slot))
	case hostCallPlayerMoveItemsToInventory:
		playerID, ok := decodePlayerID(d)
		if !ok {
			return encodeBoolPayload(false)
		}
		return encodeBoolPayload(m.PlayerMoveItemsToInventory(playerID))
	case hostCallPlayerCloseForm:
		playerID, ok := decodePlayerID(d)
		if !ok {
			return encodeBoolPayload(false)
		}
		return encodeBoolPayload(m.PlayerCloseForm(playerID))
	case hostCallPlayerCloseDialogue:
		playerID, ok := decodePlayerID(d)
		if !ok {
			return encodeBoolPayload(false)
		}
		return encodeBoolPayload(m.PlayerCloseDialogue(playerID))
	case hostCallPlayerSendMenuForm:
		playerID, ok := decodePlayerID(d)
		if !ok {
			return encodeBoolPayload(false)
		}
		value, ok := decodeMenuFormPayload(d)
		if !ok {
			return encodeBoolPayload(false)
		}
		return encodeBoolPayload(m.PlayerSendMenuForm(playerID, value))
	case hostCallPlayerSendModalForm:
		playerID, ok := decodePlayerID(d)
		if !ok {
			return encodeBoolPayload(false)
		}
		value, ok := decodeModalFormPayload(d)
		if !ok {
			return encodeBoolPayload(false)
		}
		return encodeBoolPayload(m.PlayerSendModalForm(playerID, value))
	case hostCallPlayerLatencyMillis:
		playerID, ok := decodePlayerID(d)
		if !ok {
			return encodeI64Payload(0)
		}
		return encodeI64Payload(m.PlayerLatencyMillis(playerID))
	default:
		return nil
	}
}

func decodePlayerID(d *abi.Decoder) (uint64, bool) {
	id := d.U64()
	return id, d.Ok()
}

func encodeStringListPayload(values []string) []byte {
	enc := abi.NewEncoder(32 + len(values)*12)
	enc.U32(uint32(len(values)))
	for _, value := range values {
		enc.String(value)
	}
	return enc.Data()
}

func encodeBoolPayload(value bool) []byte {
	enc := abi.NewEncoder(1)
	enc.Bool(value)
	return enc.Data()
}

func encodeI64Payload(value int64) []byte {
	enc := abi.NewEncoder(8)
	enc.I64(value)
	return enc.Data()
}

func encodeItemStackPayload(value ItemStackData) []byte {
	enc := abi.NewEncoder(64)
	encodeItemStackData(enc, value)
	return enc.Data()
}

func decodeItemStackPayload(d *abi.Decoder) (ItemStackData, bool) {
	value := ItemStackData{
		Count:      d.I32(),
		HasItem:    d.Bool(),
		ItemName:   d.String(),
		ItemMeta:   d.I32(),
		CustomName: d.String(),
	}
	return value, d.Ok()
}

func encodeItemStackListPayload(values []ItemStackData) []byte {
	enc := abi.NewEncoder(32 + len(values)*40)
	enc.U32(uint32(len(values)))
	for _, value := range values {
		encodeItemStackData(enc, value)
	}
	return enc.Data()
}

func decodeItemStackListPayload(d *abi.Decoder) ([]ItemStackData, bool) {
	n := int(d.U32())
	if !d.Ok() || n < 0 || n > 4096 {
		return nil, false
	}
	values := make([]ItemStackData, 0, n)
	for i := 0; i < n; i++ {
		value, ok := decodeItemStackPayload(d)
		if !ok {
			return nil, false
		}
		values = append(values, value)
	}
	return values, true
}

func encodeItemStackData(enc *abi.Encoder, value ItemStackData) {
	enc.I32(value.Count)
	enc.Bool(value.HasItem)
	enc.String(value.ItemName)
	enc.I32(value.ItemMeta)
	enc.String(value.CustomName)
}

func decodeMenuFormPayload(d *abi.Decoder) (MenuFormData, bool) {
	value := MenuFormData{
		Title: d.String(),
		Body:  d.String(),
	}
	n := int(d.U32())
	if !d.Ok() || n < 0 || n > 512 {
		return MenuFormData{}, false
	}
	value.Buttons = make([]string, 0, n)
	for i := 0; i < n; i++ {
		value.Buttons = append(value.Buttons, d.String())
	}
	return value, d.Ok()
}

func decodeModalFormPayload(d *abi.Decoder) (ModalFormData, bool) {
	value := ModalFormData{
		Title:   d.String(),
		Body:    d.String(),
		Confirm: d.String(),
		Cancel:  d.String(),
	}
	return value, d.Ok()
}
