using BedrockPlugin.Sdk.Abi;
using System.Diagnostics.CodeAnalysis;
using System.Reflection;

namespace BedrockPlugin.Sdk.Guest;

public sealed class CommandContext
{
    private readonly Plugin _plugin;
    private readonly ulong _playerId;
    private readonly IReadOnlyList<string> _rawArgs;

    internal CommandContext(Plugin plugin, ulong playerId, IReadOnlyList<string> rawArgs)
    {
        _plugin = plugin;
        _playerId = playerId;
        _rawArgs = rawArgs;
    }

    public bool IsConsole => _playerId == 0;

    public bool IsPlayer => _playerId != 0;

    public IReadOnlyList<string> RawArgs => _rawArgs;

    public bool TryPlayer(out PlayerRef? player)
    {
        if (_playerId == 0)
        {
            player = null;
            return false;
        }
        player = new PlayerRef(_plugin, _playerId);
        return true;
    }

    public void Message(string message)
    {
        if (_playerId == 0)
        {
            _plugin.Host.ConsoleMessage(_plugin.Name, message);
            return;
        }
        _plugin.Host.PlayerMessage(_playerId, message);
    }

    public void Messagef(string format, params object?[] args)
    {
        Message(TextFormat.Colourf(format, args));
    }
}

public sealed class EventContext
{
    private readonly Plugin _plugin;
    private bool _cancelRequested;
    private MutableArgument<ItemStackData>? _itemDrop;

    internal EventContext(Plugin plugin, EventDescriptor descriptor)
    {
        _plugin = plugin;
        Descriptor = descriptor;
    }

    public EventDescriptor Descriptor { get; }

    public ulong PlayerId => Descriptor.PlayerId;

    public bool IsPlayer => Descriptor.PlayerId != 0;

    public bool TryPlayer(out PlayerRef? player)
    {
        if (Descriptor.PlayerId == 0)
        {
            player = null;
            return false;
        }
        player = new PlayerRef(_plugin, Descriptor.PlayerId);
        return true;
    }

    public PlayerRef Player => new(_plugin, Descriptor.PlayerId);

    public void Message(string message)
    {
        if (Descriptor.PlayerId == 0)
        {
            _plugin.Host.ConsoleMessage(_plugin.Name, message);
            return;
        }
        _plugin.Host.PlayerMessage(Descriptor.PlayerId, message);
    }

    public void Messagef(string format, params object?[] args)
    {
        Message(TextFormat.Colourf(format, args));
    }

    public void Cancel()
    {
        _cancelRequested = true;
    }

    internal MutableArgument<ItemStackData> MutableItemDrop(ItemStackData original)
    {
        _itemDrop ??= new MutableArgument<ItemStackData>(original);
        return _itemDrop;
    }

    internal void CommitMutable()
    {
        if (Descriptor.RequestKey != 0)
        {
            if (_itemDrop is not null && _itemDrop.IsChanged)
            {
                _plugin.Host.EventSetItemDrop(Descriptor.RequestKey, _itemDrop.Value);
            }
            if (_cancelRequested && (Descriptor.Flags & AbiConstants.FlagCancellable) != 0)
            {
                _plugin.Host.EventCancel(Descriptor.RequestKey);
            }
        }
    }
}

public class Plugin
{
    private sealed record CommandHandlerRegistration(
        Func<CommandContext, bool>? Allow,
        Action<CommandContext, IReadOnlyList<string>> Run);

    private sealed record ConventionSpec(
        ushort EventId,
        string MethodName,
        [DynamicallyAccessedMembers(DynamicallyAccessedMemberTypes.PublicProperties)] Type PayloadType);

