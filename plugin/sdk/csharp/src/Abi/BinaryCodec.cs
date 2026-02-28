using System.Buffers.Binary;
using System.Text;

namespace BedrockPlugin.Sdk.Abi;

public sealed class AbiEncoder
{
    private readonly List<byte> _buffer;

    public AbiEncoder(int capacityHint = 0)
    {
        _buffer = capacityHint > 0 ? new List<byte>(capacityHint) : new List<byte>();
    }

    public void U8(byte value) => _buffer.Add(value);

    public void Bool(bool value) => U8(value ? (byte)1 : (byte)0);

    public void I32(int value) => U32(unchecked((uint)value));

    public void U32(uint value)
    {
        _buffer.Add((byte)value);
        _buffer.Add((byte)(value >> 8));
        _buffer.Add((byte)(value >> 16));
        _buffer.Add((byte)(value >> 24));
    }

    public void I64(long value) => U64(unchecked((ulong)value));

    public void U64(ulong value)
    {
        _buffer.Add((byte)value);
        _buffer.Add((byte)(value >> 8));
        _buffer.Add((byte)(value >> 16));
        _buffer.Add((byte)(value >> 24));
        _buffer.Add((byte)(value >> 32));
        _buffer.Add((byte)(value >> 40));
        _buffer.Add((byte)(value >> 48));
        _buffer.Add((byte)(value >> 56));
    }

    public void F64(double value) => U64((ulong)BitConverter.DoubleToInt64Bits(value));

    public void Bytes(ReadOnlySpan<byte> value)
    {
        U32((uint)value.Length);
        for (var i = 0; i < value.Length; i++)
        {
            _buffer.Add(value[i]);
        }
    }

    public void String(string value)
    {
        value ??= string.Empty;
        Bytes(Encoding.UTF8.GetBytes(value));
    }

    public byte[] Data() => _buffer.ToArray();
}

public sealed class AbiDecoder
{
    private readonly ReadOnlyMemory<byte> _data;
    private int _offset;
    private bool _ok = true;

    public AbiDecoder(ReadOnlySpan<byte> data)
    {
        _data = data.ToArray();
    }

    public AbiDecoder(byte[] data)
    {
        _data = data ?? Array.Empty<byte>();
    }

    public bool Ok => _ok;

    public byte U8()
    {
        if (!_ok || _offset + 1 > _data.Length)
        {
            _ok = false;
            return 0;
        }
        var value = _data.Span[_offset];
        _offset++;
        return value;
    }

    public bool Bool() => U8() == 1;

    public int I32() => unchecked((int)U32());

    public uint U32()
    {
        if (!TryRead(4, out var slice))
        {
            return 0;
        }
        return BinaryPrimitives.ReadUInt32LittleEndian(slice);
    }

    public long I64() => unchecked((long)U64());

    public ulong U64()
    {
        if (!TryRead(8, out var slice))
        {
            return 0;
        }
        return BinaryPrimitives.ReadUInt64LittleEndian(slice);
    }

    public double F64() => BitConverter.Int64BitsToDouble(unchecked((long)U64()));

    public byte[] Bytes()
    {
        var sizeU = U32();
        if (!_ok)
        {
            return Array.Empty<byte>();
        }
        if (sizeU > int.MaxValue)
        {
            _ok = false;
            return Array.Empty<byte>();
        }

        var size = (int)sizeU;
        if (!TryRead(size, out var slice))
        {
            return Array.Empty<byte>();
        }
        return slice.ToArray();
    }

    public string String() => Encoding.UTF8.GetString(Bytes());

    private bool TryRead(int size, out ReadOnlySpan<byte> slice)
    {
        if (!_ok || size < 0 || _offset + size > _data.Length)
        {
            _ok = false;
            slice = ReadOnlySpan<byte>.Empty;
            return false;
        }

        slice = _data.Span.Slice(_offset, size);
        _offset += size;
        return true;
    }
}
