# consulssm

## Usage
The availabe commands for consulssm are shown below:
```
$ consulssm --help
Bootstrap and manage Consul ACLs via AWS SSM parameters

Usage:
  consulssm [command]

Available Commands:
  agent       Update Consul agent ACL tokens via SSM parameters
  bootstrap   Bootstrap Consul ACLs and save token to an SSM parameter
  help        Help about any command
  sync        Synchronize Consul ACLs via SSM parameters

Flags:
      --debug   Enable debug logging
  -h, --help    help for consulssm
```

Every option can be set with an environment variable rather than command-line flags by
upper-casing the flag name, substituting `-` with `_`, and prefixing the name with `SSM_`.
For example, the flag `--consul-token-param` can be set with the environment variable
`SSM_CONSUL_TOKEN_PARAM`.

### Bootstrap Command
```
Bootstrap Consul ACLs and save token to an SSM parameter

Usage:
  consulssm bootstrap [flags]

Flags:
  -m, --consul-token-param string   SSM parameter name to write Consul bootstrap token ID
  -h, --help                        help for bootstrap
      --hide                        Hide bootstrap token from standard output
  -I, --insecure                    Skip encryption when writing token to SSM
  -k, --kms-key-id string           Optional KMS key ID for encrypting bootstrap token ID
  -o, --overwrite                   Overwrite existing SSM parameter value if it exists

Global Flags:
      --debug   Enable debug logging
```

### Sync Command
```
Synchronize Consul ACLs via SSM parameters

Usage:
  consulssm sync [flags]

Flags:
  -m, --consul-token-param string   SSM parameter name for Consul management token
  -d, --definition-prefix string    SSM heirarchy prefix to read ACL definitions (required)
  -h, --help                        help for sync
  -i, --id-prefix string            SSM heirarchy prefix to read/write ACL token IDs
  -I, --insecure                    Skip encryption when updating SSM with new token IDs
  -k, --kms-key-id string           Optional KMS key ID for encrypting new token IDs
  -l, --leader                      Manage ACLs only if Consul agent is current leader
  -o, --overwrite                   Overwrite existing SSM parameter values if they exist
  -p, --page-size int               Maximum results per SSM query
  -r, --recurring int               Make recurring and wait given number of seconds between syncs

Global Flags:
      --debug   Enable debug logging
```

### Agent Commands
```
Update Consul agent ACL tokens via SSM parameters

Usage:
  consulssm agent [command]

Available Commands:
  acl_agent_master_token Set Consul agent acl_agent_master_token
  acl_agent_token        Set Consul agent acl_agent_token
  acl_replication_token  Set Consul agent acl_replication_token
  acl_token              Set Consul agent acl_token

Flags:
  -m, --consul-token-param string   SSM parameter name for Consul management token
  -h, --help                        help for agent

Global Flags:
      --debug   Enable debug logging
```

Each agent ACL command has the same arguments and options, for example the
`agent acl_token` sub-command is shown below:

```
Set Consul agent acl_token

Usage:
  consulssm agent acl_token TOKEN [flags]

Flags:
  -h, --help   help for acl_token

Global Flags:
  -m, --consul-token-param string   SSM parameter name for Consul management token
      --debug                       Enable debug logging
```
