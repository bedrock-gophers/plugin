namespace BedrockPlugin.Sdk.Guest;

public static class ItemStackInterop
{
    private static readonly ItemStackData EmptyStack = new(0, false, string.Empty, 0, string.Empty);

    public static ItemStackData FromData(ItemStackData value)
    {
        if (!value.HasItem)
        {
            return EmptyStack;
        }
        var count = value.Count <= 0 ? 1 : value.Count;
        return new ItemStackData(
            count,
            true,
            value.ItemName ?? string.Empty,
            value.ItemMeta,
            value.CustomName ?? string.Empty);
    }

    public static Item.Stack FromDataStack(ItemStackData value)
    {
        return new Item.Stack(value);
    }

    public static ItemStackData ToData(ItemStackData value)
    {
        return FromData(value);
    }

    public static ItemStackData ToData(Item.Stack? value)
    {
        return value is null ? EmptyStack : ToData(value.ToData());
    }
}
