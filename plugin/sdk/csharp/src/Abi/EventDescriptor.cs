using System.Buffers.Binary;

namespace BedrockPlugin.Sdk.Abi;

public static class AbiConstants
{
    public const ushort Version = 1;

    public const uint FlagCancellable = 1u << 0;
    public const uint FlagSynchronous = 1u << 1;

    public const int EventDescriptorSize = 24;
}

public readonly record struct EventDescriptor(
    ushort Version,
    ushort EventId,
    uint Flags,
    ulong PlayerId,
    ulong RequestKey)
{
    public static EventDescriptor Decode(ReadOnlySpan<byte> src)
    {
        if (src.Length < AbiConstants.EventDescriptorSize)
        {
            throw new ArgumentException($"Descriptor requires {AbiConstants.EventDescriptorSize} bytes.", nameof(src));
        }

        return new EventDescriptor(
            BinaryPrimitives.ReadUInt16LittleEndian(src[0..2]),
            BinaryPrimitives.ReadUInt16LittleEndian(src[2..4]),
            BinaryPrimitives.ReadUInt32LittleEndian(src[4..8]),
            BinaryPrimitives.ReadUInt64LittleEndian(src[8..16]),
            BinaryPrimitives.ReadUInt64LittleEndian(src[16..24]));
    }

    public static void Encode(Span<byte> dst, EventDescriptor descriptor)
    {
        if (dst.Length < AbiConstants.EventDescriptorSize)
        {
            throw new ArgumentException($"Descriptor requires {AbiConstants.EventDescriptorSize} bytes.", nameof(dst));
        }

        BinaryPrimitives.WriteUInt16LittleEndian(dst[0..2], descriptor.Version);
        BinaryPrimitives.WriteUInt16LittleEndian(dst[2..4], descriptor.EventId);
        BinaryPrimitives.WriteUInt32LittleEndian(dst[4..8], descriptor.Flags);
        BinaryPrimitives.WriteUInt64LittleEndian(dst[8..16], descriptor.PlayerId);
        BinaryPrimitives.WriteUInt64LittleEndian(dst[16..24], descriptor.RequestKey);
    }
}
