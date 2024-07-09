package importer

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aaronland/go-aws-lambda"
)

const LAMBDA_SCHEME string = "lambda"

type LambdaImporter struct {
	Importer
	lambda_func *lambda.LambdaFunction
}

type ImportEvent struct {
	Ids []int64 `json:"ids"`
}

func init() {
	ctx := context.Background()
	RegisterImporter(ctx, LAMBDA_SCHEME, NewLambdaImporter)
}

func NewLambdaImporter(ctx context.Context, uri string) (Importer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	func_name := u.Host

	q := u.Query()

	region := q.Get("region")
	credentials := q.Get("credentials")
	func_type := q.Get("type")

	if func_type == "" {
		func_type = "Event"
	}

	dsn := fmt.Sprintf("region=%s credentials=%s", region, credentials)

	lambda_func, err := lambda.NewLambdaFunctionWithDSN(dsn, func_name, func_type)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new lambda function, %w", err)
	}

	i := &LambdaImporter{
		lambda_func: lambda_func,
	}

	return i, nil
}

func (i *LambdaImporter) ImportIDs(ctx context.Context, ids ...int64) error {

	import_ev := ImportEvent{
		Ids: ids,
	}

	_, err := i.lambda_func.Invoke(ctx, import_ev)

	if err != nil {
		return fmt.Errorf("Failed to invoke Lambda function, %w", err)
	}

	return nil
}
