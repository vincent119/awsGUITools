package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go/middleware"

	"github.com/vin/ck123gogo/internal/ops"
)

// mockEC2Client 實作 ops.EC2API 介面。
type mockEC2Client struct {
	startCalled  bool
	stopCalled   bool
	rebootCalled bool
	state        types.InstanceStateName
}

func (m *mockEC2Client) StartInstances(ctx context.Context, params *ec2.StartInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StartInstancesOutput, error) {
	m.startCalled = true
	return &ec2.StartInstancesOutput{
		ResultMetadata: middleware.Metadata{},
	}, nil
}

func (m *mockEC2Client) StopInstances(ctx context.Context, params *ec2.StopInstancesInput, optFns ...func(*ec2.Options)) (*ec2.StopInstancesOutput, error) {
	m.stopCalled = true
	return &ec2.StopInstancesOutput{}, nil
}

func (m *mockEC2Client) RebootInstances(ctx context.Context, params *ec2.RebootInstancesInput, optFns ...func(*ec2.Options)) (*ec2.RebootInstancesOutput, error) {
	m.rebootCalled = true
	return &ec2.RebootInstancesOutput{}, nil
}

func (m *mockEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return &ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						InstanceId: aws.String("i-12345"),
						State:      &types.InstanceState{Name: m.state},
					},
				},
			},
		},
	}, nil
}

func TestEC2OpsStartInstance(t *testing.T) {
	mock := &mockEC2Client{}
	opsClient := ops.NewEC2Ops(mock)

	err := opsClient.StartInstance(context.Background(), "i-12345", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.startCalled {
		t.Error("StartInstances was not called")
	}
}

func TestEC2OpsStopInstance(t *testing.T) {
	mock := &mockEC2Client{}
	opsClient := ops.NewEC2Ops(mock)

	err := opsClient.StopInstance(context.Background(), "i-12345", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.stopCalled {
		t.Error("StopInstances was not called")
	}
}

func TestEC2OpsRebootInstance(t *testing.T) {
	mock := &mockEC2Client{}
	opsClient := ops.NewEC2Ops(mock)

	err := opsClient.RebootInstance(context.Background(), "i-12345", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.rebootCalled {
		t.Error("RebootInstances was not called")
	}
}
