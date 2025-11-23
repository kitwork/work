package work

import (
	"fmt"
	"time"
)

// Parse converts YAML-decoded map into an Work struct (recursive).
func Parsing(data map[string]interface{}) *Work {
	work := &Work{
		Config:  make(map[string]interface{}),
		Works:   []*Work{},
		Success: []*Work{},
		Error:   []*Work{},
		Kind:    KindUnknown,
	}

	for key, val := range data {

		switch v := val.(type) {

		// ================================
		// 1) FULL FORM
		//    log:
		//      name: hello
		//      works: [...]
		// ================================
		case map[string]interface{}:
			work.Type = TypeParseSafe(key)
			work.Kind = KindFull

			for k, vv := range v {
				switch k {
				case "name":
					if s, ok := vv.(string); ok {
						work.Name = s
					}

				case "works", "work":
					work.Works = parseList(vv)
				case "success":
					work.Success = parseList(vv)
				case "error":
					work.Error = parseList(vv)
				case "timeout":
					if tStr, ok := vv.(string); ok {
						d, err := time.ParseDuration(tStr)
						if err == nil {
							work.Timeout = d
						} else {
							fmt.Printf("⚠️  Warning: cannot parse timeout %v: %v\n", vv, err)
						}
					}
				default:
					work.Config[k] = vv
				}
			}

		// ================================
		// 2) LIST FORM
		//    works:
		//      - log
		//      - fetch: {...}
		// ================================
		case []interface{}:
			work.Type = TypeParseSafe(key)
			work.Kind = KindList
			work.Works = parseList(v)

		// ================================
		// 3) VALUE FORM
		//    log: "hello"
		// ================================
		case string, int, float64, bool:
			work.Type = TypeParseSafe(key)
			work.Kind = KindValue
			work.Config["value"] = v

		// ================================
		// 4) FALLBACK / UNKNOWN
		// ================================
		default:
			work.Type = TypeCustom
			work.Kind = KindUnknown
			work.Config["value"] = v
		}
	}

	return work
}

// parseList parses:
// - short string: "log"
// - full form: { fetch: {...} }
// - unknown primitives
func parseList(v interface{}) []*Work {
	result := []*Work{}

	list, ok := v.([]interface{})
	if !ok {
		return result
	}

	for _, item := range list {

		switch x := item.(type) {

		// -----------------------
		// SHORT FORM
		// - log
		// -----------------------
		case string:
			result = append(result, &Work{
				Type:   TypeParseSafe(x),
				Kind:   KindShort,
				Config: map[string]interface{}{},
			})

		// -----------------------
		// FULL FORM
		// - log: { msg: "hi" }
		// -----------------------
		case map[string]interface{}:
			result = append(result, Parsing(x))

		// -----------------------
		// Unknown value in list
		// -----------------------
		default:
			result = append(result, &Work{
				Type:   TypeCustom,
				Kind:   KindUnknown,
				Config: map[string]interface{}{"value": x},
			})
		}
	}

	return result
}
