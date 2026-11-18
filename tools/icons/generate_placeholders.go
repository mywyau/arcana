package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type icon struct {
	id     string
	shape  string
	colour string
	accent string
}

var icons = []icon{
	{"oak_shortbow", "bow", "#a87842", "#d8bd78"},
	{"willow_shortbow", "bow", "#78a96a", "#c5d989"},
	{"maple_shortbow", "bow", "#c28245", "#e2a65d"},
	{"yew_shortbow", "bow", "#41784b", "#8fc476"},
	{"unstrung_shortbow", "unstrung_bow", "#8d6a47", "#c9a36b"},
	{"unstrung_oak_shortbow", "unstrung_bow", "#a87842", "#d8bd78"},
	{"unstrung_willow_shortbow", "unstrung_bow", "#78a96a", "#c5d989"},
	{"unstrung_maple_shortbow", "unstrung_bow", "#c28245", "#e2a65d"},
	{"unstrung_yew_shortbow", "unstrung_bow", "#41784b", "#8fc476"},
	{"wand", "wand", "#8d6a47", "#d8d6c4"},
	{"moonspun_wand", "wand", "#4967a8", "#a8d8ff"},
	{"enchanted_staff", "staff", "#6750a4", "#c7a8ff"},
	{"stardew_staff", "staff", "#253b78", "#ffe58c"},
	{"spear", "spear", "#8f969f", "#dce4eb"},
	{"halberd", "halberd", "#69727c", "#dce4eb"},
	{"crossbow", "crossbow", "#815a35", "#c8d0d8"},

	{"bronze_bolt", "bolt", "#a96832", "#e0aa68"},
	{"iron_bolt", "bolt", "#69727c", "#c8d0d8"},
	{"steel_bolt", "bolt", "#aab4bc", "#eef4f7"},
	{"mithril_bolt", "bolt", "#278eaa", "#86e4f2"},
	{"adamantite_bolt", "bolt", "#28784f", "#74d69a"},
	{"bronze_arrow_tips", "arrow_tips", "#a96832", "#e0aa68"},
	{"iron_arrow_tips", "arrow_tips", "#69727c", "#c8d0d8"},
	{"steel_arrow_tips", "arrow_tips", "#aab4bc", "#eef4f7"},
	{"mithril_arrow_tips", "arrow_tips", "#278eaa", "#86e4f2"},
	{"adamantite_arrow_tips", "arrow_tips", "#28784f", "#74d69a"},

	{"bronze_shield", "shield", "#9b5f31", "#d59a5d"},
	{"bronze_legs", "legs", "#9b5f31", "#d59a5d"},
	{"bronze_gloves", "hands", "#9b5f31", "#d59a5d"},
	{"bronze_boots", "feet", "#9b5f31", "#d59a5d"},
	{"iron_shield", "shield", "#626a73", "#aeb7bf"},
	{"iron_legs", "legs", "#626a73", "#aeb7bf"},
	{"iron_gloves", "hands", "#626a73", "#aeb7bf"},
	{"iron_boots", "feet", "#626a73", "#aeb7bf"},
	{"steel_shield", "shield", "#9da8b1", "#e2e8ec"},
	{"steel_legs", "legs", "#9da8b1", "#e2e8ec"},
	{"steel_gloves", "hands", "#9da8b1", "#e2e8ec"},
	{"steel_boots", "feet", "#9da8b1", "#e2e8ec"},
	{"mithril_shield", "shield", "#298aa8", "#80d8e8"},
	{"mithril_legs", "legs", "#298aa8", "#80d8e8"},
	{"mithril_gloves", "hands", "#298aa8", "#80d8e8"},
	{"mithril_boots", "feet", "#298aa8", "#80d8e8"},
	{"adamantite_shield", "shield", "#28784f", "#72ce91"},
	{"adamantite_legs", "legs", "#28784f", "#72ce91"},
	{"adamantite_gloves", "hands", "#28784f", "#72ce91"},
	{"adamantite_boots", "feet", "#28784f", "#72ce91"},

	{"novice_robe_hood", "hood", "#685f78", "#aaa0ba"},
	{"novice_robe_top", "robe", "#685f78", "#aaa0ba"},
	{"novice_robe_gloves", "hands", "#685f78", "#aaa0ba"},
	{"novice_robe_bottoms", "robe_legs", "#685f78", "#aaa0ba"},
	{"novice_robe_boots", "feet", "#685f78", "#aaa0ba"},
	{"mage_robe_hood", "hood", "#365a96", "#85b5ef"},
	{"mage_robe_top", "robe", "#365a96", "#85b5ef"},
	{"mage_robe_gloves", "hands", "#365a96", "#85b5ef"},
	{"mage_robe_bottoms", "robe_legs", "#365a96", "#85b5ef"},
	{"mage_robe_boots", "feet", "#365a96", "#85b5ef"},
	{"moonspun_hood", "hood", "#4b4f9d", "#b9c8ff"},
	{"moonspun_robe_top", "robe", "#4b4f9d", "#b9c8ff"},
	{"moonspun_gloves", "hands", "#4b4f9d", "#b9c8ff"},
	{"moonspun_bottoms", "robe_legs", "#4b4f9d", "#b9c8ff"},
	{"moonspun_boots", "feet", "#4b4f9d", "#b9c8ff"},
	{"enchanted_hood", "hood", "#7250a8", "#d6aaff"},
	{"enchanted_robe_top", "robe", "#7250a8", "#d6aaff"},
	{"enchanted_gloves", "hands", "#7250a8", "#d6aaff"},
	{"enchanted_bottoms", "robe_legs", "#7250a8", "#d6aaff"},
	{"enchanted_boots", "feet", "#7250a8", "#d6aaff"},
	{"stardew_hood", "hood", "#263765", "#ffe28a"},
	{"stardew_robe_top", "robe", "#263765", "#ffe28a"},
	{"stardew_gloves", "hands", "#263765", "#ffe28a"},
	{"stardew_bottoms", "robe_legs", "#263765", "#ffe28a"},
	{"stardew_boots", "feet", "#263765", "#ffe28a"},

	{"novice_ranger_hood", "ranger_hood", "#65734c", "#9eae72"},
	{"novice_ranger_body", "ranger_body", "#65734c", "#9eae72"},
	{"novice_ranger_gloves", "hands", "#65734c", "#9eae72"},
	{"novice_ranger_legs", "legs", "#65734c", "#9eae72"},
	{"novice_ranger_boots", "feet", "#65734c", "#9eae72"},
	{"scout_hood", "ranger_hood", "#557447", "#92bd72"},
	{"scout_body", "ranger_body", "#557447", "#92bd72"},
	{"scout_gloves", "hands", "#557447", "#92bd72"},
	{"scout_legs", "legs", "#557447", "#92bd72"},
	{"scout_boots", "feet", "#557447", "#92bd72"},
	{"tracker_hood", "ranger_hood", "#426f59", "#78b890"},
	{"tracker_body", "ranger_body", "#426f59", "#78b890"},
	{"tracker_gloves", "hands", "#426f59", "#78b890"},
	{"tracker_legs", "legs", "#426f59", "#78b890"},
	{"tracker_boots", "feet", "#426f59", "#78b890"},
	{"hunter_hood", "ranger_hood", "#6d563c", "#b58b58"},
	{"hunter_body", "ranger_body", "#6d563c", "#b58b58"},
	{"hunter_gloves", "hands", "#6d563c", "#b58b58"},
	{"hunter_legs", "legs", "#6d563c", "#b58b58"},
	{"hunter_boots", "feet", "#6d563c", "#b58b58"},
	{"marksman_hood", "ranger_hood", "#3f4e5c", "#8eb2c8"},
	{"marksman_body", "ranger_body", "#3f4e5c", "#8eb2c8"},
	{"marksman_gloves", "hands", "#3f4e5c", "#8eb2c8"},
	{"marksman_legs", "legs", "#3f4e5c", "#8eb2c8"},
	{"marksman_boots", "feet", "#3f4e5c", "#8eb2c8"},

	{"bones", "bones", "#ded8c5", "#fff9e6"},
	{"bowstring", "string", "#d8d2bd", "#ffffff"},
	{"sunleaf", "herb", "#72a946", "#f0cf55"},
	{"moonleaf", "herb", "#426fa5", "#b9d5ff"},
	{"ruby_amulet", "amulet", "#b72f3f", "#ff8993"},
	{"emerald_amulet", "amulet", "#268956", "#76e0a4"},
	{"sapphire_amulet", "amulet", "#2d67b1", "#84b9ff"},
	{"diamond_amulet", "amulet", "#a5d7e5", "#ffffff"},
	{"ruby_ring", "ring", "#b72f3f", "#ff8993"},
	{"emerald_ring", "ring", "#268956", "#76e0a4"},
	{"sapphire_ring", "ring", "#2d67b1", "#84b9ff"},
	{"diamond_ring", "ring", "#a5d7e5", "#ffffff"},
	{"tyrants_ring", "tyrant_ring", "#b72f3f", "#ffd36a"},
	{"ranged_tyrants_ring", "tyrant_ring", "#268956", "#ffd36a"},
	{"mage_tyrants_ring", "tyrant_ring", "#6550b8", "#ffd36a"},
	{"standard_spellbook", "book", "#526c9b", "#d7e4ff"},
	{"skin_bible", "book", "#42234f", "#bf79d4"},
	{"holy_book", "book", "#a98228", "#fff0a3"},
	{"clerics_tome", "book", "#657b8b", "#e8f6ff"},
}

