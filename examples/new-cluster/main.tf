provider rancher2 {
}

resource "rancher2_cluster" "new_cluster" {
  name        = "example-new-cluster"
  description = "Empty cluster that has not (yet) been linked to a *real* kubernetes cluster"
}

resource "rancher2_cluster_registration_token" "registration_token" {
  cluster_id = "${rancher2_cluster.new_cluster.cluster_id}"
}