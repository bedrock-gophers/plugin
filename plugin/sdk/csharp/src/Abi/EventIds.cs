namespace BedrockPlugin.Sdk.Abi;

public static class EventIds
{
    public const ushort EventMove = 1;
    public const ushort EventJump = 2;
    public const ushort EventTeleport = 3;
    public const ushort EventChangeWorld = 4;
    public const ushort EventToggleSprint = 5;
    public const ushort EventToggleSneak = 6;
    public const ushort EventChat = 7;
    public const ushort EventFoodLoss = 8;
    public const ushort EventHeal = 9;
    public const ushort EventHurt = 10;
    public const ushort EventDeath = 11;
    public const ushort EventRespawn = 12;
    public const ushort EventSkinChange = 13;
    public const ushort EventFireExtinguish = 14;
    public const ushort EventStartBreak = 15;
    public const ushort EventBlockBreak = 16;
    public const ushort EventBlockPlace = 17;
    public const ushort EventBlockPick = 18;
    public const ushort EventItemUse = 19;
    public const ushort EventItemUseOnBlock = 20;
    public const ushort EventItemUseOnEntity = 21;
    public const ushort EventItemRelease = 22;
    public const ushort EventItemConsume = 23;
    public const ushort EventAttackEntity = 24;
    public const ushort EventExperienceGain = 25;
    public const ushort EventPunchAir = 26;
    public const ushort EventSignEdit = 27;
    public const ushort EventSleep = 28;
    public const ushort EventLecternPageTurn = 29;
    public const ushort EventItemDamage = 30;
    public const ushort EventItemPickup = 31;
    public const ushort EventHeldSlotChange = 32;
    public const ushort EventItemDrop = 33;
    public const ushort EventTransfer = 34;
    public const ushort EventCommandExecution = 35;
    public const ushort EventQuit = 36;
    public const ushort EventDiagnostics = 37;
    public const ushort EventJoin = 38;
    public const ushort EventPluginCommand = 39;

    public static string EventName(ushort id) => id switch
    {
        EventMove => "HandleMove",
        EventJump => "HandleJump",
        EventTeleport => "HandleTeleport",
        EventChangeWorld => "HandleChangeWorld",
        EventToggleSprint => "HandleToggleSprint",
        EventToggleSneak => "HandleToggleSneak",
        EventChat => "HandleChat",
        EventFoodLoss => "HandleFoodLoss",
        EventHeal => "HandleHeal",
        EventHurt => "HandleHurt",
        EventDeath => "HandleDeath",
        EventRespawn => "HandleRespawn",
        EventSkinChange => "HandleSkinChange",
        EventFireExtinguish => "HandleFireExtinguish",
        EventStartBreak => "HandleStartBreak",
        EventBlockBreak => "HandleBlockBreak",
        EventBlockPlace => "HandleBlockPlace",
        EventBlockPick => "HandleBlockPick",
        EventItemUse => "HandleItemUse",
        EventItemUseOnBlock => "HandleItemUseOnBlock",
        EventItemUseOnEntity => "HandleItemUseOnEntity",
        EventItemRelease => "HandleItemRelease",
        EventItemConsume => "HandleItemConsume",
        EventAttackEntity => "HandleAttackEntity",
        EventExperienceGain => "HandleExperienceGain",
        EventPunchAir => "HandlePunchAir",
        EventSignEdit => "HandleSignEdit",
        EventSleep => "HandleSleep",
        EventLecternPageTurn => "HandleLecternPageTurn",
        EventItemDamage => "HandleItemDamage",
        EventItemPickup => "HandleItemPickup",
        EventHeldSlotChange => "HandleHeldSlotChange",
        EventItemDrop => "HandleItemDrop",
        EventTransfer => "HandleTransfer",
        EventCommandExecution => "HandleCommandExecution",
        EventQuit => "HandleQuit",
        EventDiagnostics => "HandleDiagnostics",
        EventJoin => "HandleJoin",
        EventPluginCommand => "HandlePluginCommand",
        _ => "unknown",
    };
}
