package aws_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"

	"github.com/vin/ck123gogo/internal/aws/repo"
	"github.com/vin/ck123gogo/internal/models"
)

func TestRoute53Repository_NilClient(t *testing.T) {
	r := repo.NewRoute53Repository()

	t.Run("ListHostedZones with nil client", func(t *testing.T) {
		_, err := r.ListHostedZones(context.Background(), nil)
		if err == nil {
			t.Error("expected error for nil client, got nil")
		}
	})

	t.Run("ListRecords with nil client", func(t *testing.T) {
		_, err := r.ListRecords(context.Background(), nil, "zone-id")
		if err == nil {
			t.Error("expected error for nil client, got nil")
		}
	})
}

func TestParseHostedZone(t *testing.T) {
	tests := []struct {
		name     string
		zone     types.HostedZone
		wantID   string
		wantName string
		wantPrivate bool
	}{
		{
			name: "public zone with /hostedzone/ prefix",
			zone: types.HostedZone{
				Id:                     aws.String("/hostedzone/Z1234567890ABC"),
				Name:                   aws.String("example.com."),
				ResourceRecordSetCount: aws.Int64(10),
				Config: &types.HostedZoneConfig{
					PrivateZone: false,
					Comment:     aws.String("Production zone"),
				},
			},
			wantID:      "Z1234567890ABC",
			wantName:    "example.com.",
			wantPrivate: false,
		},
		{
			name: "private zone",
			zone: types.HostedZone{
				Id:                     aws.String("/hostedzone/ZPRIVATE123"),
				Name:                   aws.String("internal.local."),
				ResourceRecordSetCount: aws.Int64(5),
				Config: &types.HostedZoneConfig{
					PrivateZone: true,
					Comment:     aws.String("Internal VPC zone"),
				},
			},
			wantID:      "ZPRIVATE123",
			wantName:    "internal.local.",
			wantPrivate: true,
		},
		{
			name: "zone without config",
			zone: types.HostedZone{
				Id:                     aws.String("/hostedzone/ZNOCONFIG"),
				Name:                   aws.String("noconfig.com."),
				ResourceRecordSetCount: aws.Int64(2),
				Config:                 nil,
			},
			wantID:      "ZNOCONFIG",
			wantName:    "noconfig.com.",
			wantPrivate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 解析 zone ID（移除前綴）
			id := aws.ToString(tt.zone.Id)
			if len(id) > 12 && id[:12] == "/hostedzone/" {
				id = id[12:]
			}

			if id != tt.wantID {
				t.Errorf("zone ID = %q, want %q", id, tt.wantID)
			}

			name := aws.ToString(tt.zone.Name)
			if name != tt.wantName {
				t.Errorf("zone name = %q, want %q", name, tt.wantName)
			}

			isPrivate := tt.zone.Config != nil && tt.zone.Config.PrivateZone
			if isPrivate != tt.wantPrivate {
				t.Errorf("isPrivate = %v, want %v", isPrivate, tt.wantPrivate)
			}
		})
	}
}

