using BedrockPlugin.Sdk.Abi;

namespace BedrockPlugin.Sdk.Guest;

public static class EventPayloadDecoder
{
    public static bool TryDecode(ushort eventId, ReadOnlySpan<byte> payload, out IEventPayload? decoded, out string? error)
    {
        try
        {
            decoded = Decode(eventId, payload);
            error = null;
            return true;
        }
        catch (Exception ex)
        {
            decoded = null;
            error = ex.Message;
            return false;
        }
    }

    public static IEventPayload Decode(ushort eventId, ReadOnlySpan<byte> payload)
    {
        var d = new AbiDecoder(payload);

        IEventPayload decoded = eventId switch
        {
            EventIds.EventMove => new MovePayload(DecodeVec3(d), DecodeRotation(d)),
            EventIds.EventJump => new JumpPayload(DecodePlayerIdentity(d)),
            EventIds.EventTeleport => new TeleportPayload(DecodeVec3(d)),
            EventIds.EventChangeWorld => new ChangeWorldPayload(DecodeWorld(d), DecodeWorld(d)),
            EventIds.EventToggleSprint => new ToggleSprintPayload(d.Bool()),
            EventIds.EventToggleSneak => new ToggleSneakPayload(d.Bool()),
            EventIds.EventChat => new ChatPayload(d.String()),
            EventIds.EventFoodLoss => new FoodLossPayload(d.I32(), d.I32()),
            EventIds.EventHeal => new HealPayload(d.F64(), DecodeHealingSource(d)),
            EventIds.EventHurt => new HurtPayload(d.F64(), d.Bool(), d.I64(), DecodeDamageSource(d)),
            EventIds.EventDeath => new DeathPayload(DecodeDamageSource(d), d.Bool()),
            EventIds.EventRespawn => new RespawnPayload(DecodeVec3(d), d.String()),
            EventIds.EventSkinChange => new SkinChangePayload(DecodeSkin(d)),
            EventIds.EventFireExtinguish => new FireExtinguishPayload(DecodePos(d)),
            EventIds.EventStartBreak => new StartBreakPayload(DecodePos(d)),
            EventIds.EventBlockBreak => new BlockBreakPayload(DecodePos(d), DecodeItemStacks(d), d.I32()),
            EventIds.EventBlockPlace => new BlockPlacePayload(DecodePos(d), DecodeBlock(d)),
            EventIds.EventBlockPick => new BlockPickPayload(DecodePos(d), DecodeBlock(d)),
            EventIds.EventItemUse => new ItemUsePayload(),
            EventIds.EventItemUseOnBlock => new ItemUseOnBlockPayload(DecodePos(d), d.U8(), DecodeVec3(d)),
            EventIds.EventItemUseOnEntity => new ItemUseOnEntityPayload(DecodeEntity(d)),
            EventIds.EventItemRelease => new ItemReleasePayload(DecodeItemStack(d), d.I64()),
            EventIds.EventItemConsume => new ItemConsumePayload(DecodeItemStack(d)),
            EventIds.EventAttackEntity => new AttackEntityPayload(DecodeEntity(d), d.F64(), d.F64(), d.Bool()),
            EventIds.EventExperienceGain => new ExperienceGainPayload(d.I32()),
            EventIds.EventPunchAir => new PunchAirPayload(),
            EventIds.EventSignEdit => new SignEditPayload(DecodePos(d), d.Bool(), d.String(), d.String()),
            EventIds.EventSleep => new SleepPayload(d.Bool()),
            EventIds.EventLecternPageTurn => new LecternPageTurnPayload(DecodePos(d), d.I32(), d.I32()),
            EventIds.EventItemDamage => new ItemDamagePayload(DecodeItemStack(d), d.I32()),
            EventIds.EventItemPickup => new ItemPickupPayload(DecodeItemStack(d)),
            EventIds.EventHeldSlotChange => new HeldSlotChangePayload(d.I32(), d.I32()),
            EventIds.EventItemDrop => new ItemDropPayload(DecodeItemStack(d)),
            EventIds.EventTransfer => new TransferPayload(d.String()),
            EventIds.EventCommandExecution => new CommandExecutionPayload(DecodeCommand(d)),
            EventIds.EventQuit => new QuitPayload(DecodePlayerIdentity(d)),
            EventIds.EventDiagnostics => new DiagnosticsPayload(DecodeDiagnostics(d)),
            EventIds.EventJoin => new JoinPayload(DecodePlayerIdentity(d)),
            EventIds.EventPluginCommand => DecodePluginCommand(d),
            _ => throw new NotSupportedException($"Unsupported event id {eventId}."),
        };

        if (!d.Ok)
        {
            throw new InvalidDataException($"Failed to decode payload for event {eventId} ({EventIds.EventName(eventId)}).");
        }

        return decoded;
    }

