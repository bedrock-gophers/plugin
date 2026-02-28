package abi

const (
	Version uint16 = 1
)

const (
	FlagCancellable uint32 = 1 << iota
	FlagSynchronous
)

const EventDescriptorSize = 24

const (
	EventMove uint16 = iota + 1
	EventJump
	EventTeleport
	EventChangeWorld
	EventToggleSprint
	EventToggleSneak
	EventChat
	EventFoodLoss
	EventHeal
	EventHurt
	EventDeath
	EventRespawn
	EventSkinChange
	EventFireExtinguish
	EventStartBreak
	EventBlockBreak
	EventBlockPlace
	EventBlockPick
	EventItemUse
	EventItemUseOnBlock
	EventItemUseOnEntity
	EventItemRelease
	EventItemConsume
	EventAttackEntity
	EventExperienceGain
	EventPunchAir
	EventSignEdit
	EventSleep
	EventLecternPageTurn
	EventItemDamage
	EventItemPickup
	EventHeldSlotChange
	EventItemDrop
	EventTransfer
	EventCommandExecution
	EventQuit
	EventDiagnostics
	EventJoin
	EventPluginCommand
)

func EventName(id uint16) string {
	switch id {
	case EventMove:
		return "HandleMove"
	case EventJump:
		return "HandleJump"
	case EventTeleport:
		return "HandleTeleport"
	case EventChangeWorld:
		return "HandleChangeWorld"
	case EventToggleSprint:
		return "HandleToggleSprint"
	case EventToggleSneak:
		return "HandleToggleSneak"
	case EventChat:
		return "HandleChat"
	case EventFoodLoss:
		return "HandleFoodLoss"
	case EventHeal:
		return "HandleHeal"
	case EventHurt:
		return "HandleHurt"
	case EventDeath:
		return "HandleDeath"
	case EventRespawn:
		return "HandleRespawn"
	case EventSkinChange:
		return "HandleSkinChange"
	case EventFireExtinguish:
		return "HandleFireExtinguish"
	case EventStartBreak:
		return "HandleStartBreak"
	case EventBlockBreak:
		return "HandleBlockBreak"
	case EventBlockPlace:
		return "HandleBlockPlace"
	case EventBlockPick:
		return "HandleBlockPick"
	case EventItemUse:
		return "HandleItemUse"
	case EventItemUseOnBlock:
		return "HandleItemUseOnBlock"
	case EventItemUseOnEntity:
		return "HandleItemUseOnEntity"
	case EventItemRelease:
		return "HandleItemRelease"
	case EventItemConsume:
		return "HandleItemConsume"
	case EventAttackEntity:
		return "HandleAttackEntity"
	case EventExperienceGain:
		return "HandleExperienceGain"
	case EventPunchAir:
		return "HandlePunchAir"
	case EventSignEdit:
		return "HandleSignEdit"
	case EventSleep:
		return "HandleSleep"
	case EventLecternPageTurn:
		return "HandleLecternPageTurn"
	case EventItemDamage:
		return "HandleItemDamage"
	case EventItemPickup:
		return "HandleItemPickup"
	case EventHeldSlotChange:
		return "HandleHeldSlotChange"
	case EventItemDrop:
		return "HandleItemDrop"
	case EventTransfer:
		return "HandleTransfer"
	case EventCommandExecution:
		return "HandleCommandExecution"
	case EventQuit:
		return "HandleQuit"
	case EventDiagnostics:
		return "HandleDiagnostics"
	case EventJoin:
		return "HandleJoin"
	case EventPluginCommand:
		return "HandlePluginCommand"
	default:
		return "unknown"
	}
}
