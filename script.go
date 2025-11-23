package work

import (
	"errors"
	"fmt"
)

type Script struct {
	Run   string `work:"url,required,fallback"`       // URL endpoint
	Agent Type   `work:"agent,ignore" default:"http"` // http | client

}

func (t *Work) Script(ctx *Context) error {
	if t.Type != TypeScript {
		return errors.New("type is not script")
	}

	fmt.Println("→ [script] chạy script ...")
	return nil
}
