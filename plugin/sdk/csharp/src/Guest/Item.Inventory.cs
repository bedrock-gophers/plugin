namespace BedrockPlugin.Sdk.Guest;

public static partial class Item
{
    internal enum InventoryTarget
    {
        Inventory,
        EnderChest,
        Armour,
    }

    public sealed class Inventory
    {
        private readonly Plugin _plugin;
        private readonly ulong _playerId;
        private readonly InventoryTarget _target;

        internal Inventory(Plugin plugin, ulong playerId, InventoryTarget target)
        {
            _plugin = plugin ?? throw new ArgumentNullException(nameof(plugin));
            _playerId = playerId;
            _target = target;
        }

        private IReadOnlyList<ItemStackData> RawItems()
        {
            return _target switch
            {
                InventoryTarget.Inventory => _plugin.Host.PlayerInventoryItems(_playerId),
                InventoryTarget.EnderChest => _plugin.Host.PlayerEnderChestItems(_playerId),
                InventoryTarget.Armour => _plugin.Host.PlayerArmourItems(_playerId),
                _ => Array.Empty<ItemStackData>(),
            };
        }

        private bool SetRawItems(IReadOnlyList<ItemStackData> items)
        {
            return _target switch
            {
                InventoryTarget.Inventory => _plugin.Host.SetPlayerInventoryItems(_playerId, items),
                InventoryTarget.EnderChest => _plugin.Host.SetPlayerEnderChestItems(_playerId, items),
                InventoryTarget.Armour => _plugin.Host.SetPlayerArmourItems(_playerId, items),
                _ => false,
            };
        }

        public IReadOnlyList<Stack> Items()
        {
            return RawItems().Select(static item => Stack.FromData(item)).ToArray();
        }

        public bool SetItems(IReadOnlyList<Stack> items)
        {
            if (items is null)
            {
                return SetRawItems(Array.Empty<ItemStackData>());
            }
            return SetRawItems(items.Select(static item => item.ToData()).ToArray());
        }

        public bool AddItem(Stack item)
        {
            if (item is null)
            {
                throw new ArgumentNullException(nameof(item));
            }

            var items = RawItems().ToArray();
            for (var slot = 0; slot < items.Length; slot++)
            {
                if (items[slot].HasItem)
                {
                    continue;
                }

                items[slot] = item.ToData();
                return SetRawItems(items);
            }
            return false;
        }

        public bool AddItem(ItemStackData item)
        {
            return AddItem(Stack.FromData(item));
        }

        public int Size()
        {
            return RawItems().Count;
        }

        public Stack ItemAt(int slot)
        {
            var items = RawItems();
            if (slot < 0 || slot >= items.Count)
            {
                throw new ArgumentOutOfRangeException(nameof(slot));
            }
            return Stack.FromData(items[slot]);
        }

        public bool SetItemAt(int slot, Stack item)
        {
            if (item is null)
            {
                throw new ArgumentNullException(nameof(item));
            }

            var items = RawItems().ToArray();
            if (slot < 0 || slot >= items.Length)
            {
                throw new ArgumentOutOfRangeException(nameof(slot));
            }
            items[slot] = item.ToData();
            return SetRawItems(items);
        }

        public bool Clear()
        {
            var items = RawItems().ToArray();
            for (var i = 0; i < items.Length; i++)
            {
                items[i] = new ItemStackData(0, false, string.Empty, 0, string.Empty);
            }
            return SetRawItems(items);
        }
    }
}
