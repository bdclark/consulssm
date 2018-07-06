package cmd

import (
	"github.com/bdclark/consulssm/acl"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	agentACLTokenName            = "acl_token"
	agentACLAgentTokenName       = "acl_agent_token"
	agentACLAgentMasterTokenName = "acl_agent_master_token"
	agentACLReplicationTokenName = "acl_replication_token"
	agentACLTypeKeyName          = "acl-type"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Update Consul agent ACL tokens via SSM parameters",
}

var agentACLTokenCmd = &cobra.Command{
	Use:   "acl_token TOKEN",
	Short: "Set Consul agent acl_token",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		// bind non-unique flags only when command is executed
		// https://github.com/spf13/viper/issues/233
		bindFlag(cmd, ConsulTokenParamFlagName)

		viper.Set(agentACLTypeKeyName, agentACLTokenName)
	},
	Run: agentACLRun,
}

var agentACLAgentTokenCmd = &cobra.Command{
	Use:   "acl_agent_token TOKEN",
	Short: "Set Consul agent acl_agent_token",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		// bind commonly-named flags only when command is executed
		bindFlag(cmd, ConsulTokenParamFlagName)
		viper.Set(agentACLTypeKeyName, agentACLAgentTokenName)
	},
	Run: agentACLRun,
}

var agentACLAgentMasterTokenCmd = &cobra.Command{
	Use:   "acl_agent_master_token TOKEN",
	Short: "Set Consul agent acl_agent_master_token",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		// bind commonly-named flags only when command is executed
		bindFlag(cmd, ConsulTokenParamFlagName)
		viper.Set(agentACLTypeKeyName, agentACLAgentMasterTokenName)
	},
	Run: agentACLRun,
}

var agentACLReplicationTokenCmd = &cobra.Command{
	Use:   "acl_replication_token TOKEN",
	Short: "Set Consul agent acl_replication_token",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		// bind commonly-named flags only when command is executed
		bindFlag(cmd, ConsulTokenParamFlagName)
		viper.Set(agentACLTypeKeyName, agentACLReplicationTokenName)
	},
	Run: agentACLRun,
}

func init() {
	agentCmd.AddCommand(agentACLTokenCmd)
	agentCmd.AddCommand(agentACLAgentTokenCmd)
	agentCmd.AddCommand(agentACLAgentMasterTokenCmd)
	agentCmd.AddCommand(agentACLReplicationTokenCmd)

	agentCmd.PersistentFlags().StringP(ConsulTokenParamFlagName, "m", "", "SSM parameter name for Consul management token")
}

func agentACLRun(cmd *cobra.Command, args []string) {
	c, err := acl.NewClientSet(&acl.ClientSetInput{
		ConsulTokenParam: viper.GetString(ConsulTokenParamFlagName),
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	tokenParam := args[0]
	token, err := c.GetStringParameter(tokenParam, true)
	if err != nil {
		log.Fatal(err.Error())
	}

	switch viper.GetString(agentACLTypeKeyName) {
	case agentACLTokenName:
		_, err = c.Consul.Agent().UpdateACLToken(*token, nil)
	case agentACLAgentTokenName:
		_, err = c.Consul.Agent().UpdateACLAgentToken(*token, nil)
	case agentACLAgentMasterTokenName:
		_, err = c.Consul.Agent().UpdateACLAgentMasterToken(*token, nil)
	case agentACLReplicationTokenName:
		_, err = c.Consul.Agent().UpdateACLReplicationToken(*token, nil)
	}
	if err != nil {
		log.Fatal(err.Error())
	}
}
