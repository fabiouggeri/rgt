package run

import "rgt-server/option"

type RunAppConfig interface {
	ShowConsole() option.TypedOption[bool]
}
