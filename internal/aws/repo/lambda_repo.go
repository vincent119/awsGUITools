package repo

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"

	"github.com/vin/ck123gogo/internal/models"
)

type LambdaRepository struct{}

func NewLambdaRepository() *LambdaRepository {
	return &LambdaRepository{}
}

func (r *LambdaRepository) ListFunctions(ctx context.Context, client *lambda.Client, input *lambda.ListFunctionsInput) ([]models.LambdaFunction, error) {
	if client == nil {
		return nil, fmt.Errorf("lambda client is nil")
	}
	if input == nil {
		input = &lambda.ListFunctionsInput{}
	}

	paginator := lambda.NewListFunctionsPaginator(client, input)
	var functions []models.LambdaFunction

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list lambda functions: %w", err)
		}

		for _, fn := range page.Functions {
			functions = append(functions, convertLambdaFunction(fn))
		}
	}

	return functions, nil
}

func convertLambdaFunction(fn lambdatypes.FunctionConfiguration) models.LambdaFunction {
	return models.LambdaFunction{
		Name:         deref(fn.FunctionName),
		ARN:          deref(fn.FunctionArn),
		Runtime:      string(fn.Runtime),
		MemoryMB:     deref(fn.MemorySize),
		TimeoutSec:   deref(fn.Timeout),
		Role:         deref(fn.Role),
		EnvVars:      convertEnv(fn.Environment),
		Triggers:     []string{},
		Tags:         nil,
		LastModified: deref(fn.LastModified),
	}
}

func convertEnv(env *lambdatypes.EnvironmentResponse) map[string]string {
	if env == nil || env.Variables == nil {
		return nil
	}
	result := make(map[string]string, len(env.Variables))
	for k, v := range env.Variables {
		result[k] = v
	}
	return result
}
