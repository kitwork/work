// KitWork - Work Engine Core
// Copyright (C) 2025 Huỳnh Nhân Quốc

// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License version 3 (AGPL-3.0).

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the AGPL-3.0 License along with this program.
// If not, see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import "fmt"

// Type represents an work type in KitWork workflow engine
type Type string

// --- Work Types ---
const (
	// --- Network / Fetch ---
	TypeFetch   Type = "fetch"
	TypeHTTP    Type = "http"
	TypeClient  Type = "client"
	TypeRequest Type = "request"

	// --- Script / Command ---
	TypeScript  Type = "script"
	TypeCmd     Type = "cmd"
	TypeCommand Type = "command"

	// --- Flow Control ---
	TypeRouter Type = "router"

	// --- Flow Control ---
	TypeForeach Type = "foreach"

	TypeSwitch   Type = "switch"
	TypeLoop     Type = "loop"
	TypeReturn   Type = "return"
	TypeWait     Type = "wait"
	TypeRoutines Type = "routines"
	// TypeIf      Type = "if"

	// --- Logging / Debug ---
	TypeLog Type = "log"

	// --- Cron / Schedule ---
	TypeCron     Type = "cron"
	TypeSchedule Type = "schedule"

	// --- IO / Mail / Storage ---
	TypeSendMail Type = "sendmail"
	TypeSave     Type = "save"
	TypeCheck    Type = "check"

	// --- Browser Automation ---
	TypeChrome   Type = "chrome"
	TypeChromedp Type = "chromedp"

	// --- Custom / Fallback ---
	TypeCustom Type = "custom"
	TypeUnknow Type = "unknow"

	// --- Parse /  ---
	TypeParser  Type = "parse"
	TypeMapping Type = "mapping"
)

// TypeParse converts a string to a Type enum
// Returns error if the string is not a valid work type
func TypeParse(s string) (Type, error) {
	switch s {
	case "fetch":
		return TypeFetch, nil
	case "router":
		return TypeRouter, nil
	case "http":
		return TypeHTTP, nil
	case "client":
		return TypeHTTP, nil
	case "script":
		return TypeScript, nil
	case "cmd":
		return TypeCmd, nil
	case "command":
		return TypeCommand, nil
	case "foreach":
		return TypeForeach, nil
	case "routines":
		return TypeRoutines, nil
	// case "if":
	// 	return TypeIf, nil
	case "switch":
		return TypeSwitch, nil
	case "cron":
		return TypeCron, nil
	case "loop":
		return TypeLoop, nil
	case "return":
		return TypeReturn, nil
	case "wait":
		return TypeWait, nil
	case "log":
		return TypeLog, nil
	case "sendmail":
		return TypeSendMail, nil
	case "save":
		return TypeSave, nil
	case "check":
		return TypeCheck, nil
	case "chrome":
		return TypeChrome, nil
	case "chromedp":
		return TypeChromedp, nil
	case "custom":
		return TypeCustom, nil

	case "parse":
		return TypeParser, nil

	case "mapping":
		return TypeMapping, nil
	default:
		return "", fmt.Errorf("invalid work type: %s", s)
	}
}

// TypeParseSafe converts string to Type enum safely
// Logs a warning and returns TypeCustom if invalid
func TypeParseSafe(s string) Type {
	t, err := TypeParse(s)
	if err != nil {
		fmt.Printf("⚠️ Warning: %v, fallback to TypeCustom\n", err)
		return TypeCustom
	}
	return t
}
