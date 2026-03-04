using BedrockPlugin.Sdk.Guest;
using System.Runtime.CompilerServices;

namespace BedrockPlugin.Examples.PingPlugin;

public sealed class PingPlugin : Plugin
{
    private readonly Dictionary<ulong, double> _lastYawByPlayerId = new();
    private static readonly IReadOnlyList<CommandParameterSpec> PingCommandParameters = new List<CommandParameterSpec>
    {
        CommandParameterSpec.Target(optional: true),
    };

    private static readonly CommandOverloadSpec[] PingCommandOverloads =
    {
        new(PingCommandParameters),
    };

    public PingPlugin(string name, IGuestHost host) : base(name, host)
    {
        RegisterCommand("ping", "Show player latency in milliseconds.", Array.Empty<string>(), PingCommandOverloads, HandlePingCommand);
    }

    private void HandlePingCommand(CommandContext ctx, IReadOnlyList<string> args)
    {
        if (args.Count > 1)
        {
            ctx.Message("usage: /ping [target]");
            return;
        }

        if (!TryResolvePingTarget(ctx, args, out var target))
        {
            return;
        }

        var latencyNanos = target.Latency();
        var latencyMillis = Math.Max(0L, (long)Math.Round(latencyNanos / 1_000_000d, MidpointRounding.AwayFromZero));
        var targetName = target.Name();
        if (string.IsNullOrWhiteSpace(targetName))
        {
            targetName = "target";
        }

        if (ctx.IsPlayer && args.Count == 0)
        {
            ctx.Messagef("your latency is {0}ms", latencyMillis);
            return;
        }
        ctx.Messagef("{0} latency is {1}ms", targetName, latencyMillis);
    }

    private bool TryResolvePingTarget(CommandContext ctx, IReadOnlyList<string> args, out Player target)
    {
        if (args.Count == 0)
        {
            var sourcePlayer = ctx.Player;
            if (sourcePlayer is null)
            {
                ctx.Message("console must provide a target player");
                target = null!;
                return false;
            }
            target = sourcePlayer;
            return true;
        }

        var targetName = (args[0] ?? string.Empty).Trim();
        if (targetName.Length == 0)
        {
            ctx.Message("target must be a non-empty player name");
            target = null!;
            return false;
        }

        var resolved = PlayerByName(targetName);
        if (resolved is null)
        {
            ctx.Messagef("cannot resolve online player {0}", targetName);
            target = null!;
            return false;
        }

        target = resolved;
        return true;
    }


    private void OnMove(EventContext ctx, Vec3 _, Rotation newRot)
    {
        var player = ctx.Player;
        if (player is null)
        {
            return;
        }

        var yaw = newRot.Yaw;
        if (_lastYawByPlayerId.TryGetValue(ctx.PlayerId, out var previousYaw) && !NearlyEqual(previousYaw, yaw))
        {
            ctx.Message("your yaw changed");
        }
        _lastYawByPlayerId[ctx.PlayerId] = yaw;
    }

    private static bool NearlyEqual(double a, double b)
    {
        return Math.Abs(a - b) <= 0.00001d;
    }

    private void OnItemDrop(EventContext _, MutableArgument<ItemStackData> i)
    {
        var dropped = i.Get();
        if (!dropped.HasItem || !string.Equals(dropped.ItemName, "minecraft:cooked_beef", StringComparison.OrdinalIgnoreCase))
        {
            return;
        }

        i.Set(Item.NewStack("minecraft:beef", dropped.Count, 0, dropped.CustomName));
    }

    [ModuleInitializer]
    internal static void RegisterNativePlugin()
    {
        NativePlugin.Register<PingPlugin>();
    }
}
