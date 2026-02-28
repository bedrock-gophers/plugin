namespace BedrockPlugin.Sdk.Guest;

public enum GameMode
{
    Survival = 0,
    Creative = 1,
    Adventure = 2,
    Spectator = 3,
}

public sealed class PlayerRef
{
    private readonly Plugin _plugin;

    internal PlayerRef(Plugin plugin, ulong id)
    {
        _plugin = plugin ?? throw new ArgumentNullException(nameof(plugin));
        Id = id;
    }

    public ulong Id { get; }

    private T HostValue<T>(T fallback, Func<IGuestHost, T> fn)
    {
        if (Id == 0)
        {
            return fallback;
        }
        return fn(_plugin.Host);
    }

    private bool HostBool(Func<IGuestHost, bool> fn)
    {
        return HostValue(false, fn);
    }

    private void HostDo(Action<IGuestHost> fn)
    {
        if (Id == 0)
        {
            return;
        }
        fn(_plugin.Host);
    }

    public double Health() => HostValue(0d, h => h.PlayerHealth(Id));

    public bool SetHealth(double health) => HostBool(h => h.SetPlayerHealth(Id, health));

    public int Food() => HostValue(0, h => h.PlayerFood(Id));

    public bool SetFood(int food) => HostBool(h => h.SetPlayerFood(Id, food));

    public string Name() => HostValue(string.Empty, h => h.PlayerName(Id));

    public GameMode GameMode() => HostValue(default(GameMode), h => (GameMode)h.PlayerGameMode(Id));

    public bool SetGameMode(GameMode mode) => HostBool(h => h.SetPlayerGameMode(Id, (int)mode));

    public string Xuid() => HostValue(string.Empty, h => h.PlayerXUID(Id));

    public string DeviceId() => HostValue(string.Empty, h => h.PlayerDeviceID(Id));

    public string DeviceModel() => HostValue(string.Empty, h => h.PlayerDeviceModel(Id));

    public string SelfSignedId() => HostValue(string.Empty, h => h.PlayerSelfSignedID(Id));

    public string NameTag() => HostValue(string.Empty, h => h.PlayerNameTag(Id));

    public bool SetNameTag(string value) => HostBool(h => h.SetPlayerNameTag(Id, value));

    public string ScoreTag() => HostValue(string.Empty, h => h.PlayerScoreTag(Id));

    public bool SetScoreTag(string value) => HostBool(h => h.SetPlayerScoreTag(Id, value));

    public double Absorption() => HostValue(0d, h => h.PlayerAbsorption(Id));

    public bool SetAbsorption(double value) => HostBool(h => h.SetPlayerAbsorption(Id, value));

    public double MaxHealth() => HostValue(0d, h => h.PlayerMaxHealth(Id));

    public bool SetMaxHealth(double value) => HostBool(h => h.SetPlayerMaxHealth(Id, value));

    public double Speed() => HostValue(0d, h => h.PlayerSpeed(Id));

    public bool SetSpeed(double value) => HostBool(h => h.SetPlayerSpeed(Id, value));

    public double FlightSpeed() => HostValue(0d, h => h.PlayerFlightSpeed(Id));

    public bool SetFlightSpeed(double value) => HostBool(h => h.SetPlayerFlightSpeed(Id, value));

    public double VerticalFlightSpeed() => HostValue(0d, h => h.PlayerVerticalFlightSpeed(Id));

    public bool SetVerticalFlightSpeed(double value) => HostBool(h => h.SetPlayerVerticalFlightSpeed(Id, value));

    public int Experience() => HostValue(0, h => h.PlayerExperience(Id));

    public int ExperienceLevel() => HostValue(0, h => h.PlayerExperienceLevel(Id));

    public bool SetExperienceLevel(int value) => HostBool(h => h.SetPlayerExperienceLevel(Id, value));

    public double ExperienceProgress() => HostValue(0d, h => h.PlayerExperienceProgress(Id));

    public bool SetExperienceProgress(double value) => HostBool(h => h.SetPlayerExperienceProgress(Id, value));

    public bool OnGround() => HostBool(h => h.PlayerOnGround(Id));

    public bool Sneaking() => HostBool(h => h.PlayerSneaking(Id));

    public bool SetSneaking(bool value) => HostBool(h => h.SetPlayerSneaking(Id, value));

    public bool Sprinting() => HostBool(h => h.PlayerSprinting(Id));

    public bool SetSprinting(bool value) => HostBool(h => h.SetPlayerSprinting(Id, value));

