package repo

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"

	"github.com/vincent119/awsGUITools/internal/models"
)

type RDSRepository struct{}

func NewRDSRepository() *RDSRepository {
	return &RDSRepository{}
}

func (r *RDSRepository) ListInstances(ctx context.Context, client *rds.Client, input *rds.DescribeDBInstancesInput) ([]models.RDSInstance, error) {
	if client == nil {
		return nil, fmt.Errorf("rds client is nil")
	}
	if input == nil {
		input = &rds.DescribeDBInstancesInput{}
	}

	paginator := rds.NewDescribeDBInstancesPaginator(client, input)
	var instances []models.RDSInstance

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("describe db instances: %w", err)
		}

		for _, inst := range page.DBInstances {
			instances = append(instances, convertRDSInstance(inst))
		}
	}

	return instances, nil
}

func convertRDSInstance(inst rdstypes.DBInstance) models.RDSInstance {
	var parameterGroups []string
	for _, grp := range inst.DBParameterGroups {
		if grp.DBParameterGroupName != nil {
			parameterGroups = append(parameterGroups, *grp.DBParameterGroupName)
		}
	}

	var sg []string
	for _, vpc := range inst.VpcSecurityGroups {
		if vpc.VpcSecurityGroupId != nil {
			sg = append(sg, *vpc.VpcSecurityGroupId)
		}
	}

	return models.RDSInstance{
		ID:             deref(inst.DBInstanceIdentifier),
		Engine:         deref(inst.Engine),
		EngineVersion:  deref(inst.EngineVersion),
		MultiAZ:        aws.ToBool(inst.MultiAZ),
		Endpoint:       extractEndpoint(inst.Endpoint),
		SubnetGroup:    extractSubnetGroup(inst.DBSubnetGroup),
		ParameterGroup: parameterGroups,
		SecurityGroups: sg,
		Tags:           nil,
	}
}

func extractEndpoint(endpoint *rdstypes.Endpoint) string {
	if endpoint == nil || endpoint.Address == nil {
		return ""
	}
	port := ""
	if endpoint.Port != nil {
		port = fmt.Sprintf(":%d", *endpoint.Port)
	}
	return *endpoint.Address + port
}

func extractSubnetGroup(group *rdstypes.DBSubnetGroup) string {
	if group == nil || group.DBSubnetGroupName == nil {
		return ""
	}
	return *group.DBSubnetGroupName
}
