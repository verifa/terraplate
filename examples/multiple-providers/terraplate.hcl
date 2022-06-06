# In this example we generate the Kubernetes provider blocks to work with
# multiple clusters

template "kubernetes-providers" {
  contents = <<EOL
  {{ range $cluster, $value := .Values.eks_clusters }}
  data "aws_eks_cluster" "{{ $cluster }}" {
    name     = "{{ $cluster }}"
  }

  data "aws_eks_cluster_auth" "{{ $cluster }}" {
    name     = "{{ $cluster }}"
  }

  provider "kubernetes" {
    alias                  = "{{ $cluster }}"
    host                   = data.aws_eks_cluster.{{ $cluster }}.endpoint
    token                  = data.aws_eks_cluster_auth.{{ $cluster }}.token
    cluster_ca_certificate = base64decode(data.aws_eks_cluster.{{ $cluster }}.certificate_authority.0.data)
  }
  {{ end }}
  EOL
}

values {
  eks_clusters = {
    my_eks_cluster = {
      some_config = "whatever"
    }
  }
}
