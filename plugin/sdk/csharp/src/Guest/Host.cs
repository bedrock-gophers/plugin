namespace BedrockPlugin.Sdk.Guest;

public interface IGuestHost
{
    bool RegisterCommand(
        string pluginName,
        string name,
        string description,
        IReadOnlyList<string> aliases,
        uint handlerId,
        IReadOnlyList<CommandOverloadSpec> overloads);

    IReadOnlyList<string> ManagePlugins(uint action, string target);

    ulong ResolvePlayerByName(string name);

    IReadOnlyList<string> OnlinePlayerNames();
    IReadOnlyList<string> BlockNames();
    IReadOnlyList<string> ItemNames();
    IReadOnlyList<string> WorldNames();

    void ConsoleMessage(string pluginName, string message);
    bool EventCancel(ulong requestKey);
    bool EventSetItemDrop(ulong requestKey, ItemStackData stack);

    double PlayerHealth(ulong playerId);
    bool SetPlayerHealth(ulong playerId, double health);
    int PlayerFood(ulong playerId);
    bool SetPlayerFood(ulong playerId, int food);
    string PlayerName(ulong playerId);
    int PlayerGameMode(ulong playerId);
    bool SetPlayerGameMode(ulong playerId, int mode);
    string PlayerXUID(ulong playerId);
    string PlayerDeviceID(ulong playerId);
    string PlayerDeviceModel(ulong playerId);
    string PlayerSelfSignedID(ulong playerId);
    string PlayerNameTag(ulong playerId);
    bool SetPlayerNameTag(ulong playerId, string value);
    string PlayerScoreTag(ulong playerId);
    bool SetPlayerScoreTag(ulong playerId, string value);
    double PlayerAbsorption(ulong playerId);
    bool SetPlayerAbsorption(ulong playerId, double value);
    double PlayerMaxHealth(ulong playerId);
    bool SetPlayerMaxHealth(ulong playerId, double value);
    double PlayerSpeed(ulong playerId);
    bool SetPlayerSpeed(ulong playerId, double value);
    double PlayerFlightSpeed(ulong playerId);
    bool SetPlayerFlightSpeed(ulong playerId, double value);
    double PlayerVerticalFlightSpeed(ulong playerId);
    bool SetPlayerVerticalFlightSpeed(ulong playerId, double value);
    int PlayerExperience(ulong playerId);
    int PlayerExperienceLevel(ulong playerId);
    bool SetPlayerExperienceLevel(ulong playerId, int value);
    double PlayerExperienceProgress(ulong playerId);
    bool SetPlayerExperienceProgress(ulong playerId, double value);
    bool PlayerOnGround(ulong playerId);
    bool PlayerSneaking(ulong playerId);
    bool SetPlayerSneaking(ulong playerId, bool value);
    bool PlayerSprinting(ulong playerId);
    bool SetPlayerSprinting(ulong playerId, bool value);
    bool PlayerSwimming(ulong playerId);
    bool SetPlayerSwimming(ulong playerId, bool value);
    bool PlayerFlying(ulong playerId);
    bool SetPlayerFlying(ulong playerId, bool value);
    bool PlayerGliding(ulong playerId);
    bool SetPlayerGliding(ulong playerId, bool value);
    bool PlayerCrawling(ulong playerId);
    bool SetPlayerCrawling(ulong playerId, bool value);
    bool PlayerUsingItem(ulong playerId);
    bool PlayerInvisible(ulong playerId);
    bool SetPlayerInvisible(ulong playerId, bool value);
    bool PlayerImmobile(ulong playerId);
    bool SetPlayerImmobile(ulong playerId, bool value);
    bool PlayerDead(ulong playerId);
    long PlayerLatencyMillis(ulong playerId);
    bool SetPlayerOnFireMillis(ulong playerId, long millis);
    bool AddPlayerFood(ulong playerId, int points);
    bool PlayerUseItem(ulong playerId);
    bool PlayerJump(ulong playerId);
    bool PlayerSwingArm(ulong playerId);
    bool PlayerWake(ulong playerId);
    bool PlayerExtinguish(ulong playerId);
    bool SetPlayerShowCoordinates(ulong playerId, bool value);
    ItemStackData PlayerMainHandItem(ulong playerId);
    bool SetPlayerMainHandItem(ulong playerId, ItemStackData stack);
    ItemStackData PlayerOffHandItem(ulong playerId);
    bool SetPlayerOffHandItem(ulong playerId, ItemStackData stack);
    IReadOnlyList<ItemStackData> PlayerInventoryItems(ulong playerId);
    bool SetPlayerInventoryItems(ulong playerId, IReadOnlyList<ItemStackData> items);
    IReadOnlyList<ItemStackData> PlayerEnderChestItems(ulong playerId);
    bool SetPlayerEnderChestItems(ulong playerId, IReadOnlyList<ItemStackData> items);
    IReadOnlyList<ItemStackData> PlayerArmourItems(ulong playerId);
    bool SetPlayerArmourItems(ulong playerId, IReadOnlyList<ItemStackData> items);
    bool SetPlayerHeldSlot(ulong playerId, int slot);
    bool PlayerMoveItemsToInventory(ulong playerId);
    bool PlayerCloseForm(ulong playerId);
    bool PlayerCloseDialogue(ulong playerId);
    bool PlayerSendMenuForm(ulong playerId, MenuFormData value);
    bool PlayerSendModalForm(ulong playerId, ModalFormData value);
    void PlayerMessage(ulong playerId, string message);
}
