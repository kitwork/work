// Work Engine Core
// Copyright (C) 2025 KitWork
// Author: Huỳnh Nhân Quốc | GitHub: https://github.com/huynhnhanquoc
// Licensed under GNU AGPL v3.0 — see <https://www.gnu.org/licenses/agpl-3.0.html>

package work

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/device"
)

type Screenshot struct {
	Name string `work:"name"`
	As   string `work:"as"`

	Navigate string `work:"navigate"`
	Quality  int    `work:"quality"`
	Element  string `work:"element"`

	Timeout time.Duration `work:"timeout"`
	Delay   time.Duration `work:"delay"`

	Device        chromedp.Device `work:"device"`
	Mode          string          `work:"mode"`
	Ads           bool            `work:"ads"`
	Trackers      bool            `work:"tracker"`
	CookieBanners bool            `work:"banners"`

	Height int `work:"height"`
	Width  int `work:"width"`
}

func (t *Work) Screenshot(ctx *Context) error {
	cfg := Screenshot{Mode: "light", Device: device.IPadProlandscape}

	switch t.Kind {
	case KindValue:
		cfg.Navigate = t.value()

	case KindFull:
		for k, v := range t.Config { // bỏ len(t.Config) == 1

			switch k {
			case "as":
				cfg.As = ToString(v)
			case "navigate", "url":
				val, err := ctx.evaluate(v)
				if err != nil {
					return err
				}

				cfg.Navigate = NormalizeURL(val)
			case "quality":
				cfg.Quality = ToInt(v)
			case "element":
				cfg.Element = ToString(v)
			case "device":

				val, err := ctx.evaluate(v)
				if err != nil {
					return err
				}
				name := ToString(val)
				switch name {
				case "mobile":
					cfg.Device = device.IPhone13ProMax
				case "tablet":
					cfg.Device = device.IPadgen6
				case "laptop":
					cfg.Device = device.IPadPro11landscape
				case "desktop":
					cfg.Device = device.IPadProlandscape
				case "full":
					cfg.Device = nil
				default:
					cfg.Device = parseDevice(name)
				}

			case "dark_mode":
				if ToBool(v) {
					cfg.Mode = "dark"
				}
			case "light_mode":
				if ToBool(v) {
					cfg.Mode = "light"
				}
			case "delay":
				val, err := ctx.evaluate(v)
				if err != nil {
					return err
				}

				cfg.Delay, _ = time.ParseDuration(ToString(val))
			case "block_ads":
				cfg.Ads = ToBool(v)
			case "trackers":
				cfg.Trackers = ToBool(v)
			case "banners":
				cfg.CookieBanners = ToBool(v)
			case "width":
				cfg.Width = ToInt(v)
			case "height":
				cfg.Height = ToInt(v)
			default:
				return fmt.Errorf("unknown type: %s", k)
			}
		}

	default:
		return fmt.Errorf("type is not a request/fetch type: %s", t.Type)
	}

	result, err := cfg.Run()
	if err != nil {
		return err
	}

	return ctx.as(result, cfg.As)
}