    private static readonly ConventionSpec[] ConventionSpecs =
    {
        new(EventIds.EventMove, "OnMove", typeof(MovePayload)),
        new(EventIds.EventJump, "OnJump", typeof(JumpPayload)),
        new(EventIds.EventTeleport, "OnTeleport", typeof(TeleportPayload)),
        new(EventIds.EventChangeWorld, "OnChangeWorld", typeof(ChangeWorldPayload)),
        new(EventIds.EventToggleSprint, "OnToggleSprint", typeof(ToggleSprintPayload)),
        new(EventIds.EventToggleSneak, "OnToggleSneak", typeof(ToggleSneakPayload)),
        new(EventIds.EventChat, "OnChat", typeof(ChatPayload)),
        new(EventIds.EventFoodLoss, "OnFoodLoss", typeof(FoodLossPayload)),
        new(EventIds.EventHeal, "OnHeal", typeof(HealPayload)),
        new(EventIds.EventHurt, "OnHurt", typeof(HurtPayload)),
        new(EventIds.EventDeath, "OnDeath", typeof(DeathPayload)),
        new(EventIds.EventRespawn, "OnRespawn", typeof(RespawnPayload)),
        new(EventIds.EventSkinChange, "OnSkinChange", typeof(SkinChangePayload)),
        new(EventIds.EventFireExtinguish, "OnFireExtinguish", typeof(FireExtinguishPayload)),
        new(EventIds.EventStartBreak, "OnStartBreak", typeof(StartBreakPayload)),
        new(EventIds.EventBlockBreak, "OnBlockBreak", typeof(BlockBreakPayload)),
        new(EventIds.EventBlockPlace, "OnBlockPlace", typeof(BlockPlacePayload)),
        new(EventIds.EventBlockPick, "OnBlockPick", typeof(BlockPickPayload)),
        new(EventIds.EventItemUse, "OnItemUse", typeof(ItemUsePayload)),
        new(EventIds.EventItemUseOnBlock, "OnItemUseOnBlock", typeof(ItemUseOnBlockPayload)),
        new(EventIds.EventItemUseOnEntity, "OnItemUseOnEntity", typeof(ItemUseOnEntityPayload)),
        new(EventIds.EventItemRelease, "OnItemRelease", typeof(ItemReleasePayload)),
        new(EventIds.EventItemConsume, "OnItemConsume", typeof(ItemConsumePayload)),
        new(EventIds.EventAttackEntity, "OnAttackEntity", typeof(AttackEntityPayload)),
        new(EventIds.EventExperienceGain, "OnExperienceGain", typeof(ExperienceGainPayload)),
        new(EventIds.EventPunchAir, "OnPunchAir", typeof(PunchAirPayload)),
        new(EventIds.EventSignEdit, "OnSignEdit", typeof(SignEditPayload)),
        new(EventIds.EventSleep, "OnSleep", typeof(SleepPayload)),
        new(EventIds.EventLecternPageTurn, "OnLecternPageTurn", typeof(LecternPageTurnPayload)),
        new(EventIds.EventItemDamage, "OnItemDamage", typeof(ItemDamagePayload)),
        new(EventIds.EventItemPickup, "OnItemPickup", typeof(ItemPickupPayload)),
        new(EventIds.EventHeldSlotChange, "OnHeldSlotChange", typeof(HeldSlotChangePayload)),
        new(EventIds.EventItemDrop, "OnItemDrop", typeof(ItemDropPayload)),
        new(EventIds.EventTransfer, "OnTransfer", typeof(TransferPayload)),
        new(EventIds.EventCommandExecution, "OnCommandExecution", typeof(CommandExecutionPayload)),
        new(EventIds.EventQuit, "OnQuit", typeof(QuitPayload)),
        new(EventIds.EventDiagnostics, "OnDiagnostics", typeof(DiagnosticsPayload)),
        new(EventIds.EventJoin, "OnJoin", typeof(JoinPayload)),
        new(EventIds.EventPluginCommand, "OnPluginCommand", typeof(PluginCommandPayload)),
    };

    private readonly Dictionary<uint, CommandHandlerRegistration> _commandHandlers = new();
    private readonly Dictionary<ushort, List<Action<EventContext, IEventPayload>>> _eventHandlers = new();
    private bool _conventionHandlersRegistered;
    private uint _nextCommandHandlerId = 1;

    public Plugin(string name, IGuestHost host)
    {
        Name = NormalizeToken(name);
        if (Name.Length == 0)
        {
            throw new ArgumentException("plugin name must be non-empty and contain no spaces", nameof(name));
        }
        Host = host ?? throw new ArgumentNullException(nameof(host));
        RegisterConventionHandlers();
    }

    public string Name { get; }

    public IGuestHost Host { get; }

    public void RegisterCommand(string name, string description, IReadOnlyList<string>? aliases, Action<CommandContext> run)
    {
        if (run is null)
        {
            throw new ArgumentNullException(nameof(run));
        }

        RegisterCommand(name, description, aliases, Array.Empty<CommandOverloadSpec>(), allow: null, (ctx, _) => run(ctx));
    }

