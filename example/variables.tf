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
