package validator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	jd "github.com/josephburnett/jd/lib"
	"github.com/pkg/errors"
	"github.com/rerost/bq-table-validator/domain/tablemock"
	"github.com/rerost/bq-table-validator/types"
)

type Validator interface {
	Valid(ctx context.Context, valid types.Validate) (string, error)
}

type Middleware interface {
	Query(ctx context.Context, sql string) ([]map[string]interface{}, error)
}

type validatorImpl struct {
	middleware Middleware
	tm         tablemock.TableMock
}

func NewValidator(middleware Middleware, tm tablemock.TableMock) Validator {
	return validatorImpl{
		middleware: middleware,
		tm:         tm,
	}
}

func (v validatorImpl) Valid(ctx context.Context, validate types.Validate) (string, error) {
	sql, err := v.tm.Mock(ctx, validate.SQL, validate.Mocks)
	if err != nil {
		return "", errors.WithStack(err)
	}

	queryResult, err := v.middleware.Query(ctx, sql)
	if err != nil {
		return "", errors.WithStack(err)
	}

	bqueryResultJSON, err := json.Marshal(queryResult)
	if err != nil {
		return "", errors.WithStack(err)
	}

	out, err := jd.ReadJsonString(string(bqueryResultJSON))
	if err != nil {
		return "", errors.WithStack(err)
	}

	expect, err := jd.ReadJsonString(validate.Expect)
	if err != nil {
		return "", errors.WithStack(err)
	}

	if diff := out.Diff(expect).Render(); diff != "" {
		return diff, nil
	}

	return "", nil
}

func FormatJSON(j string) (string, error) {
	buf := bytes.NewBuffer([]byte{})
	err := json.Indent(buf, []byte(j), "", "  ")
	if err != nil {
		return "", errors.WithStack(err)
	}

	fmt.Println(buf.String())
	return buf.String(), nil
}