    public void RegisterCommand(
        string name,
        string description,
        IReadOnlyList<string>? aliases,
        Func<CommandContext, bool> allow,
        Action<CommandContext> run)
    {
        if (allow is null)
        {
            throw new ArgumentNullException(nameof(allow));
        }
        if (run is null)
        {
            throw new ArgumentNullException(nameof(run));
        }

        RegisterCommand(name, description, aliases, Array.Empty<CommandOverloadSpec>(), allow, (ctx, _) => run(ctx));
    }

    public void RegisterCommand(
        string name,
        string description,
        IReadOnlyList<string>? aliases,
        IReadOnlyList<CommandOverloadSpec>? overloads,
        Action<CommandContext, IReadOnlyList<string>> run)
    {
        RegisterCommand(name, description, aliases, overloads, allow: null, run);
    }

    public void RegisterCommand(
        string name,
        string description,
        IReadOnlyList<string>? aliases,
        IReadOnlyList<CommandOverloadSpec>? overloads,
        Func<CommandContext, bool>? allow,
        Action<CommandContext, IReadOnlyList<string>> run)
    {
        if (run is null)
        {
            throw new ArgumentNullException(nameof(run));
        }

        var normalizedName = NormalizeToken(name);
        if (normalizedName.Length == 0)
        {
            throw new ArgumentException("command name must be non-empty and contain no spaces", nameof(name));
        }

        var normalizedAliases = NormalizeAliases(aliases, normalizedName);
        var normalizedDescription = (description ?? string.Empty).Trim();
        var normalizedOverloads = overloads is null ? Array.Empty<CommandOverloadSpec>() : overloads.ToArray();

        var handlerId = _nextCommandHandlerId++;
        _commandHandlers[handlerId] = new CommandHandlerRegistration(allow, run);

        if (Host.RegisterCommand(Name, normalizedName, normalizedDescription, normalizedAliases, handlerId, normalizedOverloads))
        {
            return;
        }

        _commandHandlers.Remove(handlerId);
        throw new InvalidOperationException($"failed to register command {normalizedName}");
    }

    public void HandleEvent(ushort eventId, Action<EventContext, IEventPayload> handler)
    {
        if (handler is null)
        {
            throw new ArgumentNullException(nameof(handler));
        }
        if (!_eventHandlers.TryGetValue(eventId, out var handlers))
        {
            handlers = new List<Action<EventContext, IEventPayload>>();
            _eventHandlers[eventId] = handlers;
        }
        handlers.Add(handler);
    }

    public void HandleMove(Action<EventContext, Vec3, Rotation> handler)
    {
        if (handler is null)
        {
            throw new ArgumentNullException(nameof(handler));
        }

        HandleEvent(EventIds.EventMove, (ctx, payload) =>
        {
            if (payload is MovePayload move)
            {
                handler(ctx, move.NewPos, move.NewRot);
            }
        });
    }

    public void DispatchEvent(EventDescriptor descriptor, ReadOnlySpan<byte> payload)
    {
        if (!EventPayloadDecoder.TryDecode(descriptor.EventId, payload, out var decoded, out var decodeError) || decoded is null)
        {
            if (!string.IsNullOrEmpty(decodeError))
            {
                Host.ConsoleMessage(Name, $"failed to decode {EventIds.EventName(descriptor.EventId)}: {decodeError}");
            }
            return;
        }

        if (descriptor.EventId == EventIds.EventPluginCommand && decoded is PluginCommandPayload command)
        {
            DispatchCommand(descriptor.PlayerId, command);
        }

        if (!_eventHandlers.TryGetValue(descriptor.EventId, out var handlers))
        {
            return;
        }

        var ctx = new EventContext(this, descriptor);
        foreach (var handler in handlers)
        {
            try
            {
                handler(ctx, decoded);
            }
            catch (Exception ex)
            {
                Host.ConsoleMessage(Name, $"event handler panicked for {EventIds.EventName(descriptor.EventId)}: {ex.Message}");
            }
        }
        ctx.CommitMutable();
    }

    public bool TryPlayerByName(string name, out PlayerRef? player)
    {
        name = (name ?? string.Empty).Trim();
        if (name.Length == 0)
        {
            player = null;
            return false;
        }
        var id = Host.ResolvePlayerByName(name);
        if (id == 0)
        {
            player = null;
            return false;
        }
        player = new PlayerRef(this, id);
        return true;
    }

    public IReadOnlyList<string> OnlinePlayerNames()
    {
        return Host.OnlinePlayerNames().ToArray();
    }

