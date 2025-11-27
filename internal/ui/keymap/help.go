package keymap

import (
	"fmt"

	"github.com/vin/ck123gogo/internal/i18n"
)

// HelpText is kept for backward compatibility (English only).
const HelpText = `
[::b]Keyboard Shortcuts[::-]
 1-5     : Switch resource type (1=EC2, 2=RDS, 3=S3, 4=Lambda, 5=Route53)
 /       : Focus search bar
 Enter   : Select/enter bucket/zone
 Backspace: Go back to parent
 Esc     : Exit to main list
 p       : Select AWS Profile (Region auto-switches)
 a       : Show actions for selected resource
 t       : Toggle theme (dark/light/high-contrast)
 l       : Toggle language (English/中文)
 g       : Refresh current resource list
 ?       : Show this help
 q       : Quit application

[::b]Profile Picker[::-]
 j/k   : Move up/down
 Enter : Select profile
 Esc/q : Cancel
`

// GetHelpText returns the help text in the current language.
func GetHelpText() string {
	return fmt.Sprintf(`[::b]%s[::-]
 %s
 %s
 %s
 %s
 %s
 %s
 %s
 %s
 %s
 %s
 %s
 %s

[::b]%s[::-]
 %s
 %s
 %s
`,
		i18n.T("help.title"),
		i18n.T("help.resource_switch"),
		i18n.T("help.search"),
		i18n.T("help.enter"),
		i18n.T("help.backspace"),
		i18n.T("help.escape"),
		i18n.T("help.profile"),
		i18n.T("help.action"),
		i18n.T("help.theme"),
		i18n.T("help.language"),
		i18n.T("help.refresh"),
		i18n.T("help.help"),
		i18n.T("help.quit"),
		i18n.T("help.picker_title"),
		i18n.T("help.picker_nav"),
		i18n.T("help.picker_select"),
		i18n.T("help.picker_cancel"),
	)
}
