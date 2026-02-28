namespace BedrockPlugin.Sdk.Guest;

public readonly record struct Vec3(double X, double Y, double Z);

public readonly record struct Rotation(double Yaw, double Pitch);

public readonly record struct Pos(int X, int Y, int Z);

public sealed record WorldData(bool Present, string Name, string Dimension);

public sealed record BlockData(bool Present, string Name, IReadOnlyDictionary<string, string> Properties);

public sealed record EntityData(
    bool Present,
    bool HasHandle,
    string Uuid,
    string Type,
    Vec3 Position,
    Rotation Rotation);

public sealed record ItemStackData(
    int Count,
    bool HasItem,
    string ItemName,
    int ItemMeta,
    string CustomName);

public sealed record MenuFormData(
    string Title,
    string Body,
    IReadOnlyList<string> Buttons);

public sealed record ModalFormData(
    string Title,
    string Body,
    string Confirm,
    string Cancel);

public sealed record DamageSourceData(
    bool Present,
    string Type,
    bool ReducedByArmour,
    bool ReducedByResistance,
    bool Fire,
    bool IgnoreTotem);

public sealed record HealingSourceData(bool Present, string Type);

public sealed record SkinData(
    bool Present,
    int Width,
    int Height,
    bool Persona,
    string PlayFabId,
    string FullId);

public sealed record CommandData(
    string Name,
    string Description,
    string Usage,
    IReadOnlyList<string> Args);

public sealed record DiagnosticsData(
    double AverageFramesPerSecond,
    double AverageServerSimTickTime,
    double AverageClientSimTickTime,
    double AverageBeginFrameTime,
    double AverageInputTime,
    double AverageRenderTime,
    double AverageEndFrameTime,
    double AverageRemainderTimePercent,
    double AverageUnaccountedTimePercent);

public sealed record PlayerIdentity(string Name, string Uuid, string Xuid);

public enum CommandParameterKind : byte
{
    String = 0,
    Text = 1,
    Enum = 2,
    Subcommand = 3,
    PluginAvailable = 4,
    PluginLoaded = 5,
}

public sealed record CommandParameterSpec(
    string Name,
    CommandParameterKind Kind,
    bool Optional,
    IReadOnlyList<string> EnumOptions);

public sealed record CommandOverloadSpec(IReadOnlyList<CommandParameterSpec> Parameters);
