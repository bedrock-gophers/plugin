namespace BedrockPlugin.Sdk.Guest;

public static partial class Item
{
    public sealed class Stack
    {
        private static readonly ItemStackData EmptyStack = new(0, false, string.Empty, 0, string.Empty);
        private readonly ItemStackData _data;

        internal Stack(ItemStackData data)
        {
            _data = Normalize(data);
        }

        public int Count()
        {
            return _data.Count;
        }

        public string CustomName()
        {
            return _data.CustomName ?? string.Empty;
        }

        public bool Empty()
        {
            return !_data.HasItem;
        }

        public ItemStackData ToData()
        {
            return _data;
        }

        public static Stack FromData(ItemStackData value)
        {
            return new Stack(value);
        }

        private static ItemStackData Normalize(ItemStackData value)
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
    }
}
