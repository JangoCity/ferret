package eval

import (
	"context"
	"fmt"
	"github.com/MontFerret/ferret/pkg/runtime/core"
	"github.com/MontFerret/ferret/pkg/runtime/values"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/runtime"
)

func PrepareEval(exp string) string {
	return fmt.Sprintf("((function () {%s})())", exp)
}

func ParamString(param string) string {
	return "`" + param + "`"
}

func Eval(client *cdp.Client, exp string, ret bool, async bool) (core.Value, error) {
	args := runtime.
		NewEvaluateArgs(PrepareEval(exp)).
		SetReturnByValue(ret).
		SetAwaitPromise(async)

	out, err := client.Runtime.Evaluate(context.Background(), args)

	if err != nil {
		return values.None, err
	}

	if out.ExceptionDetails != nil {
		ex := out.ExceptionDetails

		return values.None, core.Error(
			core.ErrUnexpected,
			fmt.Sprintf("%s: %s", ex.Text, *ex.Exception.Description),
		)
	}

	if out.Result.Type != "undefined" {
		return values.Unmarshal(out.Result.Value)
	}

	return Unmarshal(&out.Result)
}

func Property(
	ctx context.Context,
	client *cdp.Client,
	objectId runtime.RemoteObjectID,
	propName string,
) (core.Value, error) {
	res, err := client.Runtime.GetProperties(
		ctx,
		runtime.NewGetPropertiesArgs(objectId),
	)

	if err != nil {
		return values.None, err
	}

	if res.ExceptionDetails != nil {
		return values.None, res.ExceptionDetails
	}

	// all props
	if propName == "" {
		var arr *values.Array
		arr = values.NewArray(len(res.Result))

		for _, prop := range res.Result {
			val, err := Unmarshal(prop.Value)

			if err != nil {
				return values.None, err
			}

			arr.Push(val)
		}

		return arr, nil
	}

	for _, prop := range res.Result {
		if prop.Name == propName {
			return Unmarshal(prop.Value)
		}
	}

	return values.None, nil
}

func Unmarshal(obj *runtime.RemoteObject) (core.Value, error) {
	if obj == nil {
		return values.None, nil
	}

	if obj.Type != "undefined" {
		return values.Unmarshal(obj.Value)
	}

	return values.None, nil
}
