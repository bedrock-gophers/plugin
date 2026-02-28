using BedrockPlugin.Sdk.Guest;
using System.Runtime.InteropServices;

namespace BedrockPlugin.Examples.PingPlugin;

public static unsafe class Exports
{
    [UnmanagedCallersOnly(EntryPoint = "PluginLoad")]
    public static int PluginLoad(NativeHostApi* hostApi, byte* pluginName)
    {
        return NativePlugin.Load(hostApi, pluginName);
    }

    [UnmanagedCallersOnly(EntryPoint = "PluginUnload")]
    public static void PluginUnload()
    {
        NativePlugin.Unload();
    }

    [UnmanagedCallersOnly(EntryPoint = "PluginDispatchEvent")]
    public static void PluginDispatchEvent(ushort version, ushort eventId, uint flags, ulong playerId, ulong requestKey, byte* payload, uint payloadLen)
    {
        NativePlugin.DispatchEvent(version, eventId, flags, playerId, requestKey, payload, payloadLen);
    }
}