func (s *Screenshot) Run() ([]byte, error) {

	time_default := 15 * time.Second
	timeout := s.Timeout
	if timeout == 0 || timeout > time_default {
		timeout = time_default
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	chromeCtx, chromeCancel := chromedp.NewContext(ctx)
	defer chromeCancel()

	var buf []byte

	tasks := chromedp.Tasks{
		chromedp.Navigate(s.Navigate),
	}

	// Set viewport nếu có
	if s.Width > 0 && s.Height > 0 {
		tasks = append(tasks,
			chromedp.EmulateViewport(int64(s.Width), int64(s.Height)),
		)
	}

	// Device preset
	if emu := s.Device; emu != nil {
		tasks = append(tasks, chromedp.Emulate(emu))
	}

	// Dark mode
	if s.Mode == "dark" {
		tasks = append(tasks,
			chromedp.ActionFunc(func(ctx context.Context) error {
				return emulation.SetEmulatedMedia().WithFeatures([]*emulation.MediaFeature{
					{Name: "prefers-color-scheme", Value: "dark"},
				}).Do(ctx)
			}),
		)

	}

	if s.Delay > 0 {
		tasks = append(tasks, chromedp.Sleep(s.Delay))
	}

	// Capture screenshot
	if s.Device != nil {
		tasks = append(tasks, chromedp.CaptureScreenshot(&buf))
	} else {
		tasks = append(tasks, chromedp.FullScreenshot(&buf, 100))
	}

	// Run
	if err := chromedp.Run(chromeCtx, tasks); err != nil {
		return nil, err
	}

	return buf, nil
}

// parseDevice converts a human-friendly device name into chromedp.Device.
func parseDevice(name string) chromedp.Device {
	n := strings.ToLower(strings.ReplaceAll(name, " ", ""))

	switch n {
	case "reset":
		return device.Reset

	// iPhone
	case "iphone4":
		return device.IPhone4
	case "iphone4landscape":
		return device.IPhone4landscape
	case "iphone5":
		return device.IPhone5
	case "iphone5landscape":
		return device.IPhone5landscape
	case "iphone6":
		return device.IPhone6
	case "iphone6landscape":
		return device.IPhone6landscape
	case "iphone6plus":
		return device.IPhone6Plus
	case "iphone6pluslandscape":
		return device.IPhone6Pluslandscape
	case "iphone7":
		return device.IPhone7
	case "iphone7landscape":
		return device.IPhone7landscape
	case "iphone7plus":
		return device.IPhone7Plus
	case "iphone7pluslandscape":
		return device.IPhone7Pluslandscape
	case "iphone8":
		return device.IPhone8
	case "iphone8landscape":
		return device.IPhone8landscape
	case "iphone8plus":
		return device.IPhone8Plus
	case "iphone8pluslandscape":
		return device.IPhone8Pluslandscape
	case "iphonese":
		return device.IPhoneSE
	case "iphonese_landscape":
		return device.IPhoneSElandscape
	case "iphonex":
		return device.IPhoneX
	case "iphonexlandscape":
		return device.IPhoneXlandscape
	case "iphonexr":
		return device.IPhoneXR
	case "iphonexrlandscape":
		return device.IPhoneXRlandscape
	case "iphone11":
		return device.IPhone11
	case "iphone11landscape":
		return device.IPhone11landscape
	case "iphone11pro":
		return device.IPhone11Pro
	case "iphone11prolandscape":
		return device.IPhone11Prolandscape
	case "iphone11promax":
		return device.IPhone11ProMax
	case "iphone11promaxlandscape":
		return device.IPhone11ProMaxlandscape
	case "iphone12":
		return device.IPhone12
	case "iphone12landscape":
		return device.IPhone12landscape
	case "iphone12pro":
		return device.IPhone12Pro
	case "iphone12prolandscape":
		return device.IPhone12Prolandscape
	case "iphone12promax":
		return device.IPhone12ProMax
	case "iphone12promaxlandscape":
		return device.IPhone12ProMaxlandscape
	case "iphone12mini":
		return device.IPhone12Mini
	case "iphone12minilandscape":
		return device.IPhone12Minilandscape
	case "iphone13":
		return device.IPhone13
	case "iphone13landscape":
		return device.IPhone13landscape
	case "iphone13pro":
		return device.IPhone13Pro
	case "iphone13prolandscape":
		return device.IPhone13Prolandscape
	case "iphone13promax":
		return device.IPhone13ProMax
	case "iphone13promaxlandscape":
		return device.IPhone13ProMaxlandscape
	case "iphone13mini":
		return device.IPhone13Mini
	case "iphone13minilandscape":
		return device.IPhone13Minilandscape
	case "iphone14":
		return device.IPhone14
	case "iphone14landscape":
		return device.IPhone14landscape
	case "iphone14plus":
		return device.IPhone14Plus
	case "iphone14pluslandscape":
		return device.IPhone14Pluslandscape
	case "iphone14pro":
		return device.IPhone14Pro
	case "iphone14prolandscape":
		return device.IPhone14Prolandscape
	case "iphone14promax":
		return device.IPhone14ProMax
	case "iphone14promaxlandscape":
		return device.IPhone14ProMaxlandscape
	case "iphone15":
		return device.IPhone15
	case "iphone15landscape":
		return device.IPhone15landscape
	case "iphone15plus":
		return device.IPhone15Plus
	case "iphone15pluslandscape":
		return device.IPhone15Pluslandscape
	case "iphone15pro":
		return device.IPhone15Pro
	case "iphone15prolandscape":
		return device.IPhone15Prolandscape
	case "iphone15promax":
		return device.IPhone15ProMax
	case "iphone15promaxlandscape":
		return device.IPhone15ProMaxlandscape

	// iPad
	case "ipad":
		return device.IPad
	case "ipadlandscape":
		return device.IPadlandscape
	case "ipadpro":
		return device.IPadPro
	case "ipadprolandscape":
		return device.IPadProlandscape
	case "ipadpro11":
		return device.IPadPro11
	case "ipadpro11landscape":
		return device.IPadPro11landscape
	case "ipadgen6":
		return device.IPadgen6
	case "ipadgen6landscape":
		return device.IPadgen6landscape
	case "ipadgen7":
		return device.IPadgen7
	case "ipadgen7landscape":
		return device.IPadgen7landscape
	case "ipadmini":
		return device.IPadMini
	case "ipadminilandscape":
		return device.IPadMinilandscape

	// Galaxy
	case "galaxynote3":
		return device.GalaxyNote3
	case "galaxynote3landscape":
		return device.GalaxyNote3landscape
	case "galaxys5":
		return device.GalaxyS5
	case "galaxys5landscape":
		return device.GalaxyS5landscape
	case "galaxys8":
		return device.GalaxyS8
	case "galaxys8landscape":
		return device.GalaxyS8landscape
	case "galaxys9":
		return device.GalaxyS9
	case "galaxys9landscape":
		return device.GalaxyS9landscape
	case "galaxytabs4":
		return device.GalaxyTabS4
	case "galaxytabs4landscape":
		return device.GalaxyTabS4landscape

	// Pixel
	case "pixel2":
		return device.Pixel2
	case "pixel2landscape":
		return device.Pixel2landscape
	case "pixel2xl":
		return device.Pixel2XL
	case "pixel2xllandscape":
		return device.Pixel2XLlandscape
	case "pixel3":
		return device.Pixel3
	case "pixel3landscape":
		return device.Pixel3landscape
	case "pixel4":
		return device.Pixel4
	case "pixel4landscape":
		return device.Pixel4landscape
	case "pixel4a5g":
		return device.Pixel4a5G
	case "pixel4a5glandscape":
		return device.Pixel4a5Glandscape
	case "pixel5":
		return device.Pixel5
	case "pixel5landscape":
		return device.Pixel5landscape

	// Fallback
	default:
		return device.IPadProlandscape
	}
}

func ToInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int8:
		return int(val)
	case int16:
		return int(val)
	case int32:
		return int(val)
	case int64:
		return int(val)
	case uint:
		return int(val)
	case uint8:
		return int(val)
	case uint16:
		return int(val)
	case uint32:
		return int(val)
	case uint64:
		return int(val)
	case float32:
		return int(val)
	case float64:
		return int(val)
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	// Nếu không chuyển được, trả về 0 (hoặc có thể panic / return error)
	return 0
}

func ToBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case int, int8, int16, int32, int64:
		return ToInt(val) != 0
	case uint, uint8, uint16, uint32, uint64:
		return ToInt(val) != 0
	case float32:
		return val != 0
	case float64:
		return val != 0
	case string:
		switch val {
		case "1", "true", "TRUE", "True", "yes", "YES", "Yes":
			return true
		case "0", "false", "FALSE", "False", "no", "NO", "No":
			return false
		default:
			return false
		}
	default:
		return false
	}
}

func NormalizeURL(input interface{}) string {
	if input == nil {
		return ""
	}

	u := strings.TrimSpace(toString(input))

	if u == "" {
		return ""
	}

	// Nếu không bắt đầu bằng http:// hoặc https:// thì thêm https://
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		u = "https://" + u
	}

	// Nếu không có "/" cuối và không có query/path thì thêm
	// Nhưng nếu đã có path như "/docs" thì giữ nguyên
	if !strings.Contains(u[len("https://"):], "/") && !strings.HasSuffix(u, "/") {
		u += "/"
	}

	return u
}

// chuyển bất kỳ input → string
func toString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	default:
		return strings.TrimSpace(strings.ReplaceAll(strings.Trim(fmt.Sprintf("%v", v), "\""), "\n", " "))
	}
}

// ToURL converts any input into a safe URL-compatible string.
// - string → escape
// - number → string
// - bool → string
// - map[string]interface{} → querystring
// - []any → comma-joined
// - nil → ""
func ToURL(input interface{}) string {
	if input == nil {
		return ""
	}

	v := reflect.ValueOf(input)

	switch v.Kind() {

	case reflect.String:
		return url.QueryEscape(v.String())

	case reflect.Bool:
		return url.QueryEscape(fmt.Sprintf("%v", v.Bool()))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())

	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", v.Float())

	case reflect.Map:
		return mapToQuery(v)

	case reflect.Slice, reflect.Array:
		return sliceToURL(v)

	default:
		// fallback: stringify rồi escape
		return url.QueryEscape(fmt.Sprintf("%v", input))
	}
}

// Encode map to querystring (a=1&b=2)
func mapToQuery(v reflect.Value) string {
	var parts []string

	for _, key := range v.MapKeys() {
		k := fmt.Sprintf("%v", key.Interface())
		val := v.MapIndex(key).Interface()
		parts = append(parts, fmt.Sprintf("%s=%s", url.QueryEscape(k), ToURL(val)))
	}

	return strings.Join(parts, "&")
}

// Encode slice to comma-separated values
func sliceToURL(v reflect.Value) string {
	var parts []string
	for i := 0; i < v.Len(); i++ {
		parts = append(parts, ToURL(v.Index(i).Interface()))
	}
	return strings.Join(parts, ",")
}
