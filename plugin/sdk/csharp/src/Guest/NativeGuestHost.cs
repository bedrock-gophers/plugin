using BedrockPlugin.Sdk.Abi;
using System.Runtime.InteropServices;
using System.Text;

namespace BedrockPlugin.Sdk.Guest;

internal unsafe sealed class NativeGuestHost : IGuestHost
{
    private readonly NativeHostApi* _api;

    internal NativeGuestHost(NativeHostApi* api)
    {
        _api = api;
        if (_api is null)
        {
            throw new ArgumentNullException(nameof(api));
        }
    }

    internal static string ReadUtf8(byte* ptr)
    {
        if (ptr is null)
        {
            return string.Empty;
        }
        return Marshal.PtrToStringUTF8((nint)ptr) ?? string.Empty;
    }

    private static byte[] Utf8Z(string value)
    {
        value ??= string.Empty;
        var byteCount = Encoding.UTF8.GetByteCount(value);
        var outBytes = new byte[byteCount + 1];
        if (byteCount > 0)
        {
            Encoding.UTF8.GetBytes(value, 0, value.Length, outBytes, 0);
        }
        outBytes[byteCount] = 0;
        return outBytes;
    }

    private string ReadAndFreeString(byte* ptr)
    {
        if (ptr is null)
        {
            return string.Empty;
        }
        try
        {
            return ReadUtf8(ptr);
        }
        finally
        {
            _api->free_string(ptr);
        }
    }

    private byte[] ReadAndFreeBytes(byte* ptr, uint len)
    {
        if (ptr is null || len == 0)
        {
            if (ptr is not null)
            {
                _api->free_bytes(ptr);
            }
            return Array.Empty<byte>();
        }

        var size = checked((int)len);
        var outBytes = new byte[size];
        Marshal.Copy((nint)ptr, outBytes, 0, size);
        _api->free_bytes(ptr);
        return outBytes;
    }

    private static string JoinAliases(IReadOnlyList<string> aliases)
    {
        if (aliases is null || aliases.Count == 0)
        {
            return string.Empty;
        }

        var outAliases = new List<string>(aliases.Count);
        for (var i = 0; i < aliases.Count; i++)
        {
            var alias = (aliases[i] ?? string.Empty).Trim();
            if (alias.Length == 0)
            {
                continue;
            }
            outAliases.Add(alias);
        }
        if (outAliases.Count == 0)
        {
            return string.Empty;
        }
        return string.Join(',', outAliases);
    }

    private static byte[] EncodeCommandOverloads(IReadOnlyList<CommandOverloadSpec> overloads)
    {
        if (overloads is null || overloads.Count == 0)
        {
            var empty = new AbiEncoder(4);
            empty.U32(0);
            return empty.Data();
        }

        var enc = new AbiEncoder(64 + overloads.Count * 24);
        enc.U32((uint)overloads.Count);
        for (var i = 0; i < overloads.Count; i++)
        {
            var parameters = overloads[i]?.Parameters ?? Array.Empty<CommandParameterSpec>();
            enc.U32((uint)parameters.Count);
            for (var j = 0; j < parameters.Count; j++)
            {
                var p = parameters[j];
                enc.String(p.Name ?? string.Empty);
                enc.U8((byte)p.Kind);
                enc.Bool(p.Optional);

                var enumOptions = p.EnumOptions ?? Array.Empty<string>();
                enc.U32((uint)enumOptions.Count);
                for (var k = 0; k < enumOptions.Count; k++)
                {
                    enc.String(enumOptions[k] ?? string.Empty);
                }
            }
        }
        return enc.Data();
    }

    private static string[] DecodeStringList(ReadOnlySpan<byte> payload)
    {
        if (payload.Length == 0)
        {
            return Array.Empty<string>();
        }

        var d = new AbiDecoder(payload);
        var countRaw = d.U32();
        if (!d.Ok || countRaw > int.MaxValue)
        {
            return Array.Empty<string>();
        }

        var count = (int)countRaw;
        if (count == 0)
        {
            return Array.Empty<string>();
        }

        var outValues = new string[count];
        for (var i = 0; i < count; i++)
        {
            outValues[i] = d.String();
        }
        if (!d.Ok)
        {
            return Array.Empty<string>();
        }
        return outValues;
    }

