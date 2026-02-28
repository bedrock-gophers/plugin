namespace BedrockPlugin.Sdk.Guest;

public static partial class Item
{
    public static ItemStackData NewStack(
        string itemName,
        int count = 1,
        int itemMeta = 0,
        string customName = "")
    {
        return new ItemStackData(count, true, itemName ?? string.Empty, itemMeta, customName ?? string.Empty);
    }
}