    public bool Swimming() => HostBool(h => h.PlayerSwimming(Id));

    public bool SetSwimming(bool value) => HostBool(h => h.SetPlayerSwimming(Id, value));

    public bool Flying() => HostBool(h => h.PlayerFlying(Id));

    public bool SetFlying(bool value) => HostBool(h => h.SetPlayerFlying(Id, value));

    public bool Gliding() => HostBool(h => h.PlayerGliding(Id));

    public bool SetGliding(bool value) => HostBool(h => h.SetPlayerGliding(Id, value));

    public bool Crawling() => HostBool(h => h.PlayerCrawling(Id));

    public bool SetCrawling(bool value) => HostBool(h => h.SetPlayerCrawling(Id, value));

    public bool UsingItem() => HostBool(h => h.PlayerUsingItem(Id));

    public bool Invisible() => HostBool(h => h.PlayerInvisible(Id));

    public bool SetInvisible(bool value) => HostBool(h => h.SetPlayerInvisible(Id, value));

    public bool Immobile() => HostBool(h => h.PlayerImmobile(Id));

    public bool SetImmobile(bool value) => HostBool(h => h.SetPlayerImmobile(Id, value));

    public bool Dead() => HostBool(h => h.PlayerDead(Id));

    public bool SetOnFireMillis(long millis) => HostBool(h => h.SetPlayerOnFireMillis(Id, millis));

    public bool AddFood(int points) => HostBool(h => h.AddPlayerFood(Id, points));

    public bool UseItem() => HostBool(h => h.PlayerUseItem(Id));

    public bool Jump() => HostBool(h => h.PlayerJump(Id));

    public bool SwingArm() => HostBool(h => h.PlayerSwingArm(Id));

    public bool Wake() => HostBool(h => h.PlayerWake(Id));

    public bool Extinguish() => HostBool(h => h.PlayerExtinguish(Id));

    public bool SetShowCoordinates(bool value) => HostBool(h => h.SetPlayerShowCoordinates(Id, value));

    public ItemStackData MainHandItem() => HostValue(new ItemStackData(0, false, string.Empty, 0, string.Empty), h => h.PlayerMainHandItem(Id));

    public bool SetMainHandItem(ItemStackData stack) => HostBool(h => h.SetPlayerMainHandItem(Id, stack));

    public ItemStackData OffHandItem() => HostValue(new ItemStackData(0, false, string.Empty, 0, string.Empty), h => h.PlayerOffHandItem(Id));

    public bool SetOffHandItem(ItemStackData stack) => HostBool(h => h.SetPlayerOffHandItem(Id, stack));

    public IReadOnlyList<ItemStackData> InventoryItems() => HostValue<IReadOnlyList<ItemStackData>>(Array.Empty<ItemStackData>(), h => h.PlayerInventoryItems(Id));

    public bool SetInventoryItems(IReadOnlyList<ItemStackData> items) => HostBool(h => h.SetPlayerInventoryItems(Id, items));

    public IReadOnlyList<ItemStackData> EnderChestItems() => HostValue<IReadOnlyList<ItemStackData>>(Array.Empty<ItemStackData>(), h => h.PlayerEnderChestItems(Id));

    public bool SetEnderChestItems(IReadOnlyList<ItemStackData> items) => HostBool(h => h.SetPlayerEnderChestItems(Id, items));

    public IReadOnlyList<ItemStackData> ArmourItems() => HostValue<IReadOnlyList<ItemStackData>>(Array.Empty<ItemStackData>(), h => h.PlayerArmourItems(Id));

    public bool SetArmourItems(IReadOnlyList<ItemStackData> items) => HostBool(h => h.SetPlayerArmourItems(Id, items));

    public bool SetHeldSlot(int slot) => HostBool(h => h.SetPlayerHeldSlot(Id, slot));

    public bool MoveItemsToInventory() => HostBool(h => h.PlayerMoveItemsToInventory(Id));

    public bool CloseForm() => HostBool(h => h.PlayerCloseForm(Id));

    public bool CloseDialogue() => HostBool(h => h.PlayerCloseDialogue(Id));

    public bool SendMenuForm(MenuFormData value) => HostBool(h => h.PlayerSendMenuForm(Id, value));

    public bool SendModalForm(ModalFormData value) => HostBool(h => h.PlayerSendModalForm(Id, value));

    public void Message(string message)
    {
        HostDo(h => h.PlayerMessage(Id, message));
    }

    public void Messagef(string format, params object?[] args)
    {
        Message(TextFormat.Colourf(format, args));
    }
}