    public IReadOnlyList<string> BlockNames()
    {
        return Host.BlockNames().ToArray();
    }

    public IReadOnlyList<string> ItemNames()
    {
        return Host.ItemNames().ToArray();
    }

    public IReadOnlyList<string> WorldNames()
    {
        return Host.WorldNames().ToArray();
    }

    public IReadOnlyList<string> ListPlugins()
    {
        return Host.ManagePlugins(PluginManageAction.List, string.Empty);
    }

    public IReadOnlyList<string> LoadPlugins(string target)
    {
        return Host.ManagePlugins(PluginManageAction.Load, target ?? string.Empty);
    }

    public IReadOnlyList<string> UnloadPlugins(string target)
    {
        return Host.ManagePlugins(PluginManageAction.Unload, target ?? string.Empty);
    }

    public IReadOnlyList<string> ReloadPlugins(string target)
    {
        return Host.ManagePlugins(PluginManageAction.Reload, target ?? string.Empty);
    }

    public void Message(string message)
    {
        Host.ConsoleMessage(Name, message);
    }

    public void Messagef(string format, params object?[] args)
    {
        Message(TextFormat.Colourf(format, args));
    }

    private void DispatchCommand(ulong playerId, PluginCommandPayload payload)
    {
        if (!_commandHandlers.TryGetValue(payload.HandlerId, out var registration))
        {
            return;
        }

        var rawArgs = payload.Args.ToArray();
        var ctx = new CommandContext(this, playerId, rawArgs);
        if (registration.Allow is not null)
        {
            bool allowed;
            try
            {
                allowed = registration.Allow(ctx);
            }
            catch (Exception ex)
            {
                ctx.Messagef("<red>command permission check failed: {0}</red>", ex.Message);
                return;
            }

            if (!allowed)
            {
                ctx.Message("you are not allowed to use this command");
                return;
            }
        }

        try
        {
            registration.Run(ctx, rawArgs);
        }
        catch (Exception ex)
        {
            ctx.Messagef("<red>command failed: {0}</red>", ex.Message);
        }
    }

    private static string NormalizeToken(string value)
    {
        value = (value ?? string.Empty).Trim().ToLowerInvariant();
        if (value.Length == 0)
        {
            return string.Empty;
        }
        if (value.Any(char.IsWhiteSpace))
        {
            return string.Empty;
        }
        return value;
    }

    private static string[] NormalizeAliases(IReadOnlyList<string>? aliases, string commandName)
    {
        if (aliases is null || aliases.Count == 0)
        {
            return Array.Empty<string>();
        }

        var outAliases = new List<string>(aliases.Count);
        var seen = new HashSet<string>(StringComparer.OrdinalIgnoreCase) { commandName };
        foreach (var alias in aliases)
        {
            var token = NormalizeToken(alias);
            if (token.Length == 0)
            {
                continue;
            }
            if (seen.Contains(token))
            {
                continue;
            }
            seen.Add(token);
            outAliases.Add(token);
        }
        return outAliases.ToArray();
    }

    private bool TryRegisterConvention(ConventionSpec spec)
    {
        var flags = BindingFlags.Instance | BindingFlags.Public | BindingFlags.NonPublic;

        if (spec.EventId == EventIds.EventItemDrop)
        {
            var mutableItemDropMethod = GetType().GetMethod(
                spec.MethodName,
                flags,
                binder: null,
                types: new[] { typeof(EventContext), typeof(MutableArgument<ItemStackData>) },
                modifiers: null);
            if (mutableItemDropMethod is not null && mutableItemDropMethod.ReturnType == typeof(void))
            {
                HandleEvent(spec.EventId, CreateMutableItemDropConvention(spec, mutableItemDropMethod));
                return true;
            }
        }

        var payloadMethod = GetType().GetMethod(
            spec.MethodName,
            flags,
            binder: null,
            types: new[] { typeof(EventContext), spec.PayloadType },
            modifiers: null);
        if (payloadMethod is not null && payloadMethod.ReturnType == typeof(void))
        {
            HandleEvent(spec.EventId, CreatePayloadConvention(spec, payloadMethod));
            return true;
        }

        var payloadProps = ConventionPayloadProperties(spec.PayloadType);
        if (payloadProps.Length == 0)
        {
            var emptyMethod = GetType().GetMethod(
                spec.MethodName,
                flags,
                binder: null,
                types: new[] { typeof(EventContext) },
                modifiers: null);
            if (emptyMethod is not null && emptyMethod.ReturnType == typeof(void))
            {
                HandleEvent(spec.EventId, CreateNoPayloadConvention(spec, emptyMethod));
                return true;
            }
        }

        var deconstructedTypes = new Type[payloadProps.Length + 1];
        deconstructedTypes[0] = typeof(EventContext);
        for (var i = 0; i < payloadProps.Length; i++)
        {
            deconstructedTypes[i + 1] = payloadProps[i].PropertyType;
        }
        var deconstructedMethod = GetType().GetMethod(
            spec.MethodName,
            flags,
            binder: null,
            types: deconstructedTypes,
            modifiers: null);
        if (deconstructedMethod is not null && deconstructedMethod.ReturnType == typeof(void))
        {
            HandleEvent(spec.EventId, CreateDeconstructedConvention(spec, payloadProps, deconstructedMethod));
            return true;
        }

        // NativeAOT may not preserve property metadata order in the same way as JIT runtimes.
        // Fall back to parameter-type-based mapping for deconstructed convention methods.
        if (TryResolveDeconstructedConvention(spec, flags, payloadProps, out var fallbackMethod, out var fallbackProps))
        {
            HandleEvent(spec.EventId, CreateDeconstructedConvention(spec, fallbackProps, fallbackMethod));
            return true;
        }

        return false;
    }

