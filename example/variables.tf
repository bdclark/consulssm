variable "consul_acl_definition_prefix" {
  description = "The SSM heirarchy prefix where Consul ACL data should be set"
  default     = "/example/consul/acl/definitions"
}

variable "consul_acl_id_prefix" {
  description = "The SSM heirarchy prefix to get Consul ACL IDs (used as data input in example)"
  default     = "/example/consul/acl/ids"
}

variable "consul_bootstrap_token_param" {
  description = "The SSM parameter name where the boostrap token will be written during bootstrapping"
  default     = "/example/consul/acl/master_token"
}

variable "consul_datacenter" {
  description = "The Consul datacenter"
  default     = "dc1"
}

variable "consul_acls" {
  description = "List of Consul ACL definitions"
  default = [
    {
      slug = "agent"
      Name = "Agent Token"
      Type = "client"

      Rules = <<EOF
node "" { policy = "write" }
service "" { policy = "read" }
key "_rexec/" { policy = "write" }
EOF
    },
    {
      ID   = "anonymous"
      Name = "Anonymous Token"

      Rules = <<EOF
node "" { policy = "read" }
service "" { policy = "read" }
EOF
    },
    {
      ID      = "OldDeprecatedToken"
      Name    = "Some ACL That Should Not Exist"
      Destroy = "true"
    },
    {
      ID   = "supersecret"
      Name = "Some Management Token"
      Type = "management"
    },
  ]
}