    private byte[] HostCall(uint op, ReadOnlySpan<byte> payload)
    {
        if (_api->host_call == null)
        {
            return Array.Empty<byte>();
        }

        uint outLen = 0;
        byte* outPtr;
        if (payload.Length == 0)
        {
            outPtr = _api->host_call(_api->ctx, op, null, 0, &outLen);
        }
        else
        {
            var inBytes = payload.ToArray();
            fixed (byte* inPtr = inBytes)
            {
                outPtr = _api->host_call(_api->ctx, op, inPtr, (uint)inBytes.Length, &outLen);
            }
        }
        return ReadAndFreeBytes(outPtr, outLen);
    }

    private static bool DecodeBool(ReadOnlySpan<byte> payload)
    {
        if (payload.Length == 0)
        {
            return false;
        }
        var d = new AbiDecoder(payload);
        var value = d.Bool();
        return d.Ok && value;
    }

    private static long DecodeI64(ReadOnlySpan<byte> payload)
    {
        if (payload.Length == 0)
        {
            return 0;
        }
        var d = new AbiDecoder(payload);
        var value = d.I64();
        return d.Ok ? value : 0;
    }

    private static byte[] EncodePlayerPayload(ulong playerId)
    {
        var enc = new AbiEncoder(8);
        enc.U64(playerId);
        return enc.Data();
    }

    private static byte[] EncodePlayerSlotPayload(ulong playerId, int slot)
    {
        var enc = new AbiEncoder(12);
        enc.U64(playerId);
        enc.I32(slot);
        return enc.Data();
    }

    private static void EncodeItemStack(AbiEncoder enc, ItemStackData value)
    {
        enc.I32(value.Count);
        enc.Bool(value.HasItem);
        enc.String(value.ItemName ?? string.Empty);
        enc.I32(value.ItemMeta);
        enc.String(value.CustomName ?? string.Empty);
    }

    private static ItemStackData DecodeItemStack(AbiDecoder d)
    {
        return new ItemStackData(
            d.I32(),
            d.Bool(),
            d.String(),
            d.I32(),
            d.String());
    }

    private static byte[] EncodeItemStackPayload(ItemStackData stack)
    {
        var enc = new AbiEncoder(64);
        EncodeItemStack(enc, stack);
        return enc.Data();
    }

    private static byte[] EncodePlayerItemPayload(ulong playerId, ItemStackData stack)
    {
        var enc = new AbiEncoder(80);
        enc.U64(playerId);
        EncodeItemStack(enc, stack);
        return enc.Data();
    }

    private static byte[] EncodePlayerItemsPayload(ulong playerId, IReadOnlyList<ItemStackData> items)
    {
        items ??= Array.Empty<ItemStackData>();
        var enc = new AbiEncoder(16 + items.Count * 48);
        enc.U64(playerId);
        enc.U32((uint)items.Count);
        for (var i = 0; i < items.Count; i++)
        {
            EncodeItemStack(enc, items[i]);
        }
        return enc.Data();
    }

    private static ItemStackData DecodeItemStackPayload(ReadOnlySpan<byte> payload)
    {
        if (payload.Length == 0)
        {
            return new ItemStackData(0, false, string.Empty, 0, string.Empty);
        }
        var d = new AbiDecoder(payload);
        var value = DecodeItemStack(d);
        if (!d.Ok)
        {
            return new ItemStackData(0, false, string.Empty, 0, string.Empty);
        }
        return value;
    }

    private static IReadOnlyList<ItemStackData> DecodeItemStacksPayload(ReadOnlySpan<byte> payload)
    {
        if (payload.Length == 0)
        {
            return Array.Empty<ItemStackData>();
        }

        var d = new AbiDecoder(payload);
        var countRaw = d.U32();
        if (!d.Ok || countRaw > int.MaxValue)
        {
            return Array.Empty<ItemStackData>();
        }
        var count = (int)countRaw;
        if (count == 0)
        {
            return Array.Empty<ItemStackData>();
        }

        var outValues = new List<ItemStackData>(count);
        for (var i = 0; i < count; i++)
        {
            outValues.Add(DecodeItemStack(d));
        }
        if (!d.Ok)
        {
            return Array.Empty<ItemStackData>();
        }
        return outValues;
    }

