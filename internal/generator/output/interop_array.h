#ifndef GULP_ARRAY_H
#define GULP_ARRAY_H

#include <stdbool.h>
#include <stdint.h>
#include <stddef.h>
#include <stdlib.h>

typedef struct {
    void* data;
    int length;
    int elementSize;
} GoArray;

typedef struct {
    uint32_t v0;
    uint32_t v1;
    uint32_t v2;
    uint32_t v3;
} Color_RGBA64_RGBA_Result;

typedef struct {
    GoArray v0;
    uint64_t v1;
} Dialogue_Button_MarshalJSON_Result;

typedef struct {
    GoArray v0;
    uint64_t v1;
} Dialogue_Dialogue_MarshalJSON_Result;

typedef struct {
    GoArray v0;
    uint64_t v1;
} Dialogue_DisplaySettings_MarshalJSON_Result;

typedef struct {
    int v0;
    uint64_t v1;
} Inventory_Inventory_AddItem_Result;

typedef struct {
    int v0;
    bool v1;
} Inventory_Inventory_First_Result;

typedef struct {
    int v0;
    bool v1;
} Inventory_Inventory_FirstEmpty_Result;

typedef struct {
    int v0;
    bool v1;
} Inventory_Inventory_FirstFunc_Result;

typedef struct {
    uint64_t v0;
    uint64_t v1;
} Inventory_Inventory_Item_Result;

typedef struct {
    uint64_t v0;
    uint64_t v1;
} Item_Stack_AddStack_Result;

typedef struct {
    uint64_t v0;
    bool v1;
} Item_Stack_Enchantment_Result;

typedef struct {
    uint64_t v0;
    bool v1;
} Item_Stack_Value_Result;

typedef struct {
    uint64_t v0;
    uint64_t v1;
} Language_Region_TLD_Result;

typedef struct {
    uint64_t v0;
    int v1;
} Language_Tag_Base_Result;

typedef struct {
    uint64_t v0;
    bool v1;
} Language_Tag_Extension_Result;

typedef struct {
    GoArray v0;
    uint64_t v1;
} Language_Tag_MarshalText_Result;

typedef struct {
    uint64_t v0;
    uint64_t v1;
    uint64_t v2;
} Language_Tag_Raw_Result;

typedef struct {
    uint64_t v0;
    int v1;
} Language_Tag_Region_Result;

typedef struct {
    uint64_t v0;
    int v1;
} Language_Tag_Script_Result;

typedef struct {
    uint64_t v0;
    uint64_t v1;
} Language_Tag_SetTypeForKey_Result;

typedef struct {
    int v0;
    bool v1;
} Player_Player_Collect_Result;

typedef struct {
    GoArray v0;
    uint64_t v1;
    bool v2;
} Player_Player_DeathPosition_Result;

typedef struct {
    uint64_t v0;
    bool v1;
} Player_Player_Effect_Result;

typedef struct {
    uint64_t v0;
    uint64_t v1;
} Player_Player_HeldItems_Result;

typedef struct {
    double v0;
    bool v1;
} Player_Player_Hurt_Result;

typedef struct {
    GoArray v0;
    bool v1;
} Player_Player_Sleeping_Result;

typedef struct {
    int v0;
    uint64_t v1;
} Scoreboard_Scoreboard_Write_Result;

typedef struct {
    int v0;
    uint64_t v1;
} Scoreboard_Scoreboard_WriteString_Result;

typedef struct {
    uint64_t v0;
    bool v1;
} World_EntityHandle_Entity_Result;

typedef struct {
    uint64_t v0;
    bool v1;
} World_EntityRegistry_Lookup_Result;

typedef struct {
    uint64_t v0;
    bool v1;
} World_Tx_Liquid_Result;

