
// create ssm parameters for each member of `consul_acls` above
// name will be in format <consul_acl_definition_prefix>/<slug>
// unless `slug` is not defined, then it will use
// <consul_acl_definition_prefix>/<Name>
// all spaces will be removed
resource "aws_ssm_parameter" "consul_acls" {
  count = "${length(var.consul_acls)}"
  type  = "String"
  name  = "${var.consul_acl_definition_prefix}/${replace(lookup(var.consul_acls[count.index], "slug", lookup(var.consul_acls[count.index], "Name")), " ", "")}"
  value = "${jsonencode(var.consul_acls[count.index])}"
}

locals {
  // Consul configuration applied to agent
  consul_config = {
    acl_datacenter     = "${var.consul_datacenter}"
    acl_default_policy = "deny"
    acl_down_policy    = "extend-cache"
  }
}

resource "docker_image" "consul" {
  name = "consul:latest"
}

resource "docker_container" "consul" {
  image = "${docker_image.consul.latest}"
  name  = "consul-acl-example"

  env = [
    "CONSUL_LOCAL_CONFIG=${jsonencode(local.consul_config)}",
  ]

  ports {
    internal = 8500
    external = 8500
  }

  depends_on = ["aws_ssm_parameter.consul_acls"]

  // bootstrap ACL system
  provisioner "local-exec" {
    command = "sleep 5 && consulssm bootstrap --overwrite || true"

    environment {
      SSM_CONSUL_TOKEN_PARAM = "${var.consul_bootstrap_token_param}"
    }
  }

  // sync ACLs from SSM
  provisioner "local-exec" {
    command = "consulssm sync --overwrite"

    environment {
      SSM_CONSUL_TOKEN_PARAM = "${var.consul_bootstrap_token_param}"
      SSM_DEFINITION_PREFIX  = "${var.consul_acl_definition_prefix}"
      SSM_ID_PREFIX          = "${var.consul_acl_id_prefix}"
    }
  }

  // install a token from SSM to local agent
  provisioner "local-exec" {
    command = "consulssm agent acl_agent_token ${var.consul_acl_id_prefix}/agent"

    environment {
      SSM_CONSUL_TOKEN_PARAM = "${var.consul_bootstrap_token_param}"
    }
  }
}
