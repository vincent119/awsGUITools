package aws_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// stubEC2Client 為之後整合測試建立的 smithy stub 範例。
type stubEC2Client struct {
	ec2.Client
	resp *ec2.DescribeInstancesOutput
}

func (s *stubEC2Client) DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return s.resp, nil
}

func TestStubEC2Client(t *testing.T) {
	stub := &stubEC2Client{
		resp: &ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{InstanceId: ptr("i-stub123"), InstanceType: types.InstanceTypeT3Small},
					},
				},
			},
		},
	}

	out, err := stub.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{})
	if err != nil {
		t.Fatalf("DescribeInstances returned error: %v", err)
	}
	if len(out.Reservations) != 1 || len(out.Reservations[0].Instances) != 1 {
		t.Fatalf("unexpected stub response: %#v", out)
	}
}

func ptr[T any](v T) *T {
	return &v
}
