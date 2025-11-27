package ops

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
)

// RDSAPI 定義 RDS 操作所需介面，便於測試。
type RDSAPI interface {
	StartDBInstance(ctx context.Context, params *rds.StartDBInstanceInput, optFns ...func(*rds.Options)) (*rds.StartDBInstanceOutput, error)
	StopDBInstance(ctx context.Context, params *rds.StopDBInstanceInput, optFns ...func(*rds.Options)) (*rds.StopDBInstanceOutput, error)
	RebootDBInstance(ctx context.Context, params *rds.RebootDBInstanceInput, optFns ...func(*rds.Options)) (*rds.RebootDBInstanceOutput, error)
	DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
}

// RDSOps 封裝 RDS 操作。
type RDSOps struct {
	client RDSAPI
}

// NewRDSOps 建立 RDS 操作服務。
func NewRDSOps(client RDSAPI) *RDSOps {
	return &RDSOps{client: client}
}

// CanStop 檢查 RDS 執行個體是否可停止（非 Multi-AZ、非 Read Replica 等限制）。
func (o *RDSOps) CanStop(ctx context.Context, dbInstanceID string) (bool, string, error) {
	if o.client == nil {
		return false, "", errors.New("rds client is nil")
	}
	resp, err := o.client.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(dbInstanceID),
	})
	if err != nil {
		return false, "", fmt.Errorf("describe db instance %s: %w", dbInstanceID, err)
	}
	if len(resp.DBInstances) == 0 {
		return false, "", fmt.Errorf("db instance %s not found", dbInstanceID)
	}
	inst := resp.DBInstances[0]

	// Multi-AZ 可停止（2020 年後支援），但需提醒
	if inst.MultiAZ != nil && *inst.MultiAZ {
		return true, "此為 Multi-AZ 執行個體，停止後會影響高可用性", nil
	}
	// Read Replica 不可停止
	if inst.ReadReplicaSourceDBInstanceIdentifier != nil {
		return false, "Read Replica 不可直接停止", nil
	}
	// Aurora 叢集成員需透過叢集操作
	if inst.DBClusterIdentifier != nil && *inst.DBClusterIdentifier != "" {
		return false, "Aurora 叢集成員需透過叢集操作", nil
	}
	return true, "", nil
}

// StartDBInstance 啟動 RDS 執行個體。
func (o *RDSOps) StartDBInstance(ctx context.Context, dbInstanceID string) error {
	if o.client == nil {
		return errors.New("rds client is nil")
	}
	_, err := o.client.StartDBInstance(ctx, &rds.StartDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceID),
	})
	if err != nil {
		return fmt.Errorf("start db instance %s: %w", dbInstanceID, err)
	}
	return nil
}

// StopDBInstance 停止 RDS 執行個體。
func (o *RDSOps) StopDBInstance(ctx context.Context, dbInstanceID string) error {
	if o.client == nil {
		return errors.New("rds client is nil")
	}
	_, err := o.client.StopDBInstance(ctx, &rds.StopDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceID),
	})
	if err != nil {
		return fmt.Errorf("stop db instance %s: %w", dbInstanceID, err)
	}
	return nil
}

// RebootDBInstance 重新啟動 RDS 執行個體。
func (o *RDSOps) RebootDBInstance(ctx context.Context, dbInstanceID string, forceFailover bool) error {
	if o.client == nil {
		return errors.New("rds client is nil")
	}
	_, err := o.client.RebootDBInstance(ctx, &rds.RebootDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceID),
		ForceFailover:        aws.Bool(forceFailover),
	})
	if err != nil {
		return fmt.Errorf("reboot db instance %s: %w", dbInstanceID, err)
	}
	return nil
}

// WaitForStatus 輪詢等待 RDS 執行個體達到指定狀態。
func (o *RDSOps) WaitForStatus(ctx context.Context, dbInstanceID string, targetStatus string, timeout time.Duration) error {
	if o.client == nil {
		return errors.New("rds client is nil")
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	deadline := time.Now().Add(timeout)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for db instance %s to reach status %s", dbInstanceID, targetStatus)
			}
			resp, err := o.client.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{
				DBInstanceIdentifier: aws.String(dbInstanceID),
			})
			if err != nil {
				return fmt.Errorf("describe db instance %s: %w", dbInstanceID, err)
			}
			for _, inst := range resp.DBInstances {
				if inst.DBInstanceStatus != nil && *inst.DBInstanceStatus == targetStatus {
					return nil
				}
			}
		}
	}
}

// StatusAvailable 常用狀態常數。
const (
	StatusAvailable = "available"
	StatusStopped   = "stopped"
	StatusStarting  = "starting"
	StatusStopping  = "stopping"
)

// Ensure types import is used.
var _ types.DBInstance
