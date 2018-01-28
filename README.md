# terraform-provider-mongodbatlas

Terraform MongoDB Atlas provider. Groups and clusters must be deleted manually.

## example

```
provider "mongodbatlas" {
  username = "atlas@mongodb.com"
  api_key = "ATLAS_API_KEY"
  org_id = "ATLAS_ORG_ID"
}

resource "mongodbatlas_group" "mongodb_group" {
  name = "pritunl-group"
}

resource "mongodbatlas_cluster" "mongodb_cluster" {
  group_id = "${mongodbatlas_group.mongodb_group.id}"
  name = "pritunl"
  region = "us-west-2"
  size = "m10"
  disk_size_gb = 10
}

resource "mongodbatlas_user" "mongodb_user" {
  group_id = "${mongodbatlas_group.mongodb_group.id}"
  name = "${mongodbatlas_cluster.mongodb_cluster.name}"
  cluster_name = "${mongodbatlas_cluster.mongodb_cluster.name}"
  database_name = "${mongodbatlas_cluster.mongodb_cluster.name}"
  password = "4799fd096f77409da554b2e0a13ed345"
}

resource "mongodbatlas_peer" "mongodb_peer" {
  group_id = "${mongodbatlas_group.mongodb_group.id}"
  container_id = "${mongodbatlas_cluster.mongodb_cluster.container_id}"
  aws_account_id = "AWS_ACCOUNT_ID"
  vpc_id = "vpc-ce4865a9"
  vpc_cidr = "10.150.0.0/16"
}
```
