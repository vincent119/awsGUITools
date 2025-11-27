// Package ops 封裝 AWS 資源操作邏輯（start/stop/reboot 等）。
package ops

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2API 定義 EC2 操作所需介面，便於測試。
type EC2API interface {
	StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error)
	StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error)
	RebootInstances(ctx context.Context, params *ec2.RebootInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RebootInstancesOutput, error)
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

// EC2Ops 封裝 EC2 操作。
type EC2Ops struct {
	client EC2API
}

// NewEC2Ops 建立 EC2 操作服務。
func NewEC2Ops(client EC2API) *EC2Ops {
	return &EC2Ops{client: client}
}

// StartInstance 啟動 EC2 執行個體。
func (o *EC2Ops) StartInstance(ctx context.Context, instanceID string, dryRun bool) error {
	if o.client == nil {
		return errors.New("ec2 client is nil")
	}
	_, err := o.client.StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
		DryRun:      aws.Bool(dryRun),
	})
	if err != nil {
		return fmt.Errorf("start instance %s: %w", instanceID, err)
	}
	return nil
}

// StopInstance 停止 EC2 執行個體。
func (o *EC2Ops) StopInstance(ctx context.Context, instanceID string, dryRun bool) error {
	if o.client == nil {
		return errors.New("ec2 client is nil")
	}
	_, err := o.client.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
		DryRun:      aws.Bool(dryRun),
	})
	if err != nil {
		return fmt.Errorf("stop instance %s: %w", instanceID, err)
	}
	return nil
}

// RebootInstance 重新啟動 EC2 執行個體。
func (o *EC2Ops) RebootInstance(ctx context.Context, instanceID string, dryRun bool) error {
	if o.client == nil {
		return errors.New("ec2 client is nil")
	}
	_, err := o.client.RebootInstances(ctx, &ec2.RebootInstancesInput{
		InstanceIds: []string{instanceID},
		DryRun:      aws.Bool(dryRun),
	})
	if err != nil {
		return fmt.Errorf("reboot instance %s: %w", instanceID, err)
	}
	return nil
}

// WaitForState 輪詢等待執行個體達到指定狀態。
func (o *EC2Ops) WaitForState(ctx context.Context, instanceID string, targetState types.InstanceStateName, timeout time.Duration) error {
	if o.client == nil {
		return errors.New("ec2 client is nil")
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	deadline := time.Now().Add(timeout)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for instance %s to reach state %s", instanceID, targetState)
			}
			resp, err := o.client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
				InstanceIds: []string{instanceID},
			})
			if err != nil {
				return fmt.Errorf("describe instance %s: %w", instanceID, err)
			}
			for _, res := range resp.Reservations {
				for _, inst := range res.Instances {
					if inst.State != nil && inst.State.Name == targetState {
						return nil
					}
				}
			}
		}
	}
}