#ifdef __cplusplus
extern "C" {
#endif

void Go_ReleaseHandle(uint64_t handle);
void C_FreeString(char* value);
void C_FreeMemory(void* value);

uint64_t Block_ExplosionConfig_Ctor(void);
void Block_ExplosionConfig_Explode(uint64_t self, uint64_t p1, uint64_t p2);

uint64_t Bossbar_BossBar_Ctor(void);
uint64_t Bossbar_BossBar_Colour(uint64_t self);
double Bossbar_BossBar_HealthPercentage(uint64_t self);
char* Bossbar_BossBar_Text(uint64_t self);
uint64_t Bossbar_BossBar_WithColour(uint64_t self, uint64_t p1);
uint64_t Bossbar_BossBar_WithHealthPercentage(uint64_t self, double p1);

uint64_t Bossbar_Colour_Ctor(void);
uint8_t Bossbar_Colour_Uint8(uint64_t self);

uint64_t Chat_Translation_Ctor(void);
uint64_t Chat_Translation_Enc(uint64_t self, char* p1);
uint64_t Chat_Translation_F(uint64_t self, uint64_t p1);
char* Chat_Translation_Resolve(uint64_t self, uint64_t p1);
bool Chat_Translation_Zero(uint64_t self);

uint64_t Cmd_Output_Ctor(void);
void Cmd_Output_Error(uint64_t self, uint64_t p1);
int Cmd_Output_ErrorCount(uint64_t self);
void Cmd_Output_Errorf(uint64_t self, char* p1, uint64_t p2);
GoArray Cmd_Output_Errors(uint64_t self);
void Cmd_Output_Errort(uint64_t self, uint64_t p1, uint64_t p2);
int Cmd_Output_MessageCount(uint64_t self);
GoArray Cmd_Output_Messages(uint64_t self);
void Cmd_Output_Print(uint64_t self, uint64_t p1);
void Cmd_Output_Printf(uint64_t self, char* p1, uint64_t p2);
void Cmd_Output_Printt(uint64_t self, uint64_t p1, uint64_t p2);

uint64_t Color_RGBA64_Ctor(void);
Color_RGBA64_RGBA_Result Color_RGBA64_RGBA(uint64_t self);

uint64_t Cube_BBox_Ctor(void);
GoArray Cube_BBox_Corners(uint64_t self);
uint64_t Cube_BBox_Extend(uint64_t self, uint64_t p1);
uint64_t Cube_BBox_ExtendTowards(uint64_t self, int p1, double p2);
uint64_t Cube_BBox_Grow(uint64_t self, double p1);
uint64_t Cube_BBox_GrowVec3(uint64_t self, uint64_t p1);
double Cube_BBox_Height(uint64_t self);
bool Cube_BBox_IntersectsWith(uint64_t self, uint64_t p1);
double Cube_BBox_Length(uint64_t self);
GoArray Cube_BBox_Max(uint64_t self);
GoArray Cube_BBox_Min(uint64_t self);
uint64_t Cube_BBox_Mul(uint64_t self, double p1);
uint64_t Cube_BBox_Stretch(uint64_t self, int p1, double p2);
uint64_t Cube_BBox_Translate(uint64_t self, uint64_t p1);
uint64_t Cube_BBox_TranslateTowards(uint64_t self, int p1, double p2);
bool Cube_BBox_Vec3Within(uint64_t self, uint64_t p1);
bool Cube_BBox_Vec3WithinXY(uint64_t self, uint64_t p1);
bool Cube_BBox_Vec3WithinXZ(uint64_t self, uint64_t p1);
bool Cube_BBox_Vec3WithinYZ(uint64_t self, uint64_t p1);
double Cube_BBox_Volume(uint64_t self);
double Cube_BBox_Width(uint64_t self);
double Cube_BBox_XOffset(uint64_t self, uint64_t p1, double p2);
double Cube_BBox_YOffset(uint64_t self, uint64_t p1, double p2);
double Cube_BBox_ZOffset(uint64_t self, uint64_t p1, double p2);

uint64_t Dialogue_Button_Ctor(void);
Dialogue_Button_MarshalJSON_Result Dialogue_Button_MarshalJSON(uint64_t self);

uint64_t Dialogue_Dialogue_Ctor(void);
char* Dialogue_Dialogue_Body(uint64_t self);
GoArray Dialogue_Dialogue_Buttons(uint64_t self);
void Dialogue_Dialogue_Close(uint64_t self, uint64_t p1, uint64_t p2);
uint64_t Dialogue_Dialogue_Display(uint64_t self);
Dialogue_Dialogue_MarshalJSON_Result Dialogue_Dialogue_MarshalJSON(uint64_t self);
uint64_t Dialogue_Dialogue_Submit(uint64_t self, uint64_t p1, uint64_t p2, uint64_t p3);
char* Dialogue_Dialogue_Title(uint64_t self);
uint64_t Dialogue_Dialogue_WithBody(uint64_t self, uint64_t p1);
uint64_t Dialogue_Dialogue_WithButtons(uint64_t self, uint64_t p1);
uint64_t Dialogue_Dialogue_WithDisplay(uint64_t self, uint64_t p1);

uint64_t Dialogue_DisplaySettings_Ctor(void);
Dialogue_DisplaySettings_MarshalJSON_Result Dialogue_DisplaySettings_MarshalJSON(uint64_t self);

uint64_t Effect_Effect_Ctor(void);
bool Effect_Effect_Ambient(uint64_t self);
int64_t Effect_Effect_Duration(uint64_t self);
bool Effect_Effect_Infinite(uint64_t self);
int Effect_Effect_Level(uint64_t self);
bool Effect_Effect_ParticlesHidden(uint64_t self);
int Effect_Effect_Tick(uint64_t self);
uint64_t Effect_Effect_TickDuration(uint64_t self);
uint64_t Effect_Effect_Type(uint64_t self);
uint64_t Effect_Effect_WithoutParticles(uint64_t self);

uint64_t Hud_Element_Ctor(void);
uint8_t Hud_Element_Uint8(uint64_t self);

uint64_t Image_Point_Ctor(void);
uint64_t Image_Point_Add(uint64_t self, uint64_t p1);
uint64_t Image_Point_Div(uint64_t self, int p1);
bool Image_Point_Eq(uint64_t self, uint64_t p1);
bool Image_Point_In(uint64_t self, uint64_t p1);
uint64_t Image_Point_Mod(uint64_t self, uint64_t p1);
uint64_t Image_Point_Mul(uint64_t self, int p1);
char* Image_Point_String(uint64_t self);
uint64_t Image_Point_Sub(uint64_t self, uint64_t p1);

uint64_t Image_Rectangle_Ctor(void);
uint64_t Image_Rectangle_Add(uint64_t self, uint64_t p1);
uint64_t Image_Rectangle_At(uint64_t self, int p1, int p2);
uint64_t Image_Rectangle_Bounds(uint64_t self);
uint64_t Image_Rectangle_Canon(uint64_t self);
uint64_t Image_Rectangle_ColorModel(uint64_t self);
int Image_Rectangle_Dx(uint64_t self);
int Image_Rectangle_Dy(uint64_t self);
bool Image_Rectangle_Empty(uint64_t self);
bool Image_Rectangle_Eq(uint64_t self, uint64_t p1);
bool Image_Rectangle_In(uint64_t self, uint64_t p1);
uint64_t Image_Rectangle_Inset(uint64_t self, int p1);
uint64_t Image_Rectangle_Intersect(uint64_t self, uint64_t p1);
bool Image_Rectangle_Overlaps(uint64_t self, uint64_t p1);
uint64_t Image_Rectangle_RGBA64At(uint64_t self, int p1, int p2);
uint64_t Image_Rectangle_Size(uint64_t self);
char* Image_Rectangle_String(uint64_t self);
uint64_t Image_Rectangle_Sub(uint64_t self, uint64_t p1);
uint64_t Image_Rectangle_Union(uint64_t self, uint64_t p1);

uint64_t Inventory_Armour_Ctor(void);
uint64_t Inventory_Armour_Boots(uint64_t self);
uint64_t Inventory_Armour_Chestplate(uint64_t self);
GoArray Inventory_Armour_Clear(uint64_t self);
uint64_t Inventory_Armour_Close(uint64_t self);
void Inventory_Armour_Damage(uint64_t self, double p1, uint64_t p2);
double Inventory_Armour_DamageReduction(uint64_t self, double p1, uint64_t p2);
void Inventory_Armour_Handle(uint64_t self, uint64_t p1);
uint64_t Inventory_Armour_Helmet(uint64_t self);
int Inventory_Armour_HighestEnchantmentLevel(uint64_t self, uint64_t p1);
uint64_t Inventory_Armour_Inventory(uint64_t self);
GoArray Inventory_Armour_Items(uint64_t self);
double Inventory_Armour_KnockBackResistance(uint64_t self);
uint64_t Inventory_Armour_Leggings(uint64_t self);
void Inventory_Armour_Set(uint64_t self, uint64_t p1, uint64_t p2, uint64_t p3, uint64_t p4);
void Inventory_Armour_SetBoots(uint64_t self, uint64_t p1);
void Inventory_Armour_SetChestplate(uint64_t self, uint64_t p1);
void Inventory_Armour_SetHelmet(uint64_t self, uint64_t p1);
void Inventory_Armour_SetLeggings(uint64_t self, uint64_t p1);
GoArray Inventory_Armour_Slots(uint64_t self);
char* Inventory_Armour_String(uint64_t self);
double Inventory_Armour_ThornsDamage(uint64_t self, uint64_t p1);

uint64_t Inventory_Inventory_Ctor(void);
Inventory_Inventory_AddItem_Result Inventory_Inventory_AddItem(uint64_t self, uint64_t p1);
GoArray Inventory_Inventory_Clear(uint64_t self);
uint64_t Inventory_Inventory_Clone(uint64_t self, uint64_t p1);
uint64_t Inventory_Inventory_Close(uint64_t self);
bool Inventory_Inventory_ContainsItem(uint64_t self, uint64_t p1);
bool Inventory_Inventory_ContainsItemFunc(uint64_t self, int p1, uint64_t p2);
bool Inventory_Inventory_Empty(uint64_t self);
Inventory_Inventory_First_Result Inventory_Inventory_First(uint64_t self, uint64_t p1);
Inventory_Inventory_FirstEmpty_Result Inventory_Inventory_FirstEmpty(uint64_t self);
Inventory_Inventory_FirstFunc_Result Inventory_Inventory_FirstFunc(uint64_t self, uint64_t p1);
void Inventory_Inventory_Handle(uint64_t self, uint64_t p1);
uint64_t Inventory_Inventory_Handler(uint64_t self);
Inventory_Inventory_Item_Result Inventory_Inventory_Item(uint64_t self, int p1);
GoArray Inventory_Inventory_Items(uint64_t self);
uint64_t Inventory_Inventory_Merge(uint64_t self, uint64_t p1, uint64_t p2);
uint64_t Inventory_Inventory_RemoveItem(uint64_t self, uint64_t p1);
uint64_t Inventory_Inventory_RemoveItemFunc(uint64_t self, int p1, uint64_t p2);
uint64_t Inventory_Inventory_SetItem(uint64_t self, int p1, uint64_t p2);
int Inventory_Inventory_Size(uint64_t self);
void Inventory_Inventory_SlotFunc(uint64_t self, uint64_t p1);
void Inventory_Inventory_SlotValidatorFunc(uint64_t self, uint64_t p1);
GoArray Inventory_Inventory_Slots(uint64_t self);
char* Inventory_Inventory_String(uint64_t self);
uint64_t Inventory_Inventory_Swap(uint64_t self, int p1, int p2);

uint64_t Item_Enchantment_Ctor(void);
int Item_Enchantment_Level(uint64_t self);
uint64_t Item_Enchantment_Type(uint64_t self);

uint64_t Item_Stack_Ctor(void);
Item_Stack_AddStack_Result Item_Stack_AddStack(uint64_t self, uint64_t p1);
int Item_Stack_AnvilCost(uint64_t self);
uint64_t Item_Stack_AsBreakable(uint64_t self);
uint64_t Item_Stack_AsUnbreakable(uint64_t self);
double Item_Stack_AttackDamage(uint64_t self);
bool Item_Stack_Comparable(uint64_t self, uint64_t p1);
int Item_Stack_Count(uint64_t self);
char* Item_Stack_CustomName(uint64_t self);
uint64_t Item_Stack_Damage(uint64_t self, int p1);
int Item_Stack_Durability(uint64_t self);
bool Item_Stack_Empty(uint64_t self);
Item_Stack_Enchantment_Result Item_Stack_Enchantment(uint64_t self, uint64_t p1);
GoArray Item_Stack_Enchantments(uint64_t self);
bool Item_Stack_Equal(uint64_t self, uint64_t p1);
uint64_t Item_Stack_Grow(uint64_t self, int p1);
uint64_t Item_Stack_Item(uint64_t self);
GoArray Item_Stack_Lore(uint64_t self);
int Item_Stack_MaxCount(uint64_t self);
int Item_Stack_MaxDurability(uint64_t self);
char* Item_Stack_String(uint64_t self);
bool Item_Stack_Unbreakable(uint64_t self);
Item_Stack_Value_Result Item_Stack_Value(uint64_t self, char* p1);
uint64_t Item_Stack_Values(uint64_t self);
uint64_t Item_Stack_WithAnvilCost(uint64_t self, int p1);
uint64_t Item_Stack_WithCustomName(uint64_t self, uint64_t p1);
uint64_t Item_Stack_WithDurability(uint64_t self, int p1);
uint64_t Item_Stack_WithEnchantments(uint64_t self, uint64_t p1);
uint64_t Item_Stack_WithItem(uint64_t self, uint64_t p1);
uint64_t Item_Stack_WithLore(uint64_t self, uint64_t p1);
uint64_t Item_Stack_WithValue(uint64_t self, char* p1, uint64_t p2);
uint64_t Item_Stack_WithoutEnchantments(uint64_t self, uint64_t p1);

uint64_t Item_UseContext_Ctor(void);
void Item_UseContext_Consume(uint64_t self, uint64_t p1);
void Item_UseContext_DamageItem(uint64_t self, int p1);
void Item_UseContext_SubtractFromCount(uint64_t self, int p1);

uint64_t Language_Base_Ctor(void);
char* Language_Base_ISO3(uint64_t self);
bool Language_Base_IsPrivateUse(uint64_t self);
char* Language_Base_String(uint64_t self);

uint64_t Language_Extension_Ctor(void);
char* Language_Extension_String(uint64_t self);
GoArray Language_Extension_Tokens(uint64_t self);
uint8_t Language_Extension_Type(uint64_t self);

uint64_t Language_Region_Ctor(void);
uint64_t Language_Region_Canonicalize(uint64_t self);
bool Language_Region_Contains(uint64_t self, uint64_t p1);
char* Language_Region_ISO3(uint64_t self);
bool Language_Region_IsCountry(uint64_t self);
bool Language_Region_IsGroup(uint64_t self);
bool Language_Region_IsPrivateUse(uint64_t self);
int Language_Region_M49(uint64_t self);
char* Language_Region_String(uint64_t self);
Language_Region_TLD_Result Language_Region_TLD(uint64_t self);

uint64_t Language_Script_Ctor(void);
bool Language_Script_IsPrivateUse(uint64_t self);
char* Language_Script_String(uint64_t self);

uint64_t Language_Tag_Ctor(void);
Language_Tag_Base_Result Language_Tag_Base(uint64_t self);
Language_Tag_Extension_Result Language_Tag_Extension(uint64_t self, uint8_t p1);
GoArray Language_Tag_Extensions(uint64_t self);
bool Language_Tag_IsRoot(uint64_t self);
Language_Tag_MarshalText_Result Language_Tag_MarshalText(uint64_t self);
uint64_t Language_Tag_Parent(uint64_t self);
Language_Tag_Raw_Result Language_Tag_Raw(uint64_t self);
Language_Tag_Region_Result Language_Tag_Region(uint64_t self);
Language_Tag_Script_Result Language_Tag_Script(uint64_t self);
Language_Tag_SetTypeForKey_Result Language_Tag_SetTypeForKey(uint64_t self, char* p1, char* p2);
char* Language_Tag_String(uint64_t self);
char* Language_Tag_TypeForKey(uint64_t self, char* p1);
uint64_t Language_Tag_UnmarshalText(uint64_t self, uint64_t p1);
GoArray Language_Tag_Variants(uint64_t self);

uint64_t Language_Variant_Ctor(void);
char* Language_Variant_String(uint64_t self);

uint64_t Player_Config_Ctor(void);
void Player_Config_Apply(uint64_t self, uint64_t p1);

uint64_t Player_Player_Ctor(void);
void Player_Player_AbortBreaking(uint64_t self);
double Player_Player_Absorption(uint64_t self);
void Player_Player_AddDebugShape(uint64_t self, uint64_t p1);
void Player_Player_AddEffect(uint64_t self, uint64_t p1);
int Player_Player_AddExperience(uint64_t self, int p1);
void Player_Player_AddFood(uint64_t self, int p1);
uint64_t Player_Player_Addr(uint64_t self);
int64_t Player_Player_AirSupply(uint64_t self);
uint64_t Player_Player_Armour(uint64_t self);
bool Player_Player_AttackEntity(uint64_t self, uint64_t p1);
bool Player_Player_BeaconAffected(uint64_t self);
void Player_Player_BreakBlock(uint64_t self, uint64_t p1);
bool Player_Player_Breathing(uint64_t self);
void Player_Player_Chat(uint64_t self, uint64_t p1);
uint64_t Player_Player_Close(uint64_t self);
void Player_Player_CloseDialogue(uint64_t self);
void Player_Player_CloseForm(uint64_t self);
Player_Player_Collect_Result Player_Player_Collect(uint64_t self, uint64_t p1);
bool Player_Player_CollectExperience(uint64_t self, int p1);
void Player_Player_ContinueBreaking(uint64_t self, int p1);
bool Player_Player_Crawling(uint64_t self);
uint64_t Player_Player_Data(uint64_t self);
bool Player_Player_Dead(uint64_t self);
Player_Player_DeathPosition_Result Player_Player_DeathPosition(uint64_t self);
char* Player_Player_DeviceID(uint64_t self);
char* Player_Player_DeviceModel(uint64_t self);
void Player_Player_DisableInstantRespawn(uint64_t self);
void Player_Player_Disconnect(uint64_t self, uint64_t p1);
int Player_Player_Drop(uint64_t self, uint64_t p1);
uint64_t Player_Player_EditSign(uint64_t self, uint64_t p1, char* p2, char* p3);
Player_Player_Effect_Result Player_Player_Effect(uint64_t self, uint64_t p1);
GoArray Player_Player_Effects(uint64_t self);
void Player_Player_EnableInstantRespawn(uint64_t self);
int64_t Player_Player_EnchantmentSeed(uint64_t self);
uint64_t Player_Player_EnderChestInventory(uint64_t self);
void Player_Player_ExecuteCommand(uint64_t self, char* p1);
void Player_Player_Exhaust(uint64_t self, double p1);
int Player_Player_Experience(uint64_t self);
int Player_Player_ExperienceLevel(uint64_t self);
double Player_Player_ExperienceProgress(uint64_t self);
void Player_Player_Explode(uint64_t self, uint64_t p1, double p2, uint64_t p3);
void Player_Player_Extinguish(uint64_t self);
double Player_Player_EyeHeight(uint64_t self);
double Player_Player_FallDistance(uint64_t self);
double Player_Player_FinalDamageFrom(uint64_t self, double p1, uint64_t p2);
void Player_Player_FinishBreaking(uint64_t self);
bool Player_Player_FireProof(uint64_t self);
double Player_Player_FlightSpeed(uint64_t self);
bool Player_Player_Flying(uint64_t self);
int Player_Player_Food(uint64_t self);
uint64_t Player_Player_GameMode(uint64_t self);
bool Player_Player_Gliding(uint64_t self);
uint64_t Player_Player_H(uint64_t self);
void Player_Player_Handle(uint64_t self, uint64_t p1);
uint64_t Player_Player_Handler(uint64_t self);
bool Player_Player_HasCooldown(uint64_t self, uint64_t p1);
void Player_Player_Heal(uint64_t self, double p1, uint64_t p2);
double Player_Player_Health(uint64_t self);
Player_Player_HeldItems_Result Player_Player_HeldItems(uint64_t self);
void Player_Player_HideCoordinates(uint64_t self);
void Player_Player_HideEntity(uint64_t self, uint64_t p1);
void Player_Player_HideHudElement(uint64_t self, uint64_t p1);
bool Player_Player_HudElementHidden(uint64_t self, uint64_t p1);
Player_Player_Hurt_Result Player_Player_Hurt(uint64_t self, double p1, uint64_t p2);
bool Player_Player_Immobile(uint64_t self);
uint64_t Player_Player_Inventory(uint64_t self);
bool Player_Player_Invisible(uint64_t self);
void Player_Player_Jump(uint64_t self);
void Player_Player_KnockBack(uint64_t self, uint64_t p1, double p2, double p3);
int64_t Player_Player_Latency(uint64_t self);
uint64_t Player_Player_Locale(uint64_t self);
int64_t Player_Player_MaxAirSupply(uint64_t self);
double Player_Player_MaxHealth(uint64_t self);
void Player_Player_Message(uint64_t self, uint64_t p1);
void Player_Player_Messagef(uint64_t self, char* p1, uint64_t p2);
void Player_Player_Messaget(uint64_t self, uint64_t p1, uint64_t p2);
void Player_Player_Move(uint64_t self, uint64_t p1, double p2, double p3);
void Player_Player_MoveItemsToInventory(uint64_t self);
char* Player_Player_Name(uint64_t self);
char* Player_Player_NameTag(uint64_t self);
int64_t Player_Player_OnFireDuration(uint64_t self);
bool Player_Player_OnGround(uint64_t self);
void Player_Player_OpenBlockContainer(uint64_t self, uint64_t p1, uint64_t p2);
void Player_Player_OpenSign(uint64_t self, uint64_t p1, bool p2);
void Player_Player_PickBlock(uint64_t self, uint64_t p1);
void Player_Player_PlaceBlock(uint64_t self, uint64_t p1, uint64_t p2, uint64_t p3);
void Player_Player_PlaySound(uint64_t self, uint64_t p1);
GoArray Player_Player_Position(uint64_t self);
void Player_Player_PunchAir(uint64_t self);
void Player_Player_ReleaseItem(uint64_t self);
void Player_Player_RemoveAllDebugShapes(uint64_t self);
void Player_Player_RemoveBossBar(uint64_t self);
void Player_Player_RemoveDebugShape(uint64_t self, uint64_t p1);
void Player_Player_RemoveEffect(uint64_t self, uint64_t p1);
void Player_Player_RemoveExperience(uint64_t self, int p1);
void Player_Player_RemoveScoreboard(uint64_t self);
void Player_Player_ResetEnchantmentSeed(uint64_t self);
void Player_Player_ResetFallDistance(uint64_t self);
uint64_t Player_Player_Respawn(uint64_t self);
GoArray Player_Player_Rotation(uint64_t self);
void Player_Player_Saturate(uint64_t self, int p1, double p2);
double Player_Player_Scale(uint64_t self);
char* Player_Player_ScoreTag(uint64_t self);
char* Player_Player_SelfSignedID(uint64_t self);
void Player_Player_SendBossBar(uint64_t self, uint64_t p1);
void Player_Player_SendCommandOutput(uint64_t self, uint64_t p1);
void Player_Player_SendDialogue(uint64_t self, uint64_t p1, uint64_t p2);
void Player_Player_SendForm(uint64_t self, uint64_t p1);
void Player_Player_SendJukeboxPopup(uint64_t self, uint64_t p1);
void Player_Player_SendPopup(uint64_t self, uint64_t p1);
void Player_Player_SendScoreboard(uint64_t self, uint64_t p1);
void Player_Player_SendSleepingIndicator(uint64_t self, int p1, int p2);
void Player_Player_SendTip(uint64_t self, uint64_t p1);
void Player_Player_SendTitle(uint64_t self, uint64_t p1);
void Player_Player_SendToast(uint64_t self, char* p1, char* p2);
void Player_Player_SetAbsorption(uint64_t self, double p1);
void Player_Player_SetAirSupply(uint64_t self, int64_t p1);
void Player_Player_SetCooldown(uint64_t self, uint64_t p1, int64_t p2);
void Player_Player_SetExperienceLevel(uint64_t self, int p1);
void Player_Player_SetExperienceProgress(uint64_t self, double p1);
void Player_Player_SetFlightSpeed(uint64_t self, double p1);
void Player_Player_SetFood(uint64_t self, int p1);
void Player_Player_SetGameMode(uint64_t self, uint64_t p1);
void Player_Player_SetHeldItems(uint64_t self, uint64_t p1, uint64_t p2);
uint64_t Player_Player_SetHeldSlot(uint64_t self, int p1);
void Player_Player_SetImmobile(uint64_t self);
void Player_Player_SetInvisible(uint64_t self);
void Player_Player_SetMaxAirSupply(uint64_t self, int64_t p1);
void Player_Player_SetMaxHealth(uint64_t self, double p1);
void Player_Player_SetMobile(uint64_t self);
void Player_Player_SetNameTag(uint64_t self, char* p1);
void Player_Player_SetOnFire(uint64_t self, int64_t p1);
void Player_Player_SetScale(uint64_t self, double p1);
void Player_Player_SetScoreTag(uint64_t self, uint64_t p1);
void Player_Player_SetSkin(uint64_t self, uint64_t p1);
void Player_Player_SetSpeed(uint64_t self, double p1);
void Player_Player_SetVelocity(uint64_t self, uint64_t p1);
void Player_Player_SetVerticalFlightSpeed(uint64_t self, double p1);
void Player_Player_SetVisible(uint64_t self);
void Player_Player_ShowCoordinates(uint64_t self);
void Player_Player_ShowEntity(uint64_t self, uint64_t p1);
void Player_Player_ShowHudElement(uint64_t self, uint64_t p1);
void Player_Player_ShowParticle(uint64_t self, uint64_t p1, uint64_t p2);
uint64_t Player_Player_Skin(uint64_t self);
void Player_Player_Sleep(uint64_t self, uint64_t p1);
Player_Player_Sleeping_Result Player_Player_Sleeping(uint64_t self);
bool Player_Player_Sneaking(uint64_t self);
double Player_Player_Speed(uint64_t self);
void Player_Player_Splash(uint64_t self, uint64_t p1, uint64_t p2);
bool Player_Player_Sprinting(uint64_t self);
void Player_Player_StartBreaking(uint64_t self, uint64_t p1, int p2);
void Player_Player_StartCrawling(uint64_t self);
void Player_Player_StartFlying(uint64_t self);
void Player_Player_StartGliding(uint64_t self);
void Player_Player_StartSneaking(uint64_t self);
void Player_Player_StartSprinting(uint64_t self);
void Player_Player_StartSwimming(uint64_t self);
void Player_Player_StopCrawling(uint64_t self);
void Player_Player_StopFlying(uint64_t self);
void Player_Player_StopGliding(uint64_t self);
void Player_Player_StopSneaking(uint64_t self);
void Player_Player_StopSprinting(uint64_t self);
void Player_Player_StopSwimming(uint64_t self);
bool Player_Player_Swimming(uint64_t self);
void Player_Player_SwingArm(uint64_t self);
void Player_Player_Teleport(uint64_t self, uint64_t p1);
void Player_Player_Tick(uint64_t self, uint64_t p1, int64_t p2);
double Player_Player_TorsoHeight(uint64_t self);
uint64_t Player_Player_Transfer(uint64_t self, char* p1);
uint64_t Player_Player_TurnLecternPage(uint64_t self, uint64_t p1, int p2);
uint64_t Player_Player_Tx(uint64_t self);
GoArray Player_Player_UUID(uint64_t self);
void Player_Player_UpdateDiagnostics(uint64_t self, uint64_t p1);
void Player_Player_UseItem(uint64_t self);
void Player_Player_UseItemOnBlock(uint64_t self, uint64_t p1, int p2, uint64_t p3);
bool Player_Player_UseItemOnEntity(uint64_t self, uint64_t p1);
bool Player_Player_UsingItem(uint64_t self);
GoArray Player_Player_Velocity(uint64_t self);
double Player_Player_VerticalFlightSpeed(uint64_t self);
GoArray Player_Player_VisibleDebugShapes(uint64_t self);
void Player_Player_Wake(uint64_t self);
char* Player_Player_XUID(uint64_t self);

uint64_t Scoreboard_Scoreboard_Ctor(void);
bool Scoreboard_Scoreboard_Descending(uint64_t self);
GoArray Scoreboard_Scoreboard_Lines(uint64_t self);
char* Scoreboard_Scoreboard_Name(uint64_t self);
void Scoreboard_Scoreboard_Remove(uint64_t self, int p1);
void Scoreboard_Scoreboard_RemovePadding(uint64_t self);
void Scoreboard_Scoreboard_Set(uint64_t self, int p1, char* p2);
void Scoreboard_Scoreboard_SetDescending(uint64_t self);
Scoreboard_Scoreboard_Write_Result Scoreboard_Scoreboard_Write(uint64_t self, uint64_t p1);
Scoreboard_Scoreboard_WriteString_Result Scoreboard_Scoreboard_WriteString(uint64_t self, char* p1);

uint64_t Session_Diagnostics_Ctor(void);

uint64_t Skin_Skin_Ctor(void);
uint64_t Skin_Skin_At(uint64_t self, int p1, int p2);
uint64_t Skin_Skin_Bounds(uint64_t self);
uint64_t Skin_Skin_ColorModel(uint64_t self);

uint64_t Title_Title_Ctor(void);
char* Title_Title_ActionText(uint64_t self);
int64_t Title_Title_Duration(uint64_t self);
int64_t Title_Title_FadeInDuration(uint64_t self);
int64_t Title_Title_FadeOutDuration(uint64_t self);
char* Title_Title_Subtitle(uint64_t self);
char* Title_Title_Text(uint64_t self);
uint64_t Title_Title_WithActionText(uint64_t self, uint64_t p1);
uint64_t Title_Title_WithDuration(uint64_t self, int64_t p1);
uint64_t Title_Title_WithFadeInDuration(uint64_t self, int64_t p1);
uint64_t Title_Title_WithFadeOutDuration(uint64_t self, int64_t p1);
uint64_t Title_Title_WithSubtitle(uint64_t self, uint64_t p1);

uint64_t World_EntityAnimation_Ctor(void);
char* World_EntityAnimation_Controller(uint64_t self);
char* World_EntityAnimation_Name(uint64_t self);
char* World_EntityAnimation_NextState(uint64_t self);
char* World_EntityAnimation_StopCondition(uint64_t self);
uint64_t World_EntityAnimation_WithController(uint64_t self, char* p1);
uint64_t World_EntityAnimation_WithNextState(uint64_t self, char* p1);
uint64_t World_EntityAnimation_WithStopCondition(uint64_t self, char* p1);

uint64_t World_EntityData_Ctor(void);

uint64_t World_EntityHandle_Ctor(void);
uint64_t World_EntityHandle_Close(uint64_t self);
World_EntityHandle_Entity_Result World_EntityHandle_Entity(uint64_t self, uint64_t p1);
bool World_EntityHandle_ExecWorld(uint64_t self, uint64_t p1);
uint64_t World_EntityHandle_Type(uint64_t self);
GoArray World_EntityHandle_UUID(uint64_t self);

uint64_t World_EntityRegistry_Ctor(void);
uint64_t World_EntityRegistry_Config(uint64_t self);
World_EntityRegistry_Lookup_Result World_EntityRegistry_Lookup(uint64_t self, char* p1);
GoArray World_EntityRegistry_Types(uint64_t self);

uint64_t World_EntityRegistryConfig_Ctor(void);
uint64_t World_EntityRegistryConfig_New(uint64_t self, uint64_t p1);

uint64_t World_SetOpts_Ctor(void);

uint64_t World_Tx_Ctor(void);
uint64_t World_Tx_AddEntity(uint64_t self, uint64_t p1);
void World_Tx_AddParticle(uint64_t self, uint64_t p1, uint64_t p2);
uint64_t World_Tx_Biome(uint64_t self, uint64_t p1);
uint64_t World_Tx_Block(uint64_t self, uint64_t p1);
void World_Tx_BroadcastSleepingIndicator(uint64_t self);
void World_Tx_BroadcastSleepingReminder(uint64_t self, uint64_t p1);
void World_Tx_BuildStructure(uint64_t self, uint64_t p1, uint64_t p2);
uint64_t World_Tx_Entities(uint64_t self);
uint64_t World_Tx_EntitiesWithin(uint64_t self, uint64_t p1);
int World_Tx_HighestBlock(uint64_t self, int p1, int p2);
int World_Tx_HighestLightBlocker(uint64_t self, int p1, int p2);
uint8_t World_Tx_Light(uint64_t self, uint64_t p1);
World_Tx_Liquid_Result World_Tx_Liquid(uint64_t self, uint64_t p1);
void World_Tx_PlayEntityAnimation(uint64_t self, uint64_t p1, uint64_t p2);
void World_Tx_PlaySound(uint64_t self, uint64_t p1, uint64_t p2);
uint64_t World_Tx_Players(uint64_t self);
bool World_Tx_Raining(uint64_t self);
bool World_Tx_RainingAt(uint64_t self, uint64_t p1);
GoArray World_Tx_Range(uint64_t self);
uint64_t World_Tx_RemoveEntity(uint64_t self, uint64_t p1);
void World_Tx_ScheduleBlockUpdate(uint64_t self, uint64_t p1, uint64_t p2, int64_t p3);
void World_Tx_SetBiome(uint64_t self, uint64_t p1, uint64_t p2);
void World_Tx_SetBlock(uint64_t self, uint64_t p1, uint64_t p2, uint64_t p3);
void World_Tx_SetLiquid(uint64_t self, uint64_t p1, uint64_t p2);
uint8_t World_Tx_SkyLight(uint64_t self, uint64_t p1);
uint64_t World_Tx_Sleepers(uint64_t self);
bool World_Tx_SnowingAt(uint64_t self, uint64_t p1);
double World_Tx_Temperature(uint64_t self, uint64_t p1);
bool World_Tx_Thundering(uint64_t self);
bool World_Tx_ThunderingAt(uint64_t self, uint64_t p1);
GoArray World_Tx_Viewers(uint64_t self, uint64_t p1);
uint64_t World_Tx_World(uint64_t self);

uint64_t World_World_Ctor(void);
uint64_t World_World_Close(uint64_t self);
uint64_t World_World_DefaultGameMode(uint64_t self);
uint64_t World_World_Difficulty(uint64_t self);
uint64_t World_World_Dimension(uint64_t self);
uint64_t World_World_EntityRegistry(uint64_t self);
uint64_t World_World_Exec(uint64_t self, uint64_t p1);
void World_World_Handle(uint64_t self, uint64_t p1);
uint64_t World_World_Handler(uint64_t self);
char* World_World_Name(uint64_t self);
GoArray World_World_PlayerSpawn(uint64_t self, uint64_t p1);
uint64_t World_World_PortalDestination(uint64_t self, uint64_t p1);
GoArray World_World_Range(uint64_t self);
void World_World_Save(uint64_t self);
void World_World_SetDefaultGameMode(uint64_t self, uint64_t p1);
void World_World_SetDifficulty(uint64_t self, uint64_t p1);
void World_World_SetPlayerSpawn(uint64_t self, uint64_t p1, uint64_t p2);
void World_World_SetRequiredSleepDuration(uint64_t self, int64_t p1);
void World_World_SetSpawn(uint64_t self, uint64_t p1);
void World_World_SetTickRange(uint64_t self, int p1);
void World_World_SetTime(uint64_t self, int p1);
GoArray World_World_Spawn(uint64_t self);
void World_World_StartRaining(uint64_t self, int64_t p1);
void World_World_StartThundering(uint64_t self, int64_t p1);
void World_World_StartTime(uint64_t self);
void World_World_StartWeatherCycle(uint64_t self);
void World_World_StopRaining(uint64_t self);
void World_World_StopThundering(uint64_t self);
void World_World_StopTime(uint64_t self);
void World_World_StopWeatherCycle(uint64_t self);
int World_World_Time(uint64_t self);
bool World_World_TimeCycle(uint64_t self);

#ifdef __cplusplus
}
#endif

#endif
