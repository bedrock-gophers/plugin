using System.Runtime.InteropServices;

namespace BedrockPlugin.Sdk.Guest;

[StructLayout(LayoutKind.Sequential)]
public unsafe struct NativeHostApi
{
    public nuint ctx;

    public delegate* unmanaged<nuint, byte*, byte*, byte*, byte*, uint, byte*, uint, int> register_command;
    public delegate* unmanaged<nuint, uint, byte*, uint*, byte*> manage_plugins;
    public delegate* unmanaged<nuint, byte*, ulong> resolve_player_by_name;
    public delegate* unmanaged<nuint, ulong, ulong> player_handle;
    public delegate* unmanaged<nuint, uint*, byte*> online_player_names;
    public delegate* unmanaged<nuint, byte*, byte*, void> console_message;
    public delegate* unmanaged<nuint, uint, byte*, uint, uint*, byte*> host_call;
    public delegate* unmanaged<nuint, ulong, int> event_cancel;
    public delegate* unmanaged<nuint, ulong, byte*, uint, int> event_item_drop_set;

    public delegate* unmanaged<nuint, ulong, double> player_health;
    public delegate* unmanaged<nuint, ulong, double, int> set_player_health;
    public delegate* unmanaged<nuint, ulong, int> player_food;
    public delegate* unmanaged<nuint, ulong, int, int> set_player_food;
    public delegate* unmanaged<nuint, ulong, byte*> player_name;
    public delegate* unmanaged<nuint, ulong, int> player_game_mode;
    public delegate* unmanaged<nuint, ulong, int, int> set_player_game_mode;
    public delegate* unmanaged<nuint, ulong, byte*> player_xuid;
    public delegate* unmanaged<nuint, ulong, byte*> player_device_id;
    public delegate* unmanaged<nuint, ulong, byte*> player_device_model;
    public delegate* unmanaged<nuint, ulong, byte*> player_self_signed_id;
    public delegate* unmanaged<nuint, ulong, byte*> player_name_tag;
    public delegate* unmanaged<nuint, ulong, byte*, int> set_player_name_tag;
    public delegate* unmanaged<nuint, ulong, byte*> player_score_tag;
    public delegate* unmanaged<nuint, ulong, byte*, int> set_player_score_tag;
    public delegate* unmanaged<nuint, ulong, double> player_absorption;
    public delegate* unmanaged<nuint, ulong, double, int> set_player_absorption;
    public delegate* unmanaged<nuint, ulong, double> player_max_health;
    public delegate* unmanaged<nuint, ulong, double, int> set_player_max_health;
    public delegate* unmanaged<nuint, ulong, double> player_speed;
    public delegate* unmanaged<nuint, ulong, double, int> set_player_speed;
    public delegate* unmanaged<nuint, ulong, double> player_flight_speed;
    public delegate* unmanaged<nuint, ulong, double, int> set_player_flight_speed;
    public delegate* unmanaged<nuint, ulong, double> player_vertical_flight_speed;
    public delegate* unmanaged<nuint, ulong, double, int> set_player_vertical_flight_speed;
    public delegate* unmanaged<nuint, ulong, int> player_experience;
    public delegate* unmanaged<nuint, ulong, int> player_experience_level;
    public delegate* unmanaged<nuint, ulong, int, int> set_player_experience_level;
    public delegate* unmanaged<nuint, ulong, double> player_experience_progress;
    public delegate* unmanaged<nuint, ulong, double, int> set_player_experience_progress;
    public delegate* unmanaged<nuint, ulong, int> player_on_ground;
    public delegate* unmanaged<nuint, ulong, int> player_sneaking;
    public delegate* unmanaged<nuint, ulong, int, int> set_player_sneaking;
    public delegate* unmanaged<nuint, ulong, int> player_sprinting;
    public delegate* unmanaged<nuint, ulong, int, int> set_player_sprinting;
    public delegate* unmanaged<nuint, ulong, int> player_swimming;
    public delegate* unmanaged<nuint, ulong, int, int> set_player_swimming;
    public delegate* unmanaged<nuint, ulong, int> player_flying;
    public delegate* unmanaged<nuint, ulong, int, int> set_player_flying;
    public delegate* unmanaged<nuint, ulong, int> player_gliding;
    public delegate* unmanaged<nuint, ulong, int, int> set_player_gliding;
    public delegate* unmanaged<nuint, ulong, int> player_crawling;
    public delegate* unmanaged<nuint, ulong, int, int> set_player_crawling;
    public delegate* unmanaged<nuint, ulong, int> player_using_item;
    public delegate* unmanaged<nuint, ulong, int> player_invisible;
    public delegate* unmanaged<nuint, ulong, int, int> set_player_invisible;
    public delegate* unmanaged<nuint, ulong, int> player_immobile;
    public delegate* unmanaged<nuint, ulong, int, int> set_player_immobile;
    public delegate* unmanaged<nuint, ulong, int> player_dead;
    public delegate* unmanaged<nuint, ulong, long> player_latency;
    public delegate* unmanaged<nuint, ulong, long, int> set_player_on_fire_millis;
    public delegate* unmanaged<nuint, ulong, int, int> add_player_food;
    public delegate* unmanaged<nuint, ulong, int> player_use_item;
    public delegate* unmanaged<nuint, ulong, int> player_jump;
    public delegate* unmanaged<nuint, ulong, int> player_swing_arm;
    public delegate* unmanaged<nuint, ulong, int> player_wake;
    public delegate* unmanaged<nuint, ulong, int> player_extinguish;
    public delegate* unmanaged<nuint, ulong, int, int> set_player_show_coordinates;
    public delegate* unmanaged<nuint, ulong, byte*, void> player_message;

    public delegate* unmanaged<byte*, void> free_string;
    public delegate* unmanaged<byte*, void> free_bytes;
}
