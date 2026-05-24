package preset

// EnsureBuiltins writes default presets to the presets directory on first run.
// Existing presets are not overwritten.
func EnsureBuiltins() {
	for _, p := range builtins {
		if _, err := Load(p.Name); err != nil {
			_ = Save(&p)
		}
	}
}

var builtins = []Preset{
	{
		Name:      "destiny-raid",
		AutoNotes: true,
		Tags: []Tag{
			{Key: "1", Label: "boss encounter", Color: "#e8919e"},
			{Key: "2", Label: "traversal", Color: "#a8c8e8"},
			{Key: "3", Label: "adds clear", Color: "#8cc4a0"},
			{Key: "4", Label: "wipe", Color: "#f0a68c"},
			{Key: "5", Label: "loot / chest", Color: "#dbb87c"},
			{Key: "6", Label: "menu / loadout", Color: "#b8a0d4"},
			{Key: "7", Label: "cutscene", Color: "#c47a66"},
		},
	},
	{
		Name:      "default",
		AutoNotes: false,
		Tags: []Tag{
			{Key: "1", Label: "highlight", Color: "#e8919e"},
			{Key: "2", Label: "action", Color: "#a8c8e8"},
			{Key: "3", Label: "review", Color: "#8cc4a0"},
			{Key: "4", Label: "skip", Color: "#b8a0d4"},
		},
	},
}
