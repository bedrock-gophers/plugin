namespace BedrockPlugin.Sdk.Guest;

public sealed class Player
{
    private readonly Plugin _plugin;
    private Item.Inventory? _inventory;
    private Item.Inventory? _enderChest;
    private Item.Inventory? _armour;

    internal Player(Plugin plugin, ulong id, ulong handle)
    {
        _plugin = plugin ?? throw new ArgumentNullException(nameof(plugin));
        Id = id;
        Handle = handle;
    }

    public ulong Id { get; }

    public ulong Handle { get; }

    public string Name()
    {
        return _plugin.Host.PlayerName(Id);
    }

    // Dragonfly exposes player latency as nanoseconds.
    public long Latency()
    {
        return LatencyDuration().Ticks * 100L;
    }

    public TimeSpan LatencyDuration()
    {
        return _plugin.Host.PlayerLatency(Id);
    }

    public void Message(string message)
    {
        _plugin.Host.PlayerMessage(Id, message ?? string.Empty);
    }

    public Item.Stack MainHandItem()
    {
        return Item.Stack.FromData(_plugin.Host.PlayerMainHandItem(Id));
    }

    public bool SetMainHandItem(Item.Stack stack)
    {
        if (stack is null)
        {
            throw new ArgumentNullException(nameof(stack));
        }
        return _plugin.Host.SetPlayerMainHandItem(Id, stack.ToData());
    }

    public Item.Stack OffHandItem()
    {
        return Item.Stack.FromData(_plugin.Host.PlayerOffHandItem(Id));
    }

    public bool SetOffHandItem(Item.Stack stack)
    {
        if (stack is null)
        {
            throw new ArgumentNullException(nameof(stack));
        }
        return _plugin.Host.SetPlayerOffHandItem(Id, stack.ToData());
    }

    public Item.Inventory Inventory()
    {
        return _inventory ??= new Item.Inventory(_plugin, Id, Item.InventoryTarget.Inventory);
    }

    public Item.Inventory EnderChest()
    {
        return _enderChest ??= new Item.Inventory(_plugin, Id, Item.InventoryTarget.EnderChest);
    }

    public Item.Inventory Armour()
    {
        return _armour ??= new Item.Inventory(_plugin, Id, Item.InventoryTarget.Armour);
    }

    public bool MoveItemsToInventory()
    {
        return _plugin.Host.PlayerMoveItemsToInventory(Id);
    }
}
