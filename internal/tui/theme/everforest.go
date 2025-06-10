package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// EverforestTheme implements the Theme interface with Everforest colors.
// It provides both dark and light variants.
type EverforestTheme struct {
	BaseTheme
}

// NewEverforestTheme creates a new instance of the Everforest theme.
func NewEverforestTheme() *EverforestTheme {
	// Everforest color palette - Medium Dark variant
	// Official colors from https://github.com/sainnhe/everforest/wiki
	darkBackground := "#2d353b"
	darkCurrentLine := "#343f44"
	darkSelection := "#3d484d"
	darkForeground := "#d3c6aa"
	darkComment := "#859289"
	darkRed := "#e67e80"
	darkOrange := "#e69875"
	darkYellow := "#dbbc7f"
	darkGreen := "#a7c080"
	darkCyan := "#83c092"
	darkBlue := "#7fbbb3"
	darkPurple := "#d699b6"
	darkBorder := "#475258"
	darkGray := "#7a8478"

	// Light mode colors - Medium Light variant
	lightBackground := "#fdf6e3"
	lightCurrentLine := "#f4f0d9"
	lightSelection := "#efebd4"
	lightForeground := "#5c6a72"
	lightComment := "#939f91"
	lightRed := "#f85552"
	lightOrange := "#f57d26"
	lightYellow := "#dfa000"
	lightGreen := "#8da101"
	lightCyan := "#35a77c"
	lightBlue := "#3a94c5"
	lightPurple := "#df69ba"
	lightBorder := "#e6e2cc"
	lightGray := "#a6b0a0"

	theme := &EverforestTheme{}

	// Base colors
	theme.PrimaryColor = lipgloss.AdaptiveColor{
		Dark:  darkBlue,
		Light: lightBlue,
	}
	theme.SecondaryColor = lipgloss.AdaptiveColor{
		Dark:  darkPurple,
		Light: lightPurple,
	}
	theme.AccentColor = lipgloss.AdaptiveColor{
		Dark:  darkOrange,
		Light: lightOrange,
	}

	// Status colors
	theme.ErrorColor = lipgloss.AdaptiveColor{
		Dark:  darkRed,
		Light: lightRed,
	}
	theme.WarningColor = lipgloss.AdaptiveColor{
		Dark:  darkOrange,
		Light: lightOrange,
	}
	theme.SuccessColor = lipgloss.AdaptiveColor{
		Dark:  darkGreen,
		Light: lightGreen,
	}
	theme.InfoColor = lipgloss.AdaptiveColor{
		Dark:  darkBlue,
		Light: lightBlue,
	}

	// Text colors
	theme.TextColor = lipgloss.AdaptiveColor{
		Dark:  darkForeground,
		Light: lightForeground,
	}
	theme.TextMutedColor = lipgloss.AdaptiveColor{
		Dark:  darkComment,
		Light: lightComment,
	}
	theme.TextEmphasizedColor = lipgloss.AdaptiveColor{
		Dark:  darkYellow,
		Light: lightYellow,
	}

	// Background colors
	theme.BackgroundColor = lipgloss.AdaptiveColor{
		Dark:  darkBackground,
		Light: lightBackground,
	}
	theme.BackgroundSecondaryColor = lipgloss.AdaptiveColor{
		Dark:  darkCurrentLine,
		Light: lightCurrentLine,
	}
	theme.BackgroundDarkerColor = lipgloss.AdaptiveColor{
		Dark:  "#232a2e", // Background Dim from Medium Dark
		Light: "#efebd4", // Background Dim from Medium Light
	}

	// Border colors
	theme.BorderNormalColor = lipgloss.AdaptiveColor{
		Dark:  darkBorder,
		Light: lightBorder,
	}
	theme.BorderFocusedColor = lipgloss.AdaptiveColor{
		Dark:  darkBlue,
		Light: lightBlue,
	}
	theme.BorderDimColor = lipgloss.AdaptiveColor{
		Dark:  darkSelection,
		Light: lightSelection,
	}

	// Diff view colors
	theme.DiffAddedColor = lipgloss.AdaptiveColor{
		Dark:  "#425047", // Background Green from Medium Dark
		Light: "#f0f1d2", // Background Green from Medium Light
	}
	theme.DiffRemovedColor = lipgloss.AdaptiveColor{
		Dark:  "#543a48", // Background Red from Medium Dark
		Light: "#fbe3da", // Background Red from Medium Light
	}
	theme.DiffContextColor = lipgloss.AdaptiveColor{
		Dark:  darkGray,
		Light: lightGray,
	}
	theme.DiffHunkHeaderColor = lipgloss.AdaptiveColor{
		Dark:  darkGray,
		Light: lightGray,
	}
	theme.DiffHighlightAddedColor = lipgloss.AdaptiveColor{
		Dark:  darkGreen,
		Light: lightGreen,
	}
	theme.DiffHighlightRemovedColor = lipgloss.AdaptiveColor{
		Dark:  darkRed,
		Light: lightRed,
	}
	theme.DiffAddedBgColor = lipgloss.AdaptiveColor{
		Dark:  "#425047", // Background Green from Medium Dark
		Light: "#f0f1d2", // Background Green from Medium Light
	}
	theme.DiffRemovedBgColor = lipgloss.AdaptiveColor{
		Dark:  "#543a48", // Background Red from Medium Dark
		Light: "#fbe3da", // Background Red from Medium Light
	}
	theme.DiffContextBgColor = lipgloss.AdaptiveColor{
		Dark:  darkBackground,
		Light: lightBackground,
	}
	theme.DiffLineNumberColor = lipgloss.AdaptiveColor{
		Dark:  darkGray,
		Light: lightGray,
	}
	theme.DiffAddedLineNumberBgColor = lipgloss.AdaptiveColor{
		Dark:  "#3c4841", // Background Green from Hard Dark
		Light: "#e5e6c5", // Background Green from Soft Light
	}
	theme.DiffRemovedLineNumberBgColor = lipgloss.AdaptiveColor{
		Dark:  "#4c3743", // Background Red from Hard Dark
		Light: "#f4dbd0", // Background Red from Soft Light
	}

	// Markdown colors
	theme.MarkdownTextColor = lipgloss.AdaptiveColor{
		Dark:  darkForeground,
		Light: lightForeground,
	}
	theme.MarkdownHeadingColor = lipgloss.AdaptiveColor{
		Dark:  darkPurple,
		Light: lightPurple,
	}
	theme.MarkdownLinkColor = lipgloss.AdaptiveColor{
		Dark:  darkBlue,
		Light: lightBlue,
	}
	theme.MarkdownLinkTextColor = lipgloss.AdaptiveColor{
		Dark:  darkCyan,
		Light: lightCyan,
	}
	theme.MarkdownCodeColor = lipgloss.AdaptiveColor{
		Dark:  darkGreen,
		Light: lightGreen,
	}
	theme.MarkdownBlockQuoteColor = lipgloss.AdaptiveColor{
		Dark:  darkYellow,
		Light: lightYellow,
	}
	theme.MarkdownEmphColor = lipgloss.AdaptiveColor{
		Dark:  darkYellow,
		Light: lightYellow,
	}
	theme.MarkdownStrongColor = lipgloss.AdaptiveColor{
		Dark:  darkOrange,
		Light: lightOrange,
	}
	theme.MarkdownHorizontalRuleColor = lipgloss.AdaptiveColor{
		Dark:  darkComment,
		Light: lightComment,
	}
	theme.MarkdownListItemColor = lipgloss.AdaptiveColor{
		Dark:  darkBlue,
		Light: lightBlue,
	}
	theme.MarkdownListEnumerationColor = lipgloss.AdaptiveColor{
		Dark:  darkCyan,
		Light: lightCyan,
	}
	theme.MarkdownImageColor = lipgloss.AdaptiveColor{
		Dark:  darkBlue,
		Light: lightBlue,
	}
	theme.MarkdownImageTextColor = lipgloss.AdaptiveColor{
		Dark:  darkCyan,
		Light: lightCyan,
	}
	theme.MarkdownCodeBlockColor = lipgloss.AdaptiveColor{
		Dark:  darkForeground,
		Light: lightForeground,
	}

	// Syntax highlighting colors
	theme.SyntaxCommentColor = lipgloss.AdaptiveColor{
		Dark:  darkComment,
		Light: lightComment,
	}
	theme.SyntaxKeywordColor = lipgloss.AdaptiveColor{
		Dark:  darkPurple,
		Light: lightPurple,
	}
	theme.SyntaxFunctionColor = lipgloss.AdaptiveColor{
		Dark:  darkBlue,
		Light: lightBlue,
	}
	theme.SyntaxVariableColor = lipgloss.AdaptiveColor{
		Dark:  darkRed,
		Light: lightRed,
	}
	theme.SyntaxStringColor = lipgloss.AdaptiveColor{
		Dark:  darkGreen,
		Light: lightGreen,
	}
	theme.SyntaxNumberColor = lipgloss.AdaptiveColor{
		Dark:  darkOrange,
		Light: lightOrange,
	}
	theme.SyntaxTypeColor = lipgloss.AdaptiveColor{
		Dark:  darkYellow,
		Light: lightYellow,
	}
	theme.SyntaxOperatorColor = lipgloss.AdaptiveColor{
		Dark:  darkCyan,
		Light: lightCyan,
	}
	theme.SyntaxPunctuationColor = lipgloss.AdaptiveColor{
		Dark:  darkForeground,
		Light: lightForeground,
	}

	return theme
}

func init() {
	// Register the Everforest theme with the theme manager
	RegisterTheme("everforest", NewEverforestTheme())
}
