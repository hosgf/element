package assert

import (
	"context"
	"errors"
	"github.com/gogf/gf/v2/frame/g"
)

func Validator(ctx context.Context, data interface{}) error {
	e := g.Validator().Data(data).Run(ctx)
	if e != nil {
		return Error(e.String())
	}
	return nil
}

func Error(message string) error {
	return errors.New(message)
}