    private static byte[] EncodeMenuFormPayload(ulong playerId, MenuFormData value)
    {
        var buttons = value.Buttons ?? Array.Empty<string>();
        var enc = new AbiEncoder(32 + buttons.Count * 16);
        enc.U64(playerId);
        enc.String(value.Title ?? string.Empty);
        enc.String(value.Body ?? string.Empty);
        enc.U32((uint)buttons.Count);
        for (var i = 0; i < buttons.Count; i++)
        {
            enc.String(buttons[i] ?? string.Empty);
        }
        return enc.Data();
    }

    private static byte[] EncodeModalFormPayload(ulong playerId, ModalFormData value)
    {
        var enc = new AbiEncoder(96);
        enc.U64(playerId);
        enc.String(value.Title ?? string.Empty);
        enc.String(value.Body ?? string.Empty);
        enc.String(value.Confirm ?? string.Empty);
        enc.String(value.Cancel ?? string.Empty);
        return enc.Data();
    }

    public bool RegisterCommand(string pluginName, string name, string description, IReadOnlyList<string> aliases, uint handlerId, IReadOnlyList<CommandOverloadSpec> overloads)
    {
        var pluginNameZ = Utf8Z(pluginName);
        var nameZ = Utf8Z(name);
        var descriptionZ = Utf8Z(description);
        var aliasesZ = Utf8Z(JoinAliases(aliases));
        var overloadBytes = EncodeCommandOverloads(overloads ?? Array.Empty<CommandOverloadSpec>());

        fixed (byte* pluginNamePtr = pluginNameZ)
        fixed (byte* namePtr = nameZ)
        fixed (byte* descriptionPtr = descriptionZ)
        fixed (byte* aliasesPtr = aliasesZ)
        fixed (byte* overloadPtr = overloadBytes)
        {
            var payloadPtr = overloadBytes.Length == 0 ? (byte*)null : overloadPtr;
            return _api->register_command(_api->ctx, pluginNamePtr, namePtr, descriptionPtr, aliasesPtr, handlerId, payloadPtr, (uint)overloadBytes.Length) != 0;
        }
    }

    public IReadOnlyList<string> ManagePlugins(uint action, string target)
    {
        var targetZ = Utf8Z(target);
        uint outLen = 0;
        byte* payloadPtr;

        fixed (byte* targetPtr = targetZ)
        {
            payloadPtr = _api->manage_plugins(_api->ctx, action, targetPtr, &outLen);
        }

        var payload = ReadAndFreeBytes(payloadPtr, outLen);
        return DecodeStringList(payload);
    }

    public ulong ResolvePlayerByName(string name)
    {
        var nameZ = Utf8Z(name);
        fixed (byte* namePtr = nameZ)
        {
            return _api->resolve_player_by_name(_api->ctx, namePtr);
        }
    }

    public ulong PlayerHandle(ulong playerId)
    {
        if (_api->player_handle == null)
        {
            return 0;
        }
        return _api->player_handle(_api->ctx, playerId);
    }

    public IReadOnlyList<string> OnlinePlayerNames()
    {
        uint outLen = 0;
        var payloadPtr = _api->online_player_names(_api->ctx, &outLen);
        var payload = ReadAndFreeBytes(payloadPtr, outLen);
        return DecodeStringList(payload);
    }

    public IReadOnlyList<string> BlockNames()
    {
        return DecodeStringList(HostCall(HostCallOp.BlockNames, Array.Empty<byte>()));
    }

    public IReadOnlyList<string> ItemNames()
    {
        return DecodeStringList(HostCall(HostCallOp.ItemNames, Array.Empty<byte>()));
    }

    public IReadOnlyList<string> WorldNames()
    {
        return DecodeStringList(HostCall(HostCallOp.WorldNames, Array.Empty<byte>()));
    }

    public void ConsoleMessage(string pluginName, string message)
    {
        var pluginNameZ = Utf8Z(pluginName);
        var messageZ = Utf8Z(message);
        fixed (byte* pluginNamePtr = pluginNameZ)
        fixed (byte* messagePtr = messageZ)
        {
            _api->console_message(_api->ctx, pluginNamePtr, messagePtr);
        }
    }

