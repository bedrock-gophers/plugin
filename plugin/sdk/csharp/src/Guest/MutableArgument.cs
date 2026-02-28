namespace BedrockPlugin.Sdk.Guest;

public sealed class MutableArgument<T>
{
    private T _value;

    internal MutableArgument(T initial)
    {
        _value = initial;
    }

    internal bool IsChanged { get; private set; }

    internal T Value => _value;

    public T Get()
    {
        return _value;
    }

    public void Set(T value)
    {
        _value = value;
        IsChanged = true;
    }
}
