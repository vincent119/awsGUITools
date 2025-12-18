package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"

	"github.com/vincent119/awsGUITools/internal/ops"
)

// mockRDSClient 實作 ops.RDSAPI 介面。
type mockRDSClient struct {
	startCalled  bool
	stopCalled   bool
	rebootCalled bool
	multiAZ      bool
	isReplica    bool
	isAurora     bool
	status       string
}

func (m *mockRDSClient) StartDBInstance(ctx context.Context, params *rds.StartDBInstanceInput, optFns ...func(*rds.Options)) (*rds.StartDBInstanceOutput, error) {
	m.startCalled = true
	return &rds.StartDBInstanceOutput{}, nil
}

func (m *mockRDSClient) StopDBInstance(ctx context.Context, params *rds.StopDBInstanceInput, optFns ...func(*rds.Options)) (*rds.StopDBInstanceOutput, error) {
	m.stopCalled = true
	return &rds.StopDBInstanceOutput{}, nil
}

func (m *mockRDSClient) RebootDBInstance(ctx context.Context, params *rds.RebootDBInstanceInput, optFns ...func(*rds.Options)) (*rds.RebootDBInstanceOutput, error) {
	m.rebootCalled = true
	return &rds.RebootDBInstanceOutput{}, nil
}

func (m *mockRDSClient) DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
	inst := types.DBInstance{
		DBInstanceIdentifier: aws.String("mydb"),
		DBInstanceStatus:     aws.String(m.status),
		MultiAZ:              aws.Bool(m.multiAZ),
	}
	if m.isReplica {
		inst.ReadReplicaSourceDBInstanceIdentifier = aws.String("source-db")
	}
	if m.isAurora {
		inst.DBClusterIdentifier = aws.String("aurora-cluster")
	}
	return &rds.DescribeDBInstancesOutput{
		DBInstances: []types.DBInstance{inst},
	}, nil
}

func TestRDSOpsStartDBInstance(t *testing.T) {
	mock := &mockRDSClient{}
	opsClient := ops.NewRDSOps(mock)

	err := opsClient.StartDBInstance(context.Background(), "mydb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.startCalled {
		t.Error("StartDBInstance was not called")
	}
}

func TestRDSOpsStopDBInstance(t *testing.T) {
	mock := &mockRDSClient{}
	opsClient := ops.NewRDSOps(mock)

	err := opsClient.StopDBInstance(context.Background(), "mydb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.stopCalled {
		t.Error("StopDBInstance was not called")
	}
}

func TestRDSOpsCanStop_Replica(t *testing.T) {
	mock := &mockRDSClient{isReplica: true}
	opsClient := ops.NewRDSOps(mock)

	canStop, msg, err := opsClient.CanStop(context.Background(), "mydb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if canStop {
		t.Error("expected canStop to be false for replica")
	}
	if msg == "" {
		t.Error("expected warning message for replica")
	}
}

func TestRDSOpsCanStop_Aurora(t *testing.T) {
	mock := &mockRDSClient{isAurora: true}
	opsClient := ops.NewRDSOps(mock)

	canStop, msg, err := opsClient.CanStop(context.Background(), "mydb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if canStop {
		t.Error("expected canStop to be false for Aurora member")
	}
	if msg == "" {
		t.Error("expected warning message for Aurora")
	}
}

func TestRDSOpsCanStop_MultiAZ(t *testing.T) {
	mock := &mockRDSClient{multiAZ: true}
	opsClient := ops.NewRDSOps(mock)

	canStop, msg, err := opsClient.CanStop(context.Background(), "mydb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !canStop {
		t.Error("expected canStop to be true for Multi-AZ (with warning)")
	}
	if msg == "" {
		t.Error("expected warning message for Multi-AZ")
	}
}