    public bool EventCancel(ulong requestKey)
    {
        if (requestKey == 0 || _api->event_cancel == null)
        {
            return false;
        }
        return _api->event_cancel(_api->ctx, requestKey) != 0;
    }

    public bool EventSetItemDrop(ulong requestKey, ItemStackData stack)
    {
        if (requestKey == 0 || _api->event_item_drop_set == null)
        {
            return false;
        }
        var payload = EncodeItemStackPayload(stack);
        fixed (byte* payloadPtr = payload)
        {
            return _api->event_item_drop_set(_api->ctx, requestKey, payloadPtr, (uint)payload.Length) != 0;
        }
    }

    public double PlayerHealth(ulong playerId) => _api->player_health(_api->ctx, playerId);

    public bool SetPlayerHealth(ulong playerId, double health) => _api->set_player_health(_api->ctx, playerId, health) != 0;

    public int PlayerFood(ulong playerId) => _api->player_food(_api->ctx, playerId);

    public bool SetPlayerFood(ulong playerId, int food) => _api->set_player_food(_api->ctx, playerId, food) != 0;

    public string PlayerName(ulong playerId) => ReadAndFreeString(_api->player_name(_api->ctx, playerId));

    public int PlayerGameMode(ulong playerId) => _api->player_game_mode(_api->ctx, playerId);

    public bool SetPlayerGameMode(ulong playerId, int mode) => _api->set_player_game_mode(_api->ctx, playerId, mode) != 0;

    public string PlayerXUID(ulong playerId) => ReadAndFreeString(_api->player_xuid(_api->ctx, playerId));

    public string PlayerDeviceID(ulong playerId) => ReadAndFreeString(_api->player_device_id(_api->ctx, playerId));

    public string PlayerDeviceModel(ulong playerId) => ReadAndFreeString(_api->player_device_model(_api->ctx, playerId));

    public string PlayerSelfSignedID(ulong playerId) => ReadAndFreeString(_api->player_self_signed_id(_api->ctx, playerId));

    public string PlayerNameTag(ulong playerId) => ReadAndFreeString(_api->player_name_tag(_api->ctx, playerId));

    public bool SetPlayerNameTag(ulong playerId, string value)
    {
        var valueZ = Utf8Z(value);
        fixed (byte* valuePtr = valueZ)
        {
            return _api->set_player_name_tag(_api->ctx, playerId, valuePtr) != 0;
        }
    }

    public string PlayerScoreTag(ulong playerId) => ReadAndFreeString(_api->player_score_tag(_api->ctx, playerId));

    public bool SetPlayerScoreTag(ulong playerId, string value)
    {
        var valueZ = Utf8Z(value);
        fixed (byte* valuePtr = valueZ)
        {
            return _api->set_player_score_tag(_api->ctx, playerId, valuePtr) != 0;
        }
    }

    public double PlayerAbsorption(ulong playerId) => _api->player_absorption(_api->ctx, playerId);

    public bool SetPlayerAbsorption(ulong playerId, double value) => _api->set_player_absorption(_api->ctx, playerId, value) != 0;

    public double PlayerMaxHealth(ulong playerId) => _api->player_max_health(_api->ctx, playerId);

    public bool SetPlayerMaxHealth(ulong playerId, double value) => _api->set_player_max_health(_api->ctx, playerId, value) != 0;

    public double PlayerSpeed(ulong playerId) => _api->player_speed(_api->ctx, playerId);

    public bool SetPlayerSpeed(ulong playerId, double value) => _api->set_player_speed(_api->ctx, playerId, value) != 0;

    public double PlayerFlightSpeed(ulong playerId) => _api->player_flight_speed(_api->ctx, playerId);

    public bool SetPlayerFlightSpeed(ulong playerId, double value) => _api->set_player_flight_speed(_api->ctx, playerId, value) != 0;

    public double PlayerVerticalFlightSpeed(ulong playerId) => _api->player_vertical_flight_speed(_api->ctx, playerId);

    public bool SetPlayerVerticalFlightSpeed(ulong playerId, double value) => _api->set_player_vertical_flight_speed(_api->ctx, playerId, value) != 0;

