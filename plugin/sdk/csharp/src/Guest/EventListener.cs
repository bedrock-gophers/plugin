namespace BedrockPlugin.Sdk.Guest;

public interface IEventListener<in T> where T : IEventPayload
{
    void Handle(T evt);
}
