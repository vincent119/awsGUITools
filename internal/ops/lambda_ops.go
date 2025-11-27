// Package ops 封裝 AWS 資源操作邏輯（start/stop/reboot 等）。
package ops

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// LambdaAPI 定義 Lambda 操作所需介面，便於測試。
type LambdaAPI interface {
	Invoke(ctx context.Context, params *lambda.InvokeInput, optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error)
}

// LambdaOps 封裝 Lambda 操作。
type LambdaOps struct {
	client LambdaAPI
}

// NewLambdaOps 建立 Lambda 操作服務。
func NewLambdaOps(client LambdaAPI) *LambdaOps {
	return &LambdaOps{client: client}
}

// InvokeResult 代表 Lambda 呼叫結果。
type InvokeResult struct {
	StatusCode      int32
	ExecutedVersion string
	Payload         string
	FunctionError   string
	LogResult       string
}

// TestInvoke 測試呼叫 Lambda 函式。
func (o *LambdaOps) TestInvoke(ctx context.Context, functionName string, payload []byte) (*InvokeResult, error) {
	if o.client == nil {
		return nil, errors.New("lambda client is nil")
	}

	input := &lambda.InvokeInput{
		FunctionName:   aws.String(functionName),
		InvocationType: types.InvocationTypeRequestResponse,
		LogType:        types.LogTypeTail,
	}
	if len(payload) > 0 {
		input.Payload = payload
	}

	resp, err := o.client.Invoke(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("invoke lambda %s: %w", functionName, err)
	}

	result := &InvokeResult{
		StatusCode:      resp.StatusCode,
		ExecutedVersion: aws.ToString(resp.ExecutedVersion),
		FunctionError:   aws.ToString(resp.FunctionError),
		LogResult:       aws.ToString(resp.LogResult),
	}

	if resp.Payload != nil {
		result.Payload = string(resp.Payload)
	}

	return result, nil
}

// TestInvokeWithJSON 使用 JSON 物件作為 payload 測試呼叫。
func (o *LambdaOps) TestInvokeWithJSON(ctx context.Context, functionName string, payloadObj any) (*InvokeResult, error) {
	var payload []byte
	var err error

	if payloadObj != nil {
		payload, err = json.Marshal(payloadObj)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
	}

	return o.TestInvoke(ctx, functionName, payload)
}

// AsyncInvoke 非同步呼叫 Lambda 函式（不等待結果）。
func (o *LambdaOps) AsyncInvoke(ctx context.Context, functionName string, payload []byte) error {
	if o.client == nil {
		return errors.New("lambda client is nil")
	}

	input := &lambda.InvokeInput{
		FunctionName:   aws.String(functionName),
		InvocationType: types.InvocationTypeEvent,
	}
	if len(payload) > 0 {
		input.Payload = payload
	}

	_, err := o.client.Invoke(ctx, input)
	if err != nil {
		return fmt.Errorf("async invoke lambda %s: %w", functionName, err)
	}

	return nil
}