const header = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><g stroke="#17212a" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">`

func main() {
	root := filepath.Join("client", "assets", "icons", "generated", "placeholders")
	if err := os.MkdirAll(root, 0o755); err != nil {
		panic(err)
	}
	for _, definition := range icons {
		body := shape(definition)
		contents := header + body + `</g></svg>`
		path := filepath.Join(root, definition.id+".svg")
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			panic(err)
		}
	}
	fmt.Printf("generated %d placeholder icons in %s\n", len(icons), root)
}

func shape(i icon) string {
	c, a := i.colour, i.accent
	switch i.shape {
	case "bow", "unstrung_bow":
		stringPath := ""
		if i.shape == "bow" {
			stringPath = `<path d="M46 10 L18 54" fill="none" stroke="` + a + `" stroke-width="2"/>`
		}
		return `<path d="M44 8 Q57 31 20 56 Q31 33 44 8Z" fill="` + c + `"/>` + stringPath
	case "wand":
		return `<path d="M18 52 L43 17" stroke="` + c + `" stroke-width="7"/><path d="M42 8 L48 14 L42 22 L36 14Z" fill="` + a + `"/>`
	case "staff":
		return `<path d="M21 55 L39 12" stroke="` + c + `" stroke-width="6"/><circle cx="41" cy="11" r="8" fill="` + a + `"/>`
	case "spear":
		return `<path d="M14 53 L47 16" stroke="` + c + `" stroke-width="5"/><path d="M43 8 L55 5 L51 18Z" fill="` + a + `"/>`
	case "halberd":
		return `<path d="M16 55 L43 10" stroke="` + c + `" stroke-width="5"/><path d="M39 8 Q55 8 55 24 L43 19Z" fill="` + a + `"/>`
	case "crossbow":
		return `<path d="M31 24 L31 55" stroke="` + c + `" stroke-width="7"/><path d="M10 16 Q32 38 54 16" fill="none" stroke="` + c + `" stroke-width="6"/><path d="M10 16 L54 16" stroke="` + a + `" stroke-width="2"/>`
	case "bolt":
		return `<path d="M13 49 L48 14" stroke="` + c + `" stroke-width="6"/><path d="M44 9 L56 7 L53 19Z" fill="` + a + `"/><path d="M10 43 L20 53" stroke="` + a + `" stroke-width="5"/>`
	case "arrow_tips":
		return `<path d="M15 16 L29 28 L12 31Z M36 11 L52 22 L35 27Z M26 36 L43 48 L24 53Z" fill="` + c + `"/><path d="M15 16 L29 28 M36 11 L52 22 M26 36 L43 48" stroke="` + a + `"/>`
	case "shield":
		return `<path d="M12 12 Q32 4 52 12 V31 Q48 49 32 58 Q16 49 12 31Z" fill="` + c + `"/><path d="M32 12 V49 M20 26 H44" stroke="` + a + `"/>`
	case "hood", "ranger_hood":
		return `<path d="M13 47 Q10 15 32 8 Q54 15 51 47 L42 56 H22Z" fill="` + c + `"/><path d="M22 42 Q19 22 32 18 Q45 22 42 42Z" fill="#17212a" stroke="` + a + `"/>`
	case "robe", "ranger_body":
		return `<path d="M20 9 H44 L55 55 H9Z" fill="` + c + `"/><path d="M24 10 Q32 24 40 10 M32 24 V51" fill="none" stroke="` + a + `"/>`
	case "hands":
		return `<path d="M12 27 L23 14 L31 23 L25 49 L15 54Z M52 27 L41 14 L33 23 L39 49 L49 54Z" fill="` + c + `"/><path d="M18 31 L26 35 M46 31 L38 35" stroke="` + a + `"/>`
	case "legs", "robe_legs":
		return `<path d="M17 9 H47 L43 55 H31 L28 30 L25 55 H13Z" fill="` + c + `"/><path d="M19 19 H45" stroke="` + a + `"/>`
	case "feet":
		return `<path d="M13 20 H28 V45 L23 54 H7 L9 42Z M36 20 H51 L55 42 L57 54 H41 L36 45Z" fill="` + c + `"/><path d="M10 43 H27 M37 43 H54" stroke="` + a + `"/>`
	case "bones":
		return `<path d="M14 17 L50 47 M50 17 L14 47" stroke="` + c + `" stroke-width="8"/><circle cx="13" cy="16" r="5" fill="` + a + `"/><circle cx="51" cy="48" r="5" fill="` + a + `"/><circle cx="51" cy="16" r="5" fill="` + a + `"/><circle cx="13" cy="48" r="5" fill="` + a + `"/>`
	case "string":
		return `<circle cx="29" cy="34" r="19" fill="none" stroke="` + c + `" stroke-width="6"/><path d="M24 10 Q51 18 39 49 Q17 59 11 38" fill="none" stroke="` + a + `" stroke-width="3"/>`
	case "herb":
		return `<path d="M32 55 Q27 35 32 10" fill="none" stroke="` + c + `" stroke-width="5"/><path d="M30 35 Q10 33 10 17 Q27 17 32 30 Q37 13 54 15 Q53 31 34 35 Q43 34 50 45 Q37 51 31 41Z" fill="` + c + `"/><circle cx="45" cy="19" r="4" fill="` + a + `" stroke="none"/>`
	case "amulet":
		return `<path d="M14 9 Q32 48 50 9" fill="none" stroke="#d6b85f" stroke-width="4"/><path d="M32 30 L43 42 L32 55 L21 42Z" fill="` + c + `"/><circle cx="32" cy="42" r="4" fill="` + a + `" stroke="none"/>`
	case "ring", "tyrant_ring":
		crown := ""
		if i.shape == "tyrant_ring" {
			crown = `<path d="M19 24 L23 10 L32 20 L41 10 L45 24Z" fill="` + a + `"/>`
		}
		return crown + `<circle cx="32" cy="38" r="17" fill="none" stroke="#d6b85f" stroke-width="7"/><path d="M24 23 L32 13 L40 23 L32 31Z" fill="` + c + `"/><circle cx="32" cy="21" r="3" fill="` + a + `" stroke="none"/>`
	case "book":
		return `<path d="M11 10 H47 Q54 10 54 18 V54 H18 Q11 54 11 47Z" fill="` + c + `"/><path d="M18 10 V54 M27 25 H45 M27 35 H43" stroke="` + a + `"/>`
	default:
		return `<rect x="10" y="10" width="44" height="44" rx="10" fill="` + c + `"/><circle cx="32" cy="32" r="11" fill="` + a + `"/>`
	}
}
