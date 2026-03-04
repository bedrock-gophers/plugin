namespace BedrockPlugin.Sdk.Guest;

public static class ItemStackMatchExtensions
{
    public static bool Equals(this Item.Stack stack, ItemStackData item)
    {
        if (stack is null)
        {
            return false;
        }

        var data = ItemStackInterop.ToData(stack);
        if (!data.HasItem || !item.HasItem)
        {
            return data.HasItem == item.HasItem;
        }

        return string.Equals(data.ItemName, item.ItemName, StringComparison.OrdinalIgnoreCase) &&
               data.ItemMeta == item.ItemMeta;
    }
}
