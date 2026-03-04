using System.Diagnostics.CodeAnalysis;
using System.Reflection;

namespace BedrockPlugin.Sdk.Guest;

public abstract class Command
{
    public abstract string Name { get; }

    public virtual string Description => string.Empty;

    public virtual IReadOnlyList<string> Aliases => Array.Empty<string>();

    public virtual IReadOnlyList<CommandOverloadSpec> Overloads => Array.Empty<CommandOverloadSpec>();

    public virtual bool Allow(CommandContext ctx) => true;

    public abstract void Execute(CommandContext ctx);

    internal virtual void ExecuteRaw(CommandContext ctx, IReadOnlyList<string> rawArgs)
    {
        _ = rawArgs;
        Execute(ctx);
    }
}

public sealed class CommandDispatcher
{
    private readonly Plugin _plugin;

    internal CommandDispatcher(Plugin plugin)
    {
        _plugin = plugin ?? throw new ArgumentNullException(nameof(plugin));
    }

    public void Register(Command command)
    {
        if (command is null)
        {
            throw new ArgumentNullException(nameof(command));
        }

        _plugin.RegisterCommand(
            command.Name,
            command.Description,
            command.Aliases,
            command.Overloads,
            command.Allow,
            command.ExecuteRaw);
    }

    public void Register<TCommand>() where TCommand : Command, new()
    {
        Register(new TCommand());
    }
}

[AttributeUsage(AttributeTargets.Property, AllowMultiple = false, Inherited = true)]
public sealed class CommandParamAttribute : Attribute
{
    public CommandParamAttribute(string name)
    {
        Name = (name ?? string.Empty).Trim();
    }

    public string Name { get; }

    public CommandParameterKind Kind { get; set; } = CommandParameterKind.String;

    public bool Optional { get; set; }

    public int Order { get; set; }
}

public abstract class BoundCommand<[DynamicallyAccessedMembers(DynamicallyAccessedMemberTypes.PublicProperties)] TArgs> : Command where TArgs : class, new()
{
    private static readonly BoundParamSpec[] ParamSpecs = BuildParamSpecs();
    private static readonly CommandArgument[] BoundArguments = ParamSpecs.Select(v => v.Argument).ToArray();
    private static readonly IReadOnlyList<CommandOverloadSpec> BoundOverloads = BuildOverloads();

    public sealed override IReadOnlyList<CommandOverloadSpec> Overloads => BoundOverloads;

    public sealed override void Execute(CommandContext ctx)
    {
        Run(ctx, new TArgs());
    }

    protected abstract void Run(CommandContext ctx, TArgs args);

    internal sealed override void ExecuteRaw(CommandContext ctx, IReadOnlyList<string> rawArgs)
    {
        if (!CommandArgumentBinder.TryBind(BoundArguments, rawArgs, out var values, out var error))
        {
            ctx.Message(error ?? "invalid arguments");
            return;
        }

        var args = new TArgs();
        try
        {
            for (var i = 0; i < ParamSpecs.Length; i++)
            {
                var spec = ParamSpecs[i];
                if (!values.ContainsKey(spec.Argument.Name))
                {
                    continue;
                }
                var converted = ConvertForProperty(spec.Property.PropertyType, values[spec.Argument.Name], spec.Argument.Name);
                spec.Property.SetValue(args, converted);
            }
        }
        catch (Exception ex)
        {
            ctx.Message(ex.Message);
            return;
        }

        Run(ctx.WithValues(values), args);
    }

    private static IReadOnlyList<CommandOverloadSpec> BuildOverloads()
    {
        if (ParamSpecs.Length == 0)
        {
            return Array.Empty<CommandOverloadSpec>();
        }

        var parameters = new CommandParameterSpec[ParamSpecs.Length];
        for (var i = 0; i < ParamSpecs.Length; i++)
        {
            parameters[i] = ParamSpecs[i].Argument.ToSpec();
        }
        return new[] { new CommandOverloadSpec(parameters) };
    }

