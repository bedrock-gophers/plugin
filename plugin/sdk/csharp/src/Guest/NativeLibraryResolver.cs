using System.Reflection;
using System.Runtime.CompilerServices;
using System.Runtime.InteropServices;

namespace BedrockPlugin.Interop;

internal static class NativeLibraryResolver
{
    private const int RTLD_NOW = 2;

    private static IntPtr _mainProgramHandle;

    [ModuleInitializer]
    internal static void Initialize()
    {
        try
        {
            NativeLibrary.SetDllImportResolver(typeof(NativeLibraryResolver).Assembly, Resolve);
        }
        catch (InvalidOperationException)
        {
            // Resolver is already registered for this assembly.
        }
    }

    private static IntPtr Resolve(string libraryName, Assembly _, DllImportSearchPath? __)
    {
        if (!string.Equals(libraryName, "__Internal", StringComparison.Ordinal) &&
            !string.Equals(libraryName, "plugin", StringComparison.Ordinal) &&
            !string.Equals(libraryName, "libplugin", StringComparison.Ordinal))
        {
            return IntPtr.Zero;
        }

        if (_mainProgramHandle != IntPtr.Zero)
        {
            return _mainProgramHandle;
        }

        _mainProgramHandle = dlopen(IntPtr.Zero, RTLD_NOW);
        return _mainProgramHandle;
    }

    [DllImport("libdl.so.2", EntryPoint = "dlopen")]
    private static extern IntPtr dlopen(IntPtr fileName, int flags);
}
