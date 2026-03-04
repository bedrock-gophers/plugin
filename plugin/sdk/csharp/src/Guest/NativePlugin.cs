using BedrockPlugin.Sdk.Abi;
using System.Diagnostics.CodeAnalysis;

namespace BedrockPlugin.Sdk.Guest;

public static unsafe class NativePlugin
{
    private static readonly object Gate = new();

    private static Func<string, IGuestHost, Plugin>? _factory;
    private static Plugin? _plugin;
    private static NativeGuestHost? _host;
    private static string _pluginName = "csharp";

    public static void Register(Func<string, IGuestHost, Plugin> factory)
    {
        if (factory is null)
        {
            throw new ArgumentNullException(nameof(factory));
        }

        lock (Gate)
        {
            _factory = factory;
        }
    }

    public static void Register<
        [DynamicallyAccessedMembers(
            DynamicallyAccessedMemberTypes.PublicConstructors |
            DynamicallyAccessedMemberTypes.NonPublicConstructors |
            DynamicallyAccessedMemberTypes.PublicMethods |
            DynamicallyAccessedMemberTypes.NonPublicMethods)] TPlugin>()
        where TPlugin : Plugin
    {
        Register(static (name, host) =>
        {
            var instance = Activator.CreateInstance(typeof(TPlugin), new object?[] { name, host });
            if (instance is Plugin plugin)
            {
                return plugin;
            }
            throw new InvalidOperationException($"failed to create plugin instance for {typeof(TPlugin).FullName}");
        });
    }

    public static int Load(NativeHostApi* hostApi, byte* pluginName)
    {
        if (hostApi is null)
        {
            return 0;
        }

        lock (Gate)
        {
            var host = new NativeGuestHost(hostApi);
            var name = NativeGuestHost.ReadUtf8(pluginName);
            if (name.Length == 0)
            {
                name = "csharp";
            }

            var factory = _factory;
            if (factory is null)
            {
                host.ConsoleMessage(name, "PluginLoad failed: NativePlugin.Register(...) was not called.");
                return 0;
            }

            try
            {
                _host = host;
                _pluginName = name;
                _plugin = factory(name, host);
                return 1;
            }
            catch (Exception ex)
            {
                var detail = ex.ToString();
                if (ex.InnerException is not null)
                {
                    detail = $"{detail}\nInner: {ex.InnerException}";
                }
                host.ConsoleMessage(name, $"PluginLoad failed: {detail}");
                _plugin = null;
                _host = null;
                _pluginName = "csharp";
                return 0;
            }
        }
    }

    public static void Unload()
    {
        lock (Gate)
        {
            _plugin = null;
            _host = null;
            _pluginName = "csharp";
        }
    }

    public static void DispatchEvent(
        ushort version,
        ushort eventId,
        uint flags,
        ulong playerId,
        ulong requestKey,
        byte* payload,
        uint payloadLen)
    {
        Plugin? plugin;
        NativeGuestHost? host;
        string pluginName;

        lock (Gate)
        {
            plugin = _plugin;
            host = _host;
            pluginName = _pluginName;
        }

        if (plugin is null)
        {
            return;
        }

        if (payloadLen > int.MaxValue)
        {
            host?.ConsoleMessage(pluginName, $"event payload too large: {payloadLen}");
            return;
        }

        var descriptor = new EventDescriptor(version, eventId, flags, playerId, requestKey);
        var data = payload is null || payloadLen == 0
            ? ReadOnlySpan<byte>.Empty
            : new ReadOnlySpan<byte>(payload, (int)payloadLen);

        try
        {
            plugin.DispatchEvent(descriptor, data);
        }
        catch (Exception ex)
        {
            host?.ConsoleMessage(pluginName, $"PluginDispatchEvent failed for {EventIds.EventName(eventId)}: {ex.Message}");
        }
    }
}