func TestParseResourceRecord(t *testing.T) {
	tests := []struct {
		name           string
		recordSet      types.ResourceRecordSet
		wantName       string
		wantType       string
		wantTTL        int64
		wantValues     []string
		wantAliasTarget string
	}{
		{
			name: "A record with single value",
			recordSet: types.ResourceRecordSet{
				Name: aws.String("api.example.com."),
				Type: types.RRTypeA,
				TTL:  aws.Int64(300),
				ResourceRecords: []types.ResourceRecord{
					{Value: aws.String("1.2.3.4")},
				},
			},
			wantName:   "api.example.com.",
			wantType:   "A",
			wantTTL:    300,
			wantValues: []string{"1.2.3.4"},
		},
		{
			name: "A record with multiple values",
			recordSet: types.ResourceRecordSet{
				Name: aws.String("lb.example.com."),
				Type: types.RRTypeA,
				TTL:  aws.Int64(60),
				ResourceRecords: []types.ResourceRecord{
					{Value: aws.String("1.2.3.4")},
					{Value: aws.String("5.6.7.8")},
					{Value: aws.String("9.10.11.12")},
				},
			},
			wantName:   "lb.example.com.",
			wantType:   "A",
			wantTTL:    60,
			wantValues: []string{"1.2.3.4", "5.6.7.8", "9.10.11.12"},
		},
		{
			name: "CNAME record",
			recordSet: types.ResourceRecordSet{
				Name: aws.String("www.example.com."),
				Type: types.RRTypeCname,
				TTL:  aws.Int64(3600),
				ResourceRecords: []types.ResourceRecord{
					{Value: aws.String("example.com.")},
				},
			},
			wantName:   "www.example.com.",
			wantType:   "CNAME",
			wantTTL:    3600,
			wantValues: []string{"example.com."},
		},
		{
			name: "MX record",
			recordSet: types.ResourceRecordSet{
				Name: aws.String("example.com."),
				Type: types.RRTypeMx,
				TTL:  aws.Int64(3600),
				ResourceRecords: []types.ResourceRecord{
					{Value: aws.String("10 mail1.example.com.")},
					{Value: aws.String("20 mail2.example.com.")},
				},
			},
			wantName:   "example.com.",
			wantType:   "MX",
			wantTTL:    3600,
			wantValues: []string{"10 mail1.example.com.", "20 mail2.example.com."},
		},
		{
			name: "TXT record",
			recordSet: types.ResourceRecordSet{
				Name: aws.String("example.com."),
				Type: types.RRTypeTxt,
				TTL:  aws.Int64(300),
				ResourceRecords: []types.ResourceRecord{
					{Value: aws.String("\"v=spf1 include:_spf.google.com ~all\"")},
				},
			},
			wantName:   "example.com.",
			wantType:   "TXT",
			wantTTL:    300,
			wantValues: []string{"\"v=spf1 include:_spf.google.com ~all\""},
		},
		{
			name: "Alias record (no TTL)",
			recordSet: types.ResourceRecordSet{
				Name: aws.String("cdn.example.com."),
				Type: types.RRTypeA,
				AliasTarget: &types.AliasTarget{
					DNSName:              aws.String("d111111abcdef8.cloudfront.net."),
					HostedZoneId:         aws.String("Z2FDTNDATAQYW2"),
					EvaluateTargetHealth: false,
				},
			},
			wantName:        "cdn.example.com.",
			wantType:        "A",
			wantTTL:         0,
			wantValues:      nil,
			wantAliasTarget: "d111111abcdef8.cloudfront.net.",
		},
		{
			name: "NS record",
			recordSet: types.ResourceRecordSet{
				Name: aws.String("example.com."),
				Type: types.RRTypeNs,
				TTL:  aws.Int64(172800),
				ResourceRecords: []types.ResourceRecord{
					{Value: aws.String("ns-1234.awsdns-12.org.")},
					{Value: aws.String("ns-5678.awsdns-34.co.uk.")},
					{Value: aws.String("ns-9012.awsdns-56.com.")},
					{Value: aws.String("ns-3456.awsdns-78.net.")},
				},
			},
			wantName:   "example.com.",
			wantType:   "NS",
			wantTTL:    172800,
			wantValues: []string{
				"ns-1234.awsdns-12.org.",
				"ns-5678.awsdns-34.co.uk.",
				"ns-9012.awsdns-56.com.",
				"ns-3456.awsdns-78.net.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := parseRecordSet(tt.recordSet)

			if record.Name != tt.wantName {
				t.Errorf("name = %q, want %q", record.Name, tt.wantName)
			}
			if record.Type != tt.wantType {
				t.Errorf("type = %q, want %q", record.Type, tt.wantType)
			}
			if record.TTL != tt.wantTTL {
				t.Errorf("TTL = %d, want %d", record.TTL, tt.wantTTL)
			}
			if len(record.Values) != len(tt.wantValues) {
				t.Errorf("values count = %d, want %d", len(record.Values), len(tt.wantValues))
			} else {
				for i, v := range record.Values {
					if v != tt.wantValues[i] {
						t.Errorf("values[%d] = %q, want %q", i, v, tt.wantValues[i])
					}
				}
			}
			if record.AliasTarget != tt.wantAliasTarget {
				t.Errorf("aliasTarget = %q, want %q", record.AliasTarget, tt.wantAliasTarget)
			}
		})
	}
}

// parseRecordSet 解析 ResourceRecordSet 為 Route53Record model。
func parseRecordSet(rr types.ResourceRecordSet) models.Route53Record {
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

	return record
}

func TestRecordTypeString(t *testing.T) {
	tests := []struct {
		recordType types.RRType
		want       string
	}{
		{types.RRTypeA, "A"},
		{types.RRTypeAaaa, "AAAA"},
		{types.RRTypeCname, "CNAME"},
		{types.RRTypeMx, "MX"},
		{types.RRTypeTxt, "TXT"},
		{types.RRTypeNs, "NS"},
		{types.RRTypeSoa, "SOA"},
		{types.RRTypeSrv, "SRV"},
		{types.RRTypePtr, "PTR"},
	}

	for _, tt := range tests {
		t.Run(string(tt.recordType), func(t *testing.T) {
			got := string(tt.recordType)
			if got != tt.want {
				t.Errorf("recordType = %q, want %q", got, tt.want)
			}
		})
	}
}

