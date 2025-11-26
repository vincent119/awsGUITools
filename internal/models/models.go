package models

// TagMap represents AWS resource tags.
type TagMap map[string]string

// EC2Instance describes essential fields for list/detail display.
type EC2Instance struct {
	ID               string
	Name             string
	State            string
	InstanceType     string
	AvailabilityZone string
	PrivateIP        string
	PublicIP         string
	VpcID            string
	SubnetID         string
	SecurityGroups   []string
	IAMRole          string
	Volumes          []EBSVolume
	Tags             TagMap
}

// EBSVolume describes the main EBS attachment fields.
type EBSVolume struct {
	ID         string
	DeviceName string
	SizeGiB    int32
	State      string
}

// RDSInstance describes core metadata for DB instances.
type RDSInstance struct {
	ID             string
	Engine         string
	EngineVersion  string
	MultiAZ        bool
	Endpoint       string
	SubnetGroup    string
	ParameterGroup []string
	SecurityGroups []string
	Tags           TagMap
}

// S3Bucket describes bucket level configuration.
type S3Bucket struct {
	Name       string
	Region     string
	Versioning string
	Encryption string
	Policy     string
	Lifecycle  string
	Tags       TagMap
}

// LambdaFunction describes AWS Lambda metadata.
type LambdaFunction struct {
	Name         string
	ARN          string
	Runtime      string
	MemoryMB     int32
	TimeoutSec   int32
	Role         string
	EnvVars      map[string]string
	Triggers     []string
	Tags         TagMap
	LastModified string
}

// ListItem aggregates cross-resource info for list UI.
type ListItem struct {
	ID       string
	Name     string
	Type     string
	Status   string
	Region   string
	Account  string
	Tags     TagMap
	Metadata map[string]string
}

// DetailView groups detailed info for UI tabs.
type DetailView struct {
	Overview  map[string]string
	Relations map[string][]string
	Tags      TagMap
}
