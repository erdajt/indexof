// SPDX-FileCopyrightText: 2026 erdajt <erdajtsopjani.tech@gmail.com>
// SPDX-License-Identifier: GPL-3.0-or-later

package theme

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name    string
	Accent  lipgloss.Color
	Subtle  lipgloss.Color
	Text    lipgloss.Color
	Dir     lipgloss.Color
	File    lipgloss.Color
	SelBg   lipgloss.Color
	SelFg   lipgloss.Color
	Error   lipgloss.Color
	Ok      lipgloss.Color
	Warning lipgloss.Color
	Border  lipgloss.Color
}

var all = map[string]Theme{
	"default": {
		Name: "default", Accent: "99", Subtle: "238", Text: "252",
		Dir: "117", File: "251", SelBg: "236", SelFg: "229",
		Error: "196", Ok: "114", Warning: "214", Border: "238",
	},
	"tokyo": {
		Name: "tokyo", Accent: "111", Subtle: "102", Text: "189",
		Dir: "117", File: "189", SelBg: "237", SelFg: "159",
		Error: "203", Ok: "115", Warning: "221", Border: "102",
	},
	"catppuccin": {
		Name: "catppuccin", Accent: "183", Subtle: "103", Text: "224",
		Dir: "117", File: "224", SelBg: "237", SelFg: "225",
		Error: "211", Ok: "150", Warning: "223", Border: "103",
	},
	"nord": {
		Name: "nord", Accent: "109", Subtle: "60", Text: "188",
		Dir: "110", File: "188", SelBg: "59", SelFg: "189",
		Error: "167", Ok: "108", Warning: "179", Border: "60",
	},
	"gruvbox": {
		Name: "gruvbox", Accent: "172", Subtle: "243", Text: "223",
		Dir: "108", File: "223", SelBg: "237", SelFg: "214",
		Error: "167", Ok: "142", Warning: "214", Border: "243",
	},
	"dracula": {
		Name: "dracula", Accent: "141", Subtle: "61", Text: "228",
		Dir: "117", File: "228", SelBg: "236", SelFg: "159",
		Error: "203", Ok: "84", Warning: "215", Border: "61",
	},
}

func Get(name string) Theme {
	if t, ok := all[name]; ok {
		return t
	}
	return all["default"]
}

func Names() []string {
	return []string{"default", "tokyo", "catppuccin", "nord", "gruvbox", "dracula"}
}
