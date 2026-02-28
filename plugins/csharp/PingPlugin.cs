using BedrockPlugin.Sdk.Guest;
using System.Runtime.CompilerServices;

namespace BedrockPlugin.Examples.PingPlugin;

public sealed class PingPlugin : Plugin
{
    private readonly Dictionary<ulong, double> _lastYawByPlayerId = new();

    public PingPlugin(string name, IGuestHost host) : base(name, host)
    {
        RegisterCommand("ping", "Replies with pong.", Array.Empty<string>(), ctx => ctx.Message("pong"));
    }

    private void OnMove(EventContext ctx, Vec3 _, Rotation newRot)
    {
        if (!ctx.TryPlayer(out var player) || player is null)
        {
            return;
        }

        var yaw = newRot.Yaw;
        if (_lastYawByPlayerId.TryGetValue(player.Id, out var previousYaw) && !NearlyEqual(previousYaw, yaw))
        {
            player.Message("your yaw changed");
        }
        _lastYawByPlayerId[player.Id] = yaw;
    }

    private static bool NearlyEqual(double a, double b)
    {
        return Math.Abs(a - b) <= 0.00001d;
    }

    [ModuleInitializer]
    internal static void RegisterNativePlugin()
    {
        NativePlugin.Register<PingPlugin>();
    }
}
