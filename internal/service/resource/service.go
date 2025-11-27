package resource

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/vin/ck123gogo/internal/app/state"
	"github.com/vin/ck123gogo/internal/aws/clients"
	"github.com/vin/ck123gogo/internal/aws/logs"
	"github.com/vin/ck123gogo/internal/aws/metrics"
	"github.com/vin/ck123gogo/internal/aws/repo"
	"github.com/vin/ck123gogo/internal/models"
	"github.com/vin/ck123gogo/internal/observability"
	"github.com/vin/ck123gogo/internal/search"
)

// Kind 表示資源類型。
type Kind string

const (
	KindEC2            Kind = "ec2"
	KindRDS            Kind = "rds"
	KindS3             Kind = "s3"
	KindLambda         Kind = "lambda"
	KindRoute53        Kind = "route53"
	KindRoute53Records Kind = "route53-records"
	KindS3Objects      Kind = "s3-objects"
)

// Service 封裝資源查詢與轉換邏輯，供 UI 直接使用。
type Service struct {
	factory *clients.Factory
	metrics *observability.AWSCallMetrics
	timeout time.Duration
	state   *state.Store

	ec2Repo      *repo.EC2Repository
	rdsRepo      *repo.RDSRepository
	s3Repo       *repo.S3Repository
	lambdaRepo   *repo.LambdaRepository
	route53Repo  *repo.Route53Repository
	metricFetch  metrics.MetricAPI
	logFetch     *logs.Fetcher

	mu    sync.RWMutex
	cache map[Kind]map[string]models.DetailView

	// S3 瀏覽狀態
	currentBucket string
	currentPrefix string

	// Route53 瀏覽狀態
	currentZoneID   string
	currentZoneName string
}

// NewService 建立資源服務。
func NewService(factory *clients.Factory, metrics *observability.AWSCallMetrics, timeout time.Duration, st *state.Store) *Service {
	if timeout == 0 {
		timeout = 15 * time.Second
	}
	return &Service{
		factory:     factory,
		metrics:     metrics,
		timeout:     timeout,
		state:       st,
		ec2Repo:     repo.NewEC2Repository(),
		rdsRepo:     repo.NewRDSRepository(),
		s3Repo:      repo.NewS3Repository(),
		lambdaRepo:  repo.NewLambdaRepository(),
		route53Repo: repo.NewRoute53Repository(),
		cache:       make(map[Kind]map[string]models.DetailView),
	}
}