    public int PlayerExperience(ulong playerId) => _api->player_experience(_api->ctx, playerId);

    public int PlayerExperienceLevel(ulong playerId) => _api->player_experience_level(_api->ctx, playerId);

    public bool SetPlayerExperienceLevel(ulong playerId, int value) => _api->set_player_experience_level(_api->ctx, playerId, value) != 0;

    public double PlayerExperienceProgress(ulong playerId) => _api->player_experience_progress(_api->ctx, playerId);

    public bool SetPlayerExperienceProgress(ulong playerId, double value) => _api->set_player_experience_progress(_api->ctx, playerId, value) != 0;

    public bool PlayerOnGround(ulong playerId) => _api->player_on_ground(_api->ctx, playerId) != 0;

    public bool PlayerSneaking(ulong playerId) => _api->player_sneaking(_api->ctx, playerId) != 0;

    public bool SetPlayerSneaking(ulong playerId, bool value) => _api->set_player_sneaking(_api->ctx, playerId, value ? 1 : 0) != 0;

    public bool PlayerSprinting(ulong playerId) => _api->player_sprinting(_api->ctx, playerId) != 0;

    public bool SetPlayerSprinting(ulong playerId, bool value) => _api->set_player_sprinting(_api->ctx, playerId, value ? 1 : 0) != 0;

    public bool PlayerSwimming(ulong playerId) => _api->player_swimming(_api->ctx, playerId) != 0;

    public bool SetPlayerSwimming(ulong playerId, bool value) => _api->set_player_swimming(_api->ctx, playerId, value ? 1 : 0) != 0;

    public bool PlayerFlying(ulong playerId) => _api->player_flying(_api->ctx, playerId) != 0;

    public bool SetPlayerFlying(ulong playerId, bool value) => _api->set_player_flying(_api->ctx, playerId, value ? 1 : 0) != 0;

    public bool PlayerGliding(ulong playerId) => _api->player_gliding(_api->ctx, playerId) != 0;

    public bool SetPlayerGliding(ulong playerId, bool value) => _api->set_player_gliding(_api->ctx, playerId, value ? 1 : 0) != 0;

    public bool PlayerCrawling(ulong playerId) => _api->player_crawling(_api->ctx, playerId) != 0;

    public bool SetPlayerCrawling(ulong playerId, bool value) => _api->set_player_crawling(_api->ctx, playerId, value ? 1 : 0) != 0;

    public bool PlayerUsingItem(ulong playerId) => _api->player_using_item(_api->ctx, playerId) != 0;

    public bool PlayerInvisible(ulong playerId) => _api->player_invisible(_api->ctx, playerId) != 0;

    public bool SetPlayerInvisible(ulong playerId, bool value) => _api->set_player_invisible(_api->ctx, playerId, value ? 1 : 0) != 0;

    public bool PlayerImmobile(ulong playerId) => _api->player_immobile(_api->ctx, playerId) != 0;

    public bool SetPlayerImmobile(ulong playerId, bool value) => _api->set_player_immobile(_api->ctx, playerId, value ? 1 : 0) != 0;

    public bool PlayerDead(ulong playerId) => _api->player_dead(_api->ctx, playerId) != 0;

    public TimeSpan PlayerLatency(ulong playerId) => TimeSpan.FromMilliseconds(System.Math.Max(0L, _api->player_latency(_api->ctx, playerId)));

    public bool SetPlayerOnFireMillis(ulong playerId, long millis) => _api->set_player_on_fire_millis(_api->ctx, playerId, millis) != 0;

    public bool AddPlayerFood(ulong playerId, int points) => _api->add_player_food(_api->ctx, playerId, points) != 0;

    public bool PlayerUseItem(ulong playerId) => _api->player_use_item(_api->ctx, playerId) != 0;

    public bool PlayerJump(ulong playerId) => _api->player_jump(_api->ctx, playerId) != 0;

    public bool PlayerSwingArm(ulong playerId) => _api->player_swing_arm(_api->ctx, playerId) != 0;

    public bool PlayerWake(ulong playerId) => _api->player_wake(_api->ctx, playerId) != 0;

    public bool PlayerExtinguish(ulong playerId) => _api->player_extinguish(_api->ctx, playerId) != 0;