    private static PluginCommandPayload DecodePluginCommand(AbiDecoder d)
    {
        var handlerId = d.U32();
        var count = ToCount(d.U32(), "plugin command arg count");
        var args = new List<string>(count);
        for (var i = 0; i < count; i++)
        {
            args.Add(d.String());
        }
        return new PluginCommandPayload(handlerId, args);
    }

    private static Vec3 DecodeVec3(AbiDecoder d) => new(d.F64(), d.F64(), d.F64());

    private static Rotation DecodeRotation(AbiDecoder d) => new(d.F64(), d.F64());

    private static Pos DecodePos(AbiDecoder d) => new(d.I32(), d.I32(), d.I32());

    private static WorldData DecodeWorld(AbiDecoder d)
    {
        if (!d.Bool())
        {
            return new WorldData(false, string.Empty, string.Empty);
        }
        return new WorldData(true, d.String(), d.String());
    }

    private static IReadOnlyDictionary<string, string> DecodeStringMap(AbiDecoder d)
    {
        var count = ToCount(d.U32(), "string map count");
        if (count == 0)
        {
            return new Dictionary<string, string>();
        }

        var values = new Dictionary<string, string>(count);
        for (var i = 0; i < count; i++)
        {
            values[d.String()] = d.String();
        }
        return values;
    }

    private static BlockData DecodeBlock(AbiDecoder d)
    {
        if (!d.Bool())
        {
            return new BlockData(false, string.Empty, new Dictionary<string, string>());
        }
        return new BlockData(true, d.String(), DecodeStringMap(d));
    }

    private static EntityData DecodeEntity(AbiDecoder d)
    {
        if (!d.Bool())
        {
            return new EntityData(false, false, string.Empty, string.Empty, default, default);
        }

        var hasHandle = d.Bool();
        var uuid = string.Empty;
        var type = string.Empty;
        if (hasHandle)
        {
            uuid = d.String();
            type = d.String();
        }

        return new EntityData(true, hasHandle, uuid, type, DecodeVec3(d), DecodeRotation(d));
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

    private static IReadOnlyList<ItemStackData> DecodeItemStacks(AbiDecoder d)
    {
        var count = ToCount(d.U32(), "item stack count");
        if (count == 0)
        {
            return Array.Empty<ItemStackData>();
        }

        var items = new List<ItemStackData>(count);
        for (var i = 0; i < count; i++)
        {
            items.Add(DecodeItemStack(d));
        }
        return items;
    }

    private static DamageSourceData DecodeDamageSource(AbiDecoder d)
    {
        if (!d.Bool())
        {
            return new DamageSourceData(false, string.Empty, false, false, false, false);
        }

        return new DamageSourceData(
            true,
            d.String(),
            d.Bool(),
            d.Bool(),
            d.Bool(),
            d.Bool());
    }

    private static HealingSourceData DecodeHealingSource(AbiDecoder d)
    {
        if (!d.Bool())
        {
            return new HealingSourceData(false, string.Empty);
        }
        return new HealingSourceData(true, d.String());
    }

    private static SkinData DecodeSkin(AbiDecoder d)
    {
        if (!d.Bool())
        {
            return new SkinData(false, 0, 0, false, string.Empty, string.Empty);
        }

        return new SkinData(
            true,
            d.I32(),
            d.I32(),
            d.Bool(),
            d.String(),
            d.String());
    }

    private static CommandData DecodeCommand(AbiDecoder d)
    {
        var name = d.String();
        var description = d.String();
        var usage = d.String();

        var count = ToCount(d.U32(), "command arg count");
        var args = new List<string>(count);
        for (var i = 0; i < count; i++)
        {
            args.Add(d.String());
        }

        return new CommandData(name, description, usage, args);
    }

    private static PlayerIdentity DecodePlayerIdentity(AbiDecoder d)
    {
        return new PlayerIdentity(d.String(), d.String(), d.String());
    }

    private static DiagnosticsData DecodeDiagnostics(AbiDecoder d)
    {
        return new DiagnosticsData(
            d.F64(),
            d.F64(),
            d.F64(),
            d.F64(),
            d.F64(),
            d.F64(),
            d.F64(),
            d.F64(),
            d.F64());
    }

    private static int ToCount(uint raw, string label)
    {
        if (raw > int.MaxValue)
        {
            throw new InvalidDataException($"{label} exceeds Int32.MaxValue.");
        }
        return (int)raw;
    }
}
