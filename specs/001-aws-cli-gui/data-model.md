# Data Model: AWS CLI GUI

## Entities

### EC2Instance
- id（InstanceId）、name、state、instance_type、az  
- private_ip、public_ip、vpc_id、subnet_id  
- sg_ids[]、iam_role、ebs_volume_ids[]

### RDSInstance
- id（DBInstanceIdentifier）、engine、engine_version、multi_az  
- endpoint、subnet_group、parameter_group、sg_ids[]

### S3Bucket
- name、region、versioning、encryption  
- bucket_policy(optional)、lifecycle(optional)

### LambdaFunction
- name、arn、runtime、memory、timeout、role  
- env_vars、triggers(optional)

## Relations

- EC2Instance — SecurityGroup（N:N）  
- EC2Instance — IAMRole（1:0..1）  
- EC2Instance — EBSVolume（1:N）  
- RDSInstance — SubnetGroup/ParameterGroup/SecurityGroup  
- LambdaFunction — IAMRole（1:0..1）  
- S3Bucket — Policy/Lifecycle（0..1）

## View Models（UI 聚合）

- ListItem：id/name/type/tags/primary_state/primary_metrics  
- DetailView：基本屬性 + 關聯（陣列） + 快速操作入口  
- MetricsView：關鍵 KPI（依資源類型）  
- LogsView：最近 N 筆 log events（可分頁）