    private static BoundParamSpec[] BuildParamSpecs()
    {
        var discovered = typeof(TArgs)
            .GetProperties(BindingFlags.Instance | BindingFlags.Public)
            .Select(prop => new { Property = prop, Attribute = prop.GetCustomAttribute<CommandParamAttribute>() })
            .Where(v => v.Attribute is not null)
            .OrderBy(v => v.Attribute!.Order)
            .ThenBy(v => v.Property.Name, StringComparer.Ordinal)
            .ToArray();

        if (discovered.Length == 0)
        {
            return Array.Empty<BoundParamSpec>();
        }

        var outSpecs = new BoundParamSpec[discovered.Length];
        for (var i = 0; i < discovered.Length; i++)
        {
            var property = discovered[i].Property;
            var attribute = discovered[i].Attribute!;
            if (!property.CanWrite || property.GetIndexParameters().Length > 0)
            {
                throw new InvalidOperationException($"{typeof(TArgs).Name}.{property.Name} must be a writable non-indexed property");
            }

            var name = (attribute.Name ?? string.Empty).Trim();
            if (name.Length == 0)
            {
                throw new InvalidOperationException($"{typeof(TArgs).Name}.{property.Name} has an empty command argument name");
            }

            var argument = BuildArgument(name, attribute.Kind, attribute.Optional, property.PropertyType);
            outSpecs[i] = new BoundParamSpec(property, argument);
        }
        return outSpecs;
    }

    private static CommandArgument BuildArgument(string name, CommandParameterKind kind, bool optional, Type propertyType)
    {
        switch (kind)
        {
            case CommandParameterKind.String:
                return new StringArgument(name, optional);
            case CommandParameterKind.Text:
                return new TextArgument(name, optional);
            case CommandParameterKind.Target:
                return new TargetArgument(name, optional);
            case CommandParameterKind.Subcommand:
            case CommandParameterKind.PluginAvailable:
            case CommandParameterKind.PluginLoaded:
                return new TokenArgument(name, kind, optional);
            case CommandParameterKind.Enum:
                {
                    var enumType = Nullable.GetUnderlyingType(propertyType) ?? propertyType;
                    if (!enumType.IsEnum)
                    {
                        throw new InvalidOperationException($"{typeof(TArgs).Name}.{name} uses Enum kind but {propertyType.Name} is not an enum");
                    }
                    return new EnumTokenArgument(name, enumType, optional);
                }
            default:
                return new StringArgument(name, optional);
        }
    }

    private static object? ConvertForProperty(Type targetType, object? value, string argName)
    {
        var nullable = Nullable.GetUnderlyingType(targetType);
        var actualType = nullable ?? targetType;

        if (value is null)
        {
            if (!targetType.IsValueType || nullable is not null)
            {
                return null;
            }
            throw new InvalidOperationException($"argument {argName} cannot be null");
        }

        if (actualType.IsInstanceOfType(value))
        {
            return value;
        }

        if (actualType.IsEnum)
        {
            if (value is string text)
            {
                return Enum.Parse(actualType, text, true);
            }
            return Enum.ToObject(actualType, value);
        }

        try
        {
            return Convert.ChangeType(value, actualType);
        }
        catch (Exception ex)
        {
            throw new InvalidOperationException($"argument {argName} cannot be converted to {targetType.Name}: {ex.Message}", ex);
        }
    }

    private readonly record struct BoundParamSpec(PropertyInfo Property, CommandArgument Argument);
}

public abstract class CommandArgument
{
    private readonly string[] _enumOptions;

    protected CommandArgument(string name, CommandParameterKind kind, bool optional, IReadOnlyList<string>? enumOptions = null)
    {
        Name = (name ?? string.Empty).Trim();
        if (Name.Length == 0)
        {
            throw new ArgumentException("argument name cannot be empty", nameof(name));
        }

        Kind = kind;
        Optional = optional;
        _enumOptions = enumOptions is null ? Array.Empty<string>() : enumOptions.Where(v => !string.IsNullOrWhiteSpace(v)).ToArray();
    }

    public string Name { get; }

    public bool Optional { get; }

    public CommandParameterKind Kind { get; }

