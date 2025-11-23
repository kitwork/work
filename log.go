package work

import "fmt"

type Log struct {
	Message string `work:"message,required,fallback"` // log message
}

func (t *Work) Log(ctx *Context) error {
	if t.Type != TypeLog {
		return fmt.Errorf("type is not log")
	}

	cfg := Log{}
	if err := t.classify(ctx, &cfg); err != nil {
		return err
	}

	fmt.Println("→ [log]", cfg.Message)
	return nil
}
