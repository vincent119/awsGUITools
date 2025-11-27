// Package repo 提供 AWS 資源的查詢與轉換功能。
package repo

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/vin/ck123gogo/internal/models"
)

// EC2Repository 負責查詢 EC2 實例與關聯資訊。
type EC2Repository struct{}

func NewEC2Repository() *EC2Repository {
	return &EC2Repository{}
}

// ListInstances 以 DescribeInstances 取得所有 EC2 實例（支援 paginator）。
func (r *EC2Repository) ListInstances(ctx context.Context, client *ec2.Client, input *ec2.DescribeInstancesInput) ([]models.EC2Instance, error) {
	if client == nil {
		return nil, fmt.Errorf("ec2 client is nil")
	}
	if input == nil {
		input = &ec2.DescribeInstancesInput{}
	}

	paginator := ec2.NewDescribeInstancesPaginator(client, input)
	var instances []models.EC2Instance

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("describe instances: %w", err)
		}

		for _, reservation := range page.Reservations {
			for _, inst := range reservation.Instances {
				instances = append(instances, convertEC2Instance(inst))
			}
		}
	}

	return instances, nil
}

func convertEC2Instance(inst ec2types.Instance) models.EC2Instance {
	var sg []string
	for _, g := range inst.SecurityGroups {
		if g.GroupName != nil {
			sg = append(sg, *g.GroupName)
		} else if g.GroupId != nil {
			sg = append(sg, *g.GroupId)
		}
	}

	var volumes []models.EBSVolume
	for _, bd := range inst.BlockDeviceMappings {
		if bd.Ebs == nil {
			continue
		}
		vol := models.EBSVolume{
			ID:         deref(bd.Ebs.VolumeId),
			DeviceName: deref(bd.DeviceName),
			State:      string(bd.Ebs.Status),
		}
		volumes = append(volumes, vol)
	}

	return models.EC2Instance{
		ID:               deref(inst.InstanceId),
		Name:             extractTag(inst.Tags, "Name"),
		State:            string(inst.State.Name),
		InstanceType:     string(inst.InstanceType),
		AvailabilityZone: deref(inst.Placement.AvailabilityZone),
		PrivateIP:        deref(inst.PrivateIpAddress),
		PublicIP:         deref(inst.PublicIpAddress),
		VpcID:            deref(inst.VpcId),
		SubnetID:         deref(inst.SubnetId),
		SecurityGroups:   sg,
		IAMRole:          extractIAMRole(inst.IamInstanceProfile),
		Volumes:          volumes,
		Tags:             convertTags(inst.Tags),
	}
}

func extractIAMRole(profile *ec2types.IamInstanceProfile) string {
	if profile == nil || profile.Arn == nil {
		return ""
	}
	return *profile.Arn
}

func convertTags(tags []ec2types.Tag) models.TagMap {
	result := make(models.TagMap, len(tags))
	for _, tag := range tags {
		if tag.Key == nil {
			continue
		}
		result[*tag.Key] = deref(tag.Value)
	}
	return result
}

func extractTag(tags []ec2types.Tag, key string) string {
	for _, tag := range tags {
		if tag.Key != nil && *tag.Key == key {
			return deref(tag.Value)
		}
	}
	return ""
}