// ListItems 依資源類型列出清單（會套用搜尋條件並更新 Detail 快取）。
func (s *Service) ListItems(ctx context.Context, kind Kind, matcher search.Matcher) ([]models.ListItem, error) {
	if s.factory == nil {
		return nil, errors.New("aws client factory is nil")
	}
	if s.state == nil {
		return nil, errors.New("state store is nil")
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	profile := s.state.Profile()
	region := s.state.Region()

	var (
		items   []models.ListItem
		details map[string]models.DetailView
		err     error
	)

	switch kind {
	case KindEC2:
		client, errClient := s.factory.EC2(ctx, profile, region)
		if errClient != nil {
			return nil, errClient
		}
		var instances []models.EC2Instance
		start := time.Now()
		instances, err = s.ec2Repo.ListInstances(ctx, client, nil)
		s.observe(ctx, "ec2", "DescribeInstances", start, err)
		if err == nil {
			items, details = buildEC2List(instances, matcher)
		}
	case KindRDS:
		client, errClient := s.factory.RDS(ctx, profile, region)
		if errClient != nil {
			return nil, errClient
		}
		var dbs []models.RDSInstance
		start := time.Now()
		dbs, err = s.rdsRepo.ListInstances(ctx, client, nil)
		s.observe(ctx, "rds", "DescribeDBInstances", start, err)
		if err == nil {
			items, details = buildRDSList(dbs, matcher)
		}
	case KindS3:
		client, errClient := s.factory.S3(ctx, profile, region)
		if errClient != nil {
			return nil, errClient
		}
		var buckets []models.S3Bucket
		start := time.Now()
		buckets, err = s.s3Repo.ListBuckets(ctx, client)
		s.observe(ctx, "s3", "ListBuckets", start, err)
		if err == nil {
			items, details = buildS3List(buckets, matcher)
		}
	case KindLambda:
		client, errClient := s.factory.Lambda(ctx, profile, region)
		if errClient != nil {
			return nil, errClient
		}
		var fns []models.LambdaFunction
		start := time.Now()
		fns, err = s.lambdaRepo.ListFunctions(ctx, client, nil)
		s.observe(ctx, "lambda", "ListFunctions", start, err)
		if err == nil {
			items, details = buildLambdaList(fns, matcher)
		}
	case KindRoute53:
		client, errClient := s.factory.Route53(ctx, profile, region)
		if errClient != nil {
			return nil, errClient
		}
		var zones []models.Route53HostedZone
		start := time.Now()
		zones, err = s.route53Repo.ListHostedZones(ctx, client)
		s.observe(ctx, "route53", "ListHostedZones", start, err)
		if err == nil {
			items, details = buildRoute53ZoneList(zones, matcher)
		}
	case KindRoute53Records:
		if s.currentZoneID == "" {
			return nil, fmt.Errorf("no hosted zone selected")
		}
		client, errClient := s.factory.Route53(ctx, profile, region)
		if errClient != nil {
			return nil, errClient
		}
		var records []models.Route53Record
		start := time.Now()
		records, err = s.route53Repo.ListRecords(ctx, client, s.currentZoneID)
		s.observe(ctx, "route53", "ListResourceRecordSets", start, err)
		if err == nil {
			items, details = buildRoute53RecordList(records, s.currentZoneName, matcher)
		}
	case KindS3Objects:
		if s.currentBucket == "" {
			return nil, fmt.Errorf("no bucket selected")
		}
		client, errClient := s.factory.S3(ctx, profile, region)
		if errClient != nil {
			return nil, errClient
		}
		var objects []models.S3Object
		start := time.Now()
		objects, err = s.s3Repo.ListObjects(ctx, client, s.currentBucket, s.currentPrefix)
		s.observe(ctx, "s3", "ListObjectsV2", start, err)
		if err == nil {
			items, details = buildS3ObjectList(objects, s.currentBucket, s.currentPrefix, matcher)
		}
	default:
		return nil, fmt.Errorf("unknown resource kind: %s", kind)
	}

	if err != nil {
		return nil, err
	}

	s.storeDetails(kind, details)
	return items, nil
}

// Detail 取得指定資源的詳細資訊；若快取不存在會重新查詢。
func (s *Service) Detail(ctx context.Context, kind Kind, id string) (models.DetailView, error) {
	if detail, ok := s.getDetail(kind, id); ok {
		return detail, nil
	}
	_, err := s.ListItems(ctx, kind, search.NewMatcher(""))
	if err != nil {
		return models.DetailView{}, err
	}
	if detail, ok := s.getDetail(kind, id); ok {
		return detail, nil
	}
	return models.DetailView{}, fmt.Errorf("resource %s not found", id)
}

func (s *Service) observe(ctx context.Context, service, operation string, start time.Time, err error) {
	if s.metrics == nil {
		return
	}
	s.metrics.Observe(ctx, service, operation, time.Since(start), err)
}

// GetMetrics 取得指定資源的 CloudWatch 指標，含 context timeout 與錯誤處理。
func (s *Service) GetMetrics(ctx context.Context, kind Kind, resourceID string, start, end time.Time) (map[string]metrics.Series, error) {
	if s.factory == nil {
		return nil, errors.New("aws client factory is nil")
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	profile := s.state.Profile()
	region := s.state.Region()

	client, err := s.factory.CloudWatch(ctx, profile, region)
	if err != nil {
		return nil, fmt.Errorf("create cloudwatch client: %w", err)
	}

	queries := metrics.DefaultQueries(metrics.ResourceKind(kind), resourceID)
	if len(queries) == 0 {
		return nil, nil
	}

	fetcher := metrics.NewFetcher(client)
	startT := time.Now()
	result, err := fetcher.Fetch(ctx, metrics.Options{
		StartTime: start,
		EndTime:   end,
		Period:    60,
		Queries:   queries,
	})
	s.observe(ctx, "cloudwatch", "GetMetricData", startT, err)
	if err != nil {
		return nil, fmt.Errorf("fetch metrics: %w", err)
	}
	return result, nil
}

// GetLogs 取得指定資源的 CloudWatch Logs，含 context timeout 與錯誤處理。
func (s *Service) GetLogs(ctx context.Context, kind Kind, resourceID string, start, end time.Time, limit int32) (logs.Page, error) {
	if s.factory == nil {
		return logs.Page{}, errors.New("aws client factory is nil")
	}

	logGroup := logs.DeriveLogGroup(logs.ResourceKind(kind), resourceID)
	if logGroup == "" {
		return logs.Page{}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	profile := s.state.Profile()
	region := s.state.Region()

	client, err := s.factory.CloudWatchLogs(ctx, profile, region)
	if err != nil {
		return logs.Page{}, fmt.Errorf("create cloudwatchlogs client: %w", err)
	}

	fetcher := logs.NewFetcher(client)
	startT := time.Now()
	result, err := fetcher.Filter(ctx, logs.Options{
		LogGroup:  logGroup,
		StartTime: start.UnixMilli(),
		EndTime:   end.UnixMilli(),
		Limit:     limit,
	})
	s.observe(ctx, "cloudwatchlogs", "FilterLogEvents", startT, err)
	if err != nil {
		return logs.Page{}, fmt.Errorf("filter logs: %w", err)
	}
	return result, nil
}

func (s *Service) storeDetails(kind Kind, details map[string]models.DetailView) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[kind] = details
}

func (s *Service) getDetail(kind Kind, id string) (models.DetailView, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if detailMap, ok := s.cache[kind]; ok {
		detail, exists := detailMap[id]
		return detail, exists
	}
	return models.DetailView{}, false
}

func buildEC2List(instances []models.EC2Instance, matcher search.Matcher) ([]models.ListItem, map[string]models.DetailView) {
	items := make([]models.ListItem, 0, len(instances))
	details := make(map[string]models.DetailView, len(instances))

	for _, inst := range instances {
		if !matcher.Match(inst.Name + inst.ID) {
			continue
		}
		items = append(items, models.ListItem{
			ID:     inst.ID,
			Name:   fallback(inst.Name, inst.ID),
			Type:   "EC2",
			Status: inst.State,
			Region: inst.AvailabilityZone,
			Tags:   inst.Tags,
			Metadata: map[string]string{
				"type": inst.InstanceType,
			},
		})
		details[inst.ID] = models.DetailView{
			Overview: map[string]string{
				"Instance ID": inst.ID,
				"Name":        inst.Name,
				"Type":        inst.InstanceType,
				"State":       inst.State,
				"Private IP":  inst.PrivateIP,
				"Public IP":   inst.PublicIP,
				"VPC":         inst.VpcID,
				"Subnet":      inst.SubnetID,
				"IAM Role":    inst.IAMRole,
			},
			Relations: map[string][]string{
				"Security Groups": inst.SecurityGroups,
				"EBS Volumes":     volumeIDs(inst.Volumes),
			},
			Tags: inst.Tags,
		}
	}
	return items, details
}

func buildRDSList(instances []models.RDSInstance, matcher search.Matcher) ([]models.ListItem, map[string]models.DetailView) {
	items := make([]models.ListItem, 0, len(instances))
	details := make(map[string]models.DetailView, len(instances))

	for _, inst := range instances {
		if !matcher.Match(inst.ID + inst.Engine) {
			continue
		}
		items = append(items, models.ListItem{
			ID:     inst.ID,
			Name:   inst.ID,
			Type:   "RDS",
			Status: inst.Engine,
			Tags:   inst.Tags,
			Metadata: map[string]string{
				"endpoint": inst.Endpoint,
			},
		})
		details[inst.ID] = models.DetailView{
			Overview: map[string]string{
				"DB Instance": inst.ID,
				"Engine":      inst.Engine,
				"Version":     inst.EngineVersion,
				"Multi-AZ":    fmt.Sprintf("%t", inst.MultiAZ),
				"Endpoint":    inst.Endpoint,
				"SubnetGroup": inst.SubnetGroup,
			},
			Relations: map[string][]string{
				"Parameter Groups": inst.ParameterGroup,
				"Security Groups":  inst.SecurityGroups,
			},
			Tags: inst.Tags,
		}
	}
	return items, details
}

func buildS3List(buckets []models.S3Bucket, matcher search.Matcher) ([]models.ListItem, map[string]models.DetailView) {
	items := make([]models.ListItem, 0, len(buckets))
	details := make(map[string]models.DetailView, len(buckets))

	for _, bucket := range buckets {
		if !matcher.Match(bucket.Name) {
			continue
		}
		items = append(items, models.ListItem{
			ID:     bucket.Name,
			Name:   bucket.Name,
			Type:   "S3",
			Status: bucket.Versioning,
			Region: bucket.Region,
			Tags:   bucket.Tags,
		})
		details[bucket.Name] = models.DetailView{
			Overview: map[string]string{
				"Bucket":     bucket.Name,
				"Region":     bucket.Region,
				"Versioning": bucket.Versioning,
				"Encryption": bucket.Encryption,
			},
			Relations: map[string][]string{
				"Policies":  filterEmpty(bucket.Policy),
				"Lifecycle": filterEmpty(bucket.Lifecycle),
			},
			Tags: bucket.Tags,
		}
	}
	return items, details
}

func buildLambdaList(fns []models.LambdaFunction, matcher search.Matcher) ([]models.ListItem, map[string]models.DetailView) {
	items := make([]models.ListItem, 0, len(fns))
	details := make(map[string]models.DetailView, len(fns))

	for _, fn := range fns {
		if !matcher.Match(fn.Name) {
			continue
		}
		items = append(items, models.ListItem{
			ID:     fn.ARN,
			Name:   fn.Name,
			Type:   "Lambda",
			Status: fn.Runtime,
			Tags:   fn.Tags,
			Metadata: map[string]string{
				"memory": fmt.Sprintf("%d MB", fn.MemoryMB),
			},
		})
		details[fn.ARN] = models.DetailView{
			Overview: map[string]string{
				"Function":   fn.Name,
				"Runtime":    fn.Runtime,
				"Memory":     fmt.Sprintf("%d MB", fn.MemoryMB),
				"Timeout":    fmt.Sprintf("%d s", fn.TimeoutSec),
				"Role":       fn.Role,
				"LastChange": fn.LastModified,
			},
			Relations: map[string][]string{
				"Environment": flattenEnv(fn.EnvVars),
				"Triggers":    fn.Triggers,
			},
			Tags: fn.Tags,
		}
	}
	return items, details
}

func volumeIDs(vols []models.EBSVolume) []string {
	result := make([]string, 0, len(vols))
	for _, vol := range vols {
		if vol.ID != "" {
			result = append(result, fmt.Sprintf("%s (%s)", vol.ID, vol.State))
		}
	}
	return result
}

func filterEmpty(value string) []string {
	if value == "" {
		return nil
	}
	return []string{value}
}

func flattenEnv(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make([]string, 0, len(env))
	for _, k := range keys {
		result = append(result, fmt.Sprintf("%s=%s", k, env[k]))
	}
	return result
}

func fallback(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// SetCurrentBucket 設定目前瀏覽的 S3 bucket。
func (s *Service) SetCurrentBucket(bucket string) {
	s.currentBucket = bucket
	s.currentPrefix = ""
}

// SetCurrentPrefix 設定目前瀏覽的 S3 prefix（目錄）。
func (s *Service) SetCurrentPrefix(prefix string) {
	s.currentPrefix = prefix
}

// CurrentBucket 回傳目前瀏覽的 bucket 名稱。
func (s *Service) CurrentBucket() string {
	return s.currentBucket
}

// CurrentPrefix 回傳目前瀏覽的 prefix。
func (s *Service) CurrentPrefix() string {
	return s.currentPrefix
}

// NavigateUp 返回上一層目錄，回傳是否還在 bucket 內。
func (s *Service) NavigateUp() bool {
	if s.currentPrefix == "" {
		// 已在根目錄，清除 bucket 回到 bucket 列表
		s.currentBucket = ""
		return false
	}
	// 移除最後一個目錄
	prefix := strings.TrimSuffix(s.currentPrefix, "/")
	lastSlash := strings.LastIndex(prefix, "/")
	if lastSlash < 0 {
		s.currentPrefix = ""
	} else {
		s.currentPrefix = prefix[:lastSlash+1]
	}
	return true
}

// SetCurrentZone 設定目前瀏覽的 Route53 Hosted Zone。
func (s *Service) SetCurrentZone(zoneID, zoneName string) {
	s.currentZoneID = zoneID
	s.currentZoneName = zoneName
}

// CurrentZoneID 回傳目前瀏覽的 Zone ID。
func (s *Service) CurrentZoneID() string {
	return s.currentZoneID
}

// CurrentZoneName 回傳目前瀏覽的 Zone 名稱。
func (s *Service) CurrentZoneName() string {
	return s.currentZoneName
}

// ClearCurrentZone 清除目前的 Zone，回到 Zone 列表。
func (s *Service) ClearCurrentZone() {
	s.currentZoneID = ""
	s.currentZoneName = ""
}

func buildRoute53ZoneList(zones []models.Route53HostedZone, matcher search.Matcher) ([]models.ListItem, map[string]models.DetailView) {
	items := make([]models.ListItem, 0, len(zones))
	details := make(map[string]models.DetailView, len(zones))

	for _, zone := range zones {
		if !matcher.Match(zone.Name + zone.ID) {
			continue
		}
		zoneType := "Public"
		if zone.IsPrivate {
			zoneType = "Private"
		}
		items = append(items, models.ListItem{
			ID:     zone.ID,
			Name:   zone.Name,
			Type:   "Route53",
			Status: zoneType,
			Metadata: map[string]string{
				"records": fmt.Sprintf("%d", zone.RecordCount),
			},
		})
		details[zone.ID] = models.DetailView{
			Overview: map[string]string{
				"Zone ID":      zone.ID,
				"Name":         zone.Name,
				"Type":         zoneType,
				"Record Count": fmt.Sprintf("%d", zone.RecordCount),
				"Comment":      zone.Comment,
			},
		}
	}
	return items, details
}

func buildRoute53RecordList(records []models.Route53Record, zoneName string, matcher search.Matcher) ([]models.ListItem, map[string]models.DetailView) {
	items := make([]models.ListItem, 0, len(records))
	details := make(map[string]models.DetailView, len(records))

	for _, record := range records {
		if !matcher.Match(record.Name + record.Type) {
			continue
		}
		id := record.Name + "_" + record.Type
		value := ""
		if record.AliasTarget != "" {
			value = "ALIAS → " + record.AliasTarget
		} else if len(record.Values) > 0 {
			value = record.Values[0]
			if len(record.Values) > 1 {
				value += fmt.Sprintf(" (+%d)", len(record.Values)-1)
			}
		}
		items = append(items, models.ListItem{
			ID:     id,
			Name:   record.Name,
			Type:   record.Type,
			Status: value,
			Metadata: map[string]string{
				"ttl": fmt.Sprintf("%d", record.TTL),
			},
		})
		details[id] = models.DetailView{
			Overview: map[string]string{
				"Name":   record.Name,
				"Type":   record.Type,
				"TTL":    fmt.Sprintf("%d", record.TTL),
				"Zone":   zoneName,
			},
			Relations: map[string][]string{
				"Values": record.Values,
			},
		}
		if record.AliasTarget != "" {
			details[id].Overview["Alias"] = record.AliasTarget
		}
	}
	return items, details
}

func buildS3ObjectList(objects []models.S3Object, bucket, prefix string, matcher search.Matcher) ([]models.ListItem, map[string]models.DetailView) {
	items := make([]models.ListItem, 0, len(objects))
	details := make(map[string]models.DetailView, len(objects))

	for _, obj := range objects {
		// 取得顯示名稱（移除 prefix）
		displayName := strings.TrimPrefix(obj.Key, prefix)
		if obj.IsDirectory {
			displayName = strings.TrimSuffix(displayName, "/") + "/"
		}

		if !matcher.Match(displayName) {
			continue
		}

		objType := "File"
		status := formatSize(obj.Size)
		if obj.IsDirectory {
			objType = "Dir"
			status = ""
		}

		items = append(items, models.ListItem{
			ID:     obj.Key,
			Name:   displayName,
			Type:   objType,
			Status: status,
			Region: obj.LastModified,
			Metadata: map[string]string{
				"storage": obj.StorageClass,
				"bucket":  bucket,
			},
		})

		details[obj.Key] = models.DetailView{
			Overview: map[string]string{
				"Key":           obj.Key,
				"Bucket":        bucket,
				"Size":          formatSize(obj.Size),
				"Last Modified": obj.LastModified,
				"Storage Class": obj.StorageClass,
			},
		}
	}
	return items, details
}

func formatSize(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
