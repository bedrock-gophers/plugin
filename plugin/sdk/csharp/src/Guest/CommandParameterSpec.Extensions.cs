namespace BedrockPlugin.Sdk.Guest;

public sealed partial record CommandParameterSpec
{
    public static CommandParameterSpec Target(string name = "target", bool optional = false)
    {
        return new(name, CommandParameterKind.Target, optional, System.Array.Empty<string>());
    }
}
