namespace BedrockPlugin.Sdk.Guest;

public static class MutableArgumentItemStackExtensions
{
    public static ItemStackData GetData(this MutableArgument<Item.Stack> arg)
    {
        if (arg is null)
        {
            throw new ArgumentNullException(nameof(arg));
        }
        return ItemStackInterop.ToData(arg.Get());
    }

    public static void Set(this MutableArgument<Item.Stack> arg, ItemStackData value)
    {
        if (arg is null)
        {
            throw new ArgumentNullException(nameof(arg));
        }
        arg.Set(ItemStackInterop.FromDataStack(value));
    }
}
