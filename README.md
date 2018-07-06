# consulssm
A utility to bootstrap and manage Consul ACLs via AWS SSM parameters.

## Overview
The goal of consulssm is to simplify and streamline the process of managing
Consul ACLs. It allows you to store ACL definitions and their respective token
IDs securely out-of-band from Consul in AWS SSM, and simplifies bootstrapping
new clusters in highly dynamic, multiple environment scenarios.

More documentation is coming, but here's a couple nuggets to get you started:

The following command would bootstrap Consul ACLs and store the resulting
token as an SSM parameter:
```bash
export PREFIX="/dev/consul/acl" # set a prefix for brevity in later examples
consulssm bootstrap --consul-token-param "${PREFIX}/master_token"
```
If bootstrapping is successful, the resulting token is stored in
`${PREFIX}/master_token` (in this case `/dev/consul/acl/master_token`). The
token can also optionally be captured in standard output.

Then, to create/sync ACLs, the sync command requires JSON-encoded
ACL definitions to be stored in SSM. The format is the same as the payload used
with the Consul HTTP API (see https://www.consul.io/api/acl.html#parameters).

For example, assuming the following parameter was
written to SSM:

```bash
aws ssm put-parameter --name "${PREFIX}/definitions/agent" --type String --value '{
  "Name":"Agent Token",
  "Rules":"node \"\" { policy = \"write\" }\nservice \"\" { policy = \"read\" }\nkey \"_rexec/\" { policy = \"write\" }\n",
  "Type":"client"
}'
```

Then...

```bash
consulssm sync \
  --consul-token-param ${PREFIX}/master_token \
  --definition-prefix ${PREFIX}/definitions \
  --id-prefix ${PREFIX}/ids
```

The above command would write new token IDs for everything defined in `${PREFIX}/definitions/*`
to `${PREFIX}/ids/*` as long as the host running consulssm has _read_ privileges to `${PREFIX}/definitions/*`
and _write_ permissions to `${PREFIX}/ids/*`. It would also update any ACLs whose
definition does not match SSM.

In this case it would read `${PREFIX}/definitions/agent`,
create an ACL with the name `Agent Token`, then write the token ID to `${PREFIX}/ids/agent`.

Since the management token is being read from SSM as well, the host would also
need _read_ access to `${PREFIX}/master_token`. However, a token could also
be supplied via the `CONSUL_HTTP_TOKEN` environment variable.

Finally, the command below could install an `acl_agent_token` to the local
agent where the ACL was described in the parameter defined earlier, then
created with the sync command above.
```
consulssm agent acl_agent_token ${PREFIX}/ids/agent \
  --consul-token-param ${PREFIX}/master_token
```
In this case, the host would need access to read the management token used to
update the agent's ACLs.  This requires the token used to have `agent:write`
permissions, so it may not work for your use-case.

## Commands
- [bootstrap](#bootstrap-command) - Bootstrap Consul ACLs and save token to an SSM parameter
- [sync](#sync-command) - Synchronize Consul ACLs via SSM parameters
- [agent](#agent-commands) - Update Consul agent ACL tokens via SSM parameters

## Environment Variables and Flags
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
