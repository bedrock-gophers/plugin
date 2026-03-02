using BedrockPlugin.Sdk.Guest;
using System.Diagnostics.CodeAnalysis;
using System.Runtime.CompilerServices;

namespace BedrockPlugin.Examples.PingPlugin;

public sealed class PingPlugin : Plugin
{
    private readonly Dictionary<ulong, double> _lastYawByPlayerId = new();
    private static readonly CommandOverloadSpec[] PingCommandOverloads =
    {
        new(new[]
        {
            CommandParameterSpec.Target(optional: true),
        }),
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

        var latency = target.Latency();
        var latencyMillis = Math.Max(0L, (long)Math.Round(latency.TotalMilliseconds, MidpointRounding.AwayFromZero));
        var targetName = target.Name();
        if (string.IsNullOrWhiteSpace(targetName))
        {
            targetName = "target";
        }

        if (ctx.TryPlayer(out var sourcePlayer) && sourcePlayer is not null && sourcePlayer.Id == target.Id)
        {
            ctx.Messagef("your latency is {0}ms", latencyMillis);
            return;
        }
        ctx.Messagef("{0} latency is {1}ms", targetName, latencyMillis);
    }

    private bool TryResolvePingTarget(CommandContext ctx, IReadOnlyList<string> args, out PlayerRef target)
    {
        if (args.Count == 0)
        {
            if (!ctx.TryPlayer(out var sourcePlayer) || sourcePlayer is null)
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

        if (!TryPlayerByName(targetName, out var resolved) || resolved is null)
        {
            ctx.Messagef("cannot resolve online player {0}", targetName);
            target = null!;
            return false;
        }

        target = resolved;
        return true;
    }


    [SuppressMessage("Style", "IDE0051:Remove unused private members", Justification = "Invoked via Plugin convention reflection.")]
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

    [SuppressMessage("Style", "IDE0051:Remove unused private members", Justification = "Invoked via Plugin convention reflection.")]
    private void OnItemDrop(EventContext _, MutableArgument<ItemStackData> i)
    {
        var dropped = i.Get();
        if (!dropped.HasItem || !string.Equals(dropped.ItemName, Item.CookedBeef, StringComparison.OrdinalIgnoreCase))
        {
            return;
        }

        i.Set(Item.NewStack(Item.Beef, dropped.Count, 0, dropped.CustomName));
    }

    [ModuleInitializer]
    internal static void RegisterNativePlugin()
    {
        NativePlugin.Register<PingPlugin>();
    }
}