    internal virtual bool ConsumeRemainder => false;

    internal CommandParameterSpec ToSpec()
    {
        return new CommandParameterSpec(Name, Kind, Optional, _enumOptions);
    }

    internal abstract bool TryParse(string raw, out object? value, out string? error);
}

public sealed class StringArgument : CommandArgument
{
    public StringArgument(string name, bool optional = false) : base(name, CommandParameterKind.String, optional) { }

    internal override bool TryParse(string raw, out object? value, out string? error)
    {
        value = raw;
        error = null;
        return true;
    }
}

public sealed class TextArgument : CommandArgument
{
    public TextArgument(string name, bool optional = false) : base(name, CommandParameterKind.Text, optional) { }

    internal override bool ConsumeRemainder => true;

    internal override bool TryParse(string raw, out object? value, out string? error)
    {
        value = raw;
        error = null;
        return true;
    }
}

public sealed class TargetArgument : CommandArgument
{
    public TargetArgument(string name = "target", bool optional = false) : base(name, CommandParameterKind.Target, optional) { }

    internal override bool TryParse(string raw, out object? value, out string? error)
    {
        value = raw;
        error = null;
        return true;
    }
}

public sealed class TokenArgument : CommandArgument
{
    public TokenArgument(string name, CommandParameterKind kind, bool optional = false) : base(name, kind, optional) { }

    internal override bool TryParse(string raw, out object? value, out string? error)
    {
        value = raw;
        error = null;
        return true;
    }
}

public sealed class EnumArgument<TEnum> : CommandArgument where TEnum : struct, Enum
{
    public EnumArgument(string name, bool optional = false) : base(name, CommandParameterKind.Enum, optional, Enum.GetNames<TEnum>()) { }

    internal override bool TryParse(string raw, out object? value, out string? error)
    {
        if (Enum.TryParse<TEnum>(raw, ignoreCase: true, out var parsed))
        {
            value = parsed;
            error = null;
            return true;
        }

        value = default(TEnum);
        error = $"unknown value {raw}";
        return false;
    }
}

public sealed class EnumTokenArgument : CommandArgument
{
    private readonly Type _enumType;

    public EnumTokenArgument(string name, Type enumType, bool optional = false)
        : base(name, CommandParameterKind.Enum, optional, Enum.GetNames(enumType))
    {
        _enumType = enumType ?? throw new ArgumentNullException(nameof(enumType));
        if (!_enumType.IsEnum)
        {
            throw new ArgumentException($"{_enumType.Name} is not an enum type", nameof(enumType));
        }
    }

    internal override bool TryParse(string raw, out object? value, out string? error)
    {
        try
        {
            value = Enum.Parse(_enumType, raw, true);
            error = null;
            return true;
        }
        catch
        {
            value = null;
            error = $"unknown value {raw}";
            return false;
        }
    }
}

public sealed class CommandBuilder
{
    private readonly List<CommandArgument> _arguments = new();
    private readonly List<string> _aliases = new();
    private string _description = string.Empty;

    public CommandBuilder Description(string description)
    {
        _description = (description ?? string.Empty).Trim();
        return this;
    }

    public CommandBuilder Alias(string alias)
    {
        if (!string.IsNullOrWhiteSpace(alias))
        {
            _aliases.Add(alias.Trim());
        }
        return this;
    }

    public CommandBuilder Aliases(params string[] aliases)
    {
        if (aliases is null)
        {
            return this;
        }
        for (var i = 0; i < aliases.Length; i++)
        {
            Alias(aliases[i]);
        }
        return this;
    }

    public CommandBuilder Argument(CommandArgument argument)
    {
        if (argument is null)
        {
            throw new ArgumentNullException(nameof(argument));
        }
        _arguments.Add(argument);
        return this;
    }

    public Command Build(string name, Action<CommandContext> execute)
    {
        if (execute is null)
        {
            throw new ArgumentNullException(nameof(execute));
        }
        return Build(name, _ => true, execute);
    }