    public bool SetPlayerShowCoordinates(ulong playerId, bool value) => _api->set_player_show_coordinates(_api->ctx, playerId, value ? 1 : 0) != 0;

    public ItemStackData PlayerMainHandItem(ulong playerId)
    {
        return DecodeItemStackPayload(HostCall(HostCallOp.PlayerMainHandItemGet, EncodePlayerPayload(playerId)));
    }

    public bool SetPlayerMainHandItem(ulong playerId, ItemStackData stack)
    {
        return DecodeBool(HostCall(HostCallOp.PlayerMainHandItemSet, EncodePlayerItemPayload(playerId, stack)));
    }

    public ItemStackData PlayerOffHandItem(ulong playerId)
    {
        return DecodeItemStackPayload(HostCall(HostCallOp.PlayerOffHandItemGet, EncodePlayerPayload(playerId)));
    }

    public bool SetPlayerOffHandItem(ulong playerId, ItemStackData stack)
    {
        return DecodeBool(HostCall(HostCallOp.PlayerOffHandItemSet, EncodePlayerItemPayload(playerId, stack)));
    }

    public IReadOnlyList<ItemStackData> PlayerInventoryItems(ulong playerId)
    {
        return DecodeItemStacksPayload(HostCall(HostCallOp.PlayerInventoryItemsGet, EncodePlayerPayload(playerId)));
    }

    public bool SetPlayerInventoryItems(ulong playerId, IReadOnlyList<ItemStackData> items)
    {
        return DecodeBool(HostCall(HostCallOp.PlayerInventoryItemsSet, EncodePlayerItemsPayload(playerId, items)));
    }

    public IReadOnlyList<ItemStackData> PlayerEnderChestItems(ulong playerId)
    {
        return DecodeItemStacksPayload(HostCall(HostCallOp.PlayerEnderChestItemsGet, EncodePlayerPayload(playerId)));
    }

    public bool SetPlayerEnderChestItems(ulong playerId, IReadOnlyList<ItemStackData> items)
    {
        return DecodeBool(HostCall(HostCallOp.PlayerEnderChestItemsSet, EncodePlayerItemsPayload(playerId, items)));
    }

    public IReadOnlyList<ItemStackData> PlayerArmourItems(ulong playerId)
    {
        return DecodeItemStacksPayload(HostCall(HostCallOp.PlayerArmourItemsGet, EncodePlayerPayload(playerId)));
    }

    public bool SetPlayerArmourItems(ulong playerId, IReadOnlyList<ItemStackData> items)
    {
        return DecodeBool(HostCall(HostCallOp.PlayerArmourItemsSet, EncodePlayerItemsPayload(playerId, items)));
    }

    public bool SetPlayerHeldSlot(ulong playerId, int slot)
    {
        return DecodeBool(HostCall(HostCallOp.PlayerSetHeldSlot, EncodePlayerSlotPayload(playerId, slot)));
    }

    public bool PlayerMoveItemsToInventory(ulong playerId)
    {
        return DecodeBool(HostCall(HostCallOp.PlayerMoveItemsToInventory, EncodePlayerPayload(playerId)));
    }

    public bool PlayerCloseForm(ulong playerId)
    {
        return DecodeBool(HostCall(HostCallOp.PlayerCloseForm, EncodePlayerPayload(playerId)));
    }

    public bool PlayerCloseDialogue(ulong playerId)
    {
        return DecodeBool(HostCall(HostCallOp.PlayerCloseDialogue, EncodePlayerPayload(playerId)));
    }

    public bool PlayerSendMenuForm(ulong playerId, MenuFormData value)
    {
        return DecodeBool(HostCall(HostCallOp.PlayerSendMenuForm, EncodeMenuFormPayload(playerId, value)));
    }

    public bool PlayerSendModalForm(ulong playerId, ModalFormData value)
    {
        return DecodeBool(HostCall(HostCallOp.PlayerSendModalForm, EncodeModalFormPayload(playerId, value)));
    }

    public void PlayerMessage(ulong playerId, string message)
    {
        var messageZ = Utf8Z(message);
        fixed (byte* messagePtr = messageZ)
        {
            _api->player_message(_api->ctx, playerId, messagePtr);
        }
    }
}