    private Action<EventContext, IEventPayload> CreatePayloadConvention(ConventionSpec spec, MethodInfo method)
    {
        return (ctx, payload) =>
        {
            if (!spec.PayloadType.IsInstanceOfType(payload))
            {
                return;
            }
            try
            {
                method.Invoke(this, new object?[] { ctx, payload });
            }
            catch (Exception ex)
            {
                throw UnwrapInvocationException(ex);
            }
        };
    }

    private Action<EventContext, IEventPayload> CreateMutableItemDropConvention(ConventionSpec spec, MethodInfo method)
    {
        return (ctx, payload) =>
        {
            if (!spec.PayloadType.IsInstanceOfType(payload) || payload is not ItemDropPayload itemDrop)
            {
                return;
            }
            try
            {
                method.Invoke(this, new object?[] { ctx, ctx.MutableItemDrop(itemDrop.Item) });
            }
            catch (Exception ex)
            {
                throw UnwrapInvocationException(ex);
            }
        };
    }

    private Action<EventContext, IEventPayload> CreateNoPayloadConvention(ConventionSpec spec, MethodInfo method)
    {
        return (ctx, payload) =>
        {
            if (!spec.PayloadType.IsInstanceOfType(payload))
            {
                return;
            }
            try
            {
                method.Invoke(this, new object?[] { ctx });
            }
            catch (Exception ex)
            {
                throw UnwrapInvocationException(ex);
            }
        };
    }

    private Action<EventContext, IEventPayload> CreateDeconstructedConvention(ConventionSpec spec, PropertyInfo[] payloadProps, MethodInfo method)
    {
        return (ctx, payload) =>
        {
            if (!spec.PayloadType.IsInstanceOfType(payload))
            {
                return;
            }

            var args = new object?[payloadProps.Length + 1];
            args[0] = ctx;
            for (var i = 0; i < payloadProps.Length; i++)
            {
                args[i + 1] = payloadProps[i].GetValue(payload);
            }
            try
            {
                method.Invoke(this, args);
            }
            catch (Exception ex)
            {
                throw UnwrapInvocationException(ex);
            }
        };
    }

    private static PropertyInfo[] ConventionPayloadProperties(
        [DynamicallyAccessedMembers(DynamicallyAccessedMemberTypes.PublicProperties)] Type payloadType)
    {
        return payloadType
            .GetProperties(BindingFlags.Instance | BindingFlags.Public)
            .Where(p => p.CanRead && p.GetMethod is not null)
            .OrderBy(p => p.Name, StringComparer.Ordinal)
            .ToArray();
    }

    private bool TryResolveDeconstructedConvention(
        ConventionSpec spec,
        BindingFlags flags,
        IReadOnlyList<PropertyInfo> payloadProps,
        out MethodInfo method,
        out PropertyInfo[] orderedProps)
    {
        method = null!;
        orderedProps = Array.Empty<PropertyInfo>();
        if (payloadProps.Count == 0)
        {
            return false;
        }

        foreach (var candidate in GetType().GetMethods(flags))
        {
            if (!string.Equals(candidate.Name, spec.MethodName, StringComparison.Ordinal) || candidate.ReturnType != typeof(void))
            {
                continue;
            }

            var parameters = candidate.GetParameters();
            if (parameters.Length != payloadProps.Count + 1 || parameters[0].ParameterType != typeof(EventContext))
            {
                continue;
            }

            if (!TryMapPayloadProperties(parameters, payloadProps, out var mapped))
            {
                continue;
            }

            method = candidate;
            orderedProps = mapped;
            return true;
        }
        return false;
    }

