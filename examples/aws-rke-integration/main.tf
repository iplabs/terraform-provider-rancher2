provider "local" {
  version = "~> 1.1"
}

provider "tls" {
  version = "~> 1.1"
}

provider "aws" {
  alias   = "my-aws-account"
  version = "~> 1.41"
  region  = "eu-central-1"
}

provider "kubernetes" {
  version                = "~> 1.1"
  host                   = "${rke_cluster.cluster.api_server_url}"
  client_certificate     = "${rke_cluster.cluster.client_cert}"
  client_key             = "${rke_cluster.cluster.client_key}"
  cluster_ca_certificate = "${rke_cluster.cluster.ca_crt}"
}

# The RKE provider plugin can be fetched from
# https://github.com/yamamoto-febc/terraform-provider-rke
provider "rke" {
  version = "~> 0.3"
}

provider rancher2 {}

module "aws_infrastructure" {
  source = "./modules/aws_infrastructure"

  instance_type = "t2.micro"
  cluster_id    = "example-aws-rke-integration"
}

resource "rancher2_cluster" "rke_cluster" {
  name        = "example-aws-rke-integration"
  description = "Rancher Kubernetes Engine (RKE) cluster on AWS EC2 instances"
}

# The registration token is necessary if we want to import an existing cluster
# into Rancher. It will provide a Kubernetes manifest URL that can be applied
# onto the existing cluster and that will trigger the start of the rancher-agent
# deployment all further necessary steps to make Rancher our cluster manager.
resource "rancher2_cluster_registration_token" "registration_token" {
  cluster_id = "${rancher2_cluster.rke_cluster.id}"
}

resource "rancher2_project" "my_project" {
  cluster_id  = "${rancher2_cluster.rke_cluster.id}"
  name        = "My Project"
  description = "My example project"
}

resource rke_cluster "cluster" {
  cloud_provider {
    name = "aws"
  }

  addons_include = [
    # Automatically imports the cluster into Rancher.
    "${rancher2_cluster_registration_token.registration_token.manifest_url}",
  ]

  nodes = [
    {
      address = "${module.aws_infrastructure.addresses[0]}"
      user    = "${module.aws_infrastructure.ssh_username}"
      ssh_key = "${module.aws_infrastructure.private_key}"
      role    = ["controlplane", "etcd"]
    },
    {
      address = "${module.aws_infrastructure.addresses[1]}"
      user    = "${module.aws_infrastructure.ssh_username}"
      ssh_key = "${module.aws_infrastructure.private_key}"
      role    = ["worker", "etcd"]
    },
    {
      address = "${module.aws_infrastructure.addresses[2]}"
      user    = "${module.aws_infrastructure.ssh_username}"
      ssh_key = "${module.aws_infrastructure.private_key}"
      role    = ["worker", "etcd"]
    },
    {
      address = "${module.aws_infrastructure.addresses[3]}"
      user    = "${module.aws_infrastructure.ssh_username}"
      ssh_key = "${module.aws_infrastructure.private_key}"
      role    = ["worker"]
    },
  ]
}

resource kubernetes_namespace "my_namespace" {
  metadata {
    name = "my-namespace"

    labels {
      "example-label1" = "foo"
      "example-label2" = "bar"
    }

    annotations {
      "com.example/annotation1" = "baz"

      # Rancher's namespace<>project associations are handled by annotations
      # on the namespace that should be part of a project. We want to include
      # the namespace in our previously create rancher project and will therefore
      # set "field.cattle.io/projectId" accordingly.
      "field.cattle.io/projectId" = "${rancher2_project.my_project.id}"
    }
  }

  lifecycle {
    ignore_changes = [
      "metadata.0.annotations.cattle.io/status",
      "metadata.0.annotations.lifecycle.cattle.io/create.namespace-auth",
    ]
  }
}
