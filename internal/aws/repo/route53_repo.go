package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"

	"github.com/vin/ck123gogo/internal/models"
)

// Route53Repository 封裝 Route53 查詢邏輯。
type Route53Repository struct{}

// NewRoute53Repository 建立 Route53Repository。
func NewRoute53Repository() *Route53Repository {
	return &Route53Repository{}
}

// ListHostedZones 列出所有 Hosted Zones。
func (r *Route53Repository) ListHostedZones(ctx context.Context, client *route53.Client) ([]models.Route53HostedZone, error) {
	if client == nil {
		return nil, fmt.Errorf("route53 client is nil")
	}

	var zones []models.Route53HostedZone
	paginator := route53.NewListHostedZonesPaginator(client, &route53.ListHostedZonesInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list hosted zones: %w", err)
		}

		for _, zone := range page.HostedZones {
			id := aws.ToString(zone.Id)
			// 移除 /hostedzone/ 前綴
			id = strings.TrimPrefix(id, "/hostedzone/")

			comment := ""
			if zone.Config != nil {
				comment = aws.ToString(zone.Config.Comment)
			}

			zones = append(zones, models.Route53HostedZone{
				ID:          id,
				Name:        aws.ToString(zone.Name),
				RecordCount: aws.ToInt64(zone.ResourceRecordSetCount),
				IsPrivate:   zone.Config != nil && zone.Config.PrivateZone,
				Comment:     comment,
			})
		}
	}

	return zones, nil
}

// ListRecords 列出指定 Hosted Zone 的所有 Records。
func (r *Route53Repository) ListRecords(ctx context.Context, client *route53.Client, zoneID string) ([]models.Route53Record, error) {
	if client == nil {
		return nil, fmt.Errorf("route53 client is nil")
	}

	var records []models.Route53Record
	paginator := route53.NewListResourceRecordSetsPaginator(client, &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list resource record sets: %w", err)
		}

		for _, rr := range page.ResourceRecordSets {
			record := models.Route53Record{
				Name: aws.ToString(rr.Name),
				Type: string(rr.Type),
				TTL:  aws.ToInt64(rr.TTL),
			}

			// 處理一般 record values
			for _, rv := range rr.ResourceRecords {
				record.Values = append(record.Values, aws.ToString(rv.Value))
			}

			// 處理 Alias record
			if rr.AliasTarget != nil {
				record.AliasTarget = aws.ToString(rr.AliasTarget.DNSName)
			}

			records = append(records, record)
		}
	}

	return records, nil
}

// GetRecordTypeColor 根據 record type 回傳顏色標籤。
func GetRecordTypeColor(recordType types.RRType) string {
	switch recordType {
	case types.RRTypeA:
		return "[green]"
	case types.RRTypeAaaa:
		return "[cyan]"
	case types.RRTypeCname:
		return "[yellow]"
	case types.RRTypeMx:
		return "[magenta]"
	case types.RRTypeTxt:
		return "[blue]"
	case types.RRTypeNs:
		return "[orange]"
	case types.RRTypeSoa:
		return "[gray]"
	default:
		return "[white]"
	}
}