    private static bool TryMapPayloadProperties(ParameterInfo[] parameters, IReadOnlyList<PropertyInfo> payloadProps, out PropertyInfo[] mapped)
    {
        mapped = new PropertyInfo[payloadProps.Count];
        var used = new bool[payloadProps.Count];

        for (var i = 1; i < parameters.Length; i++)
        {
            var parameter = parameters[i];
            var candidates = new List<int>(payloadProps.Count);

            for (var j = 0; j < payloadProps.Count; j++)
            {
                if (used[j])
                {
                    continue;
                }
                if (ParameterMatchesProperty(parameter.ParameterType, payloadProps[j].PropertyType))
                {
                    candidates.Add(j);
                }
            }

            if (candidates.Count == 0)
            {
                return false;
            }

            var selected = candidates.Count == 1 ? candidates[0] : SelectCandidateByName(candidates, payloadProps, parameter.Name);
            if (selected < 0)
            {
                return false;
            }

            used[selected] = true;
            mapped[i - 1] = payloadProps[selected];
        }

        return true;
    }

    private static bool ParameterMatchesProperty(Type parameterType, Type propertyType)
    {
        return parameterType == propertyType || parameterType.IsAssignableFrom(propertyType);
    }

    private static int SelectCandidateByName(IReadOnlyList<int> candidates, IReadOnlyList<PropertyInfo> payloadProps, string? parameterName)
    {
        if (string.IsNullOrWhiteSpace(parameterName))
        {
            return -1;
        }

        var normalized = parameterName.Trim();
        foreach (var idx in candidates)
        {
            if (string.Equals(payloadProps[idx].Name, normalized, StringComparison.OrdinalIgnoreCase))
            {
                return idx;
            }
        }

        normalized = normalized.TrimStart('_');
        foreach (var idx in candidates)
        {
            if (string.Equals(payloadProps[idx].Name, normalized, StringComparison.OrdinalIgnoreCase))
            {
                return idx;
            }
        }

        return -1;
    }

    private static Exception UnwrapInvocationException(Exception ex)
    {
        if (ex is TargetInvocationException tie && tie.InnerException is not null)
        {
            return tie.InnerException;
        }
        return ex;
    }

    private void RegisterConventionHandlers()
    {
        if (_conventionHandlersRegistered)
        {
            return;
        }
        _conventionHandlersRegistered = true;

        foreach (var spec in ConventionSpecs)
        {
            try
            {
                var bound = TryRegisterConvention(spec);
                if (!bound && HasConventionMethodNamed(spec.MethodName))
                {
                    Host.ConsoleMessage(Name, $"ignored {spec.MethodName}: unsupported convention signature; candidates: {DescribeConventionCandidates(spec.MethodName)}");
                }
            }
            catch (Exception ex)
            {
                Host.ConsoleMessage(Name, $"failed to bind {spec.MethodName} convention handler: {ex.Message}");
            }
        }
    }

    private bool HasConventionMethodNamed(string methodName)
    {
        var flags = BindingFlags.Instance | BindingFlags.Public | BindingFlags.NonPublic;
        return GetType().GetMethods(flags).Any(m => string.Equals(m.Name, methodName, StringComparison.Ordinal));
    }

    private string DescribeConventionCandidates(string methodName)
    {
        var flags = BindingFlags.Instance | BindingFlags.Public | BindingFlags.NonPublic;
        var candidates = GetType()
            .GetMethods(flags)
            .Where(m => string.Equals(m.Name, methodName, StringComparison.Ordinal))
            .Select(DescribeMethodSignature)
            .ToArray();
        if (candidates.Length == 0)
        {
            return "none";
        }
        return string.Join(" | ", candidates);
    }

    private static string DescribeMethodSignature(MethodInfo method)
    {
        var parameters = method.GetParameters();
        var args = string.Join(", ", parameters.Select(p => $"{p.ParameterType.Name} {p.Name}"));
        return $"{method.ReturnType.Name} {method.Name}({args})";
    }
}
