using System.Globalization;

namespace BedrockPlugin.Sdk.Guest;

public static class TextFormat
{
    public static string Colourf(string format, params object?[] args)
    {
        if (string.IsNullOrEmpty(format))
        {
            return string.Empty;
        }
        if (args.Length == 0)
        {
            return format;
        }
        return string.Format(CultureInfo.InvariantCulture, format, args);
    }
}
