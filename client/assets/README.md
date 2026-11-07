# Client assets

Keep runtime assets grouped by purpose:

- `audio/music/` contains background music.
- `icons/equipment/` contains equipment and equipment-slot icons.
- `icons/items/` contains general inventory-item icons.
- `icons/runes/` contains rune item icons.
- `icons/navigation/`, `icons/prayers/`, `icons/skills/`, and `icons/spells/` contain HUD icons.

Use lowercase `snake_case` filenames that describe the in-game concept, such as
`air_rune.png` or `wanderers_glade.ogg`. Do not include export suffixes, sequence
numbers, spaces, or asset-generation prompts in filenames.

Reference item and HUD icons through `scripts/ui/icon_catalog.gd` where possible.
Godot-generated `.import` files should remain beside their source assets.