    public Command Build(string name, Func<CommandContext, bool> allow, Action<CommandContext> execute)
    {
        if (allow is null)
        {
            throw new ArgumentNullException(nameof(allow));
        }
        if (execute is null)
        {
            throw new ArgumentNullException(nameof(execute));
        }

        return new BuiltCommand(name, _description, _aliases, _arguments, allow, execute);
    }

    private sealed class BuiltCommand : Command
    {
        private readonly string _name;
        private readonly string _description;
        private readonly string[] _aliases;
        private readonly CommandArgument[] _arguments;
        private readonly Func<CommandContext, bool> _allow;
        private readonly Action<CommandContext> _execute;
        private readonly CommandOverloadSpec[] _overloads;

        internal BuiltCommand(
            string name,
            string description,
            IReadOnlyList<string> aliases,
            IReadOnlyList<CommandArgument> arguments,
            Func<CommandContext, bool> allow,
            Action<CommandContext> execute)
        {
            _name = (name ?? string.Empty).Trim();
            _description = description ?? string.Empty;
            _aliases = aliases is null ? Array.Empty<string>() : aliases.ToArray();
            _arguments = arguments is null ? Array.Empty<CommandArgument>() : arguments.ToArray();
            _allow = allow;
            _execute = execute;

            if (_arguments.Length == 0)
            {
                _overloads = Array.Empty<CommandOverloadSpec>();
            }
            else
            {
                var specs = new CommandParameterSpec[_arguments.Length];
                for (var i = 0; i < _arguments.Length; i++)
                {
                    specs[i] = _arguments[i].ToSpec();
                }
                _overloads = new[] { new CommandOverloadSpec(specs) };
            }
        }

        public override string Name => _name;

        public override string Description => _description;

        public override IReadOnlyList<string> Aliases => _aliases;

        public override IReadOnlyList<CommandOverloadSpec> Overloads => _overloads;

        public override bool Allow(CommandContext ctx) => _allow(ctx);

        public override void Execute(CommandContext ctx) => _execute(ctx);

        internal override void ExecuteRaw(CommandContext ctx, IReadOnlyList<string> rawArgs)
        {
            if (!CommandArgumentBinder.TryBind(_arguments, rawArgs, out var values, out var error))
            {
                ctx.Message(error ?? "invalid arguments");
                return;
            }
            _execute(ctx.WithValues(values));
        }
    }
}

internal static class CommandArgumentBinder
{
    internal static bool TryBind(
        IReadOnlyList<CommandArgument> arguments,
        IReadOnlyList<string> rawArgs,
        out IReadOnlyDictionary<string, object?> values,
        out string? error)
    {
        var outValues = new Dictionary<string, object?>(StringComparer.OrdinalIgnoreCase);
        var index = 0;

        for (var i = 0; i < arguments.Count; i++)
        {
            var argument = arguments[i];
            var missing = index >= rawArgs.Count;

            if (argument.ConsumeRemainder)
            {
                var joined = string.Empty;
                if (!missing)
                {
                    joined = string.Join(" ", rawArgs.Skip(index));
                    index = rawArgs.Count;
                }

                if (string.IsNullOrWhiteSpace(joined))
                {
                    if (argument.Optional)
                    {
                        continue;
                    }
                    values = outValues;
                    error = $"missing required argument {argument.Name}";
                    return false;
                }

                if (!argument.TryParse(joined, out var parsedRemainder, out error))
                {
                    values = outValues;
                    error = $"invalid {argument.Name}: {error}";
                    return false;
                }
                outValues[argument.Name] = parsedRemainder;
                continue;
            }

            if (missing)
            {
                if (argument.Optional)
                {
                    continue;
                }
                values = outValues;
                error = $"missing required argument {argument.Name}";
                return false;
            }

            var token = rawArgs[index++];
            if (!argument.TryParse(token, out var parsed, out error))
            {
                values = outValues;
                error = $"invalid {argument.Name}: {error}";
                return false;
            }
            outValues[argument.Name] = parsed;
        }

        if (index < rawArgs.Count)
        {
            values = outValues;
            error = "too many arguments";
            return false;
        }

        values = outValues;
        error = null;
        return true;
    }
}
