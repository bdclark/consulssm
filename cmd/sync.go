package cmd

import (
	"time"

	"github.com/spf13/viper"

	"github.com/bdclark/consulssm/acl"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	// PageSizeFlagName is the flag name which sets the
	// maximum results per SSM query
	PageSizeFlagName = "page-size"

	// ACLDefinitionPrefixFlagName is the flag which sets the
	// prefix used to query for SSM parameters containing ACL definitions
	ACLDefinitionPrefixFlagName = "definition-prefix"

	// ACLIDPrefixFlagName is the flag which sets the
	// prefix used to query/set SSM parameters containing ACL IDs
	ACLIDPrefixFlagName = "id-prefix"

	// RequireLeaderFlagName is the flag which sets whether
	// ACL management requires the agent to be the current leader
	RequireLeaderFlagName = "leader"

	// RecurringFlagName is the flag which sets the number
	// of seconds between recurring ACL syncs
	RecurringFlagName = "recurring"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize Consul ACLs via SSM parameters",
	PreRun: func(cmd *cobra.Command, args []string) {
		// bind non-unique flags only when command is executed
		// https://github.com/spf13/viper/issues/233
		bindFlag(cmd, KMSKeyIDFlagName, InsecureFlagName, ConsulTokenParamFlagName, OverwriteFlagName)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetBool(DebugFlagName) {
			log.SetLevel(log.DebugLevel)
		}

		consulTokenParam := viper.GetString(ConsulTokenParamFlagName)
		definitionPrefix := viper.GetString(ACLDefinitionPrefixFlagName)

		if consulTokenParam == "" {
			usageError(cmd, "SSM parameter for Consul management token is required", 1)
		}
		if definitionPrefix == "" {
			usageError(cmd, "SSM prefix is required to read Consul ACL definitions", 1)
		}

		c, err := acl.NewClientSet(&acl.ClientSetInput{
			ConsulTokenParam: consulTokenParam,
			KMSKeyID:         viper.GetString(KMSKeyIDFlagName),
			Overwrite:        viper.GetBool(OverwriteFlagName),
			Insecure:         viper.GetBool(InsecureFlagName),
		})
		if err != nil {
			log.Fatal(err.Error())
		}

		syncInput := &acl.SyncInput{
			ACLDefinitionPrefix: definitionPrefix,
			ACLIDPrefix:         viper.GetString(ACLIDPrefixFlagName),
			PageSize:            viper.GetInt64(PageSizeFlagName),
			OnlyIfConsulLeader:  viper.GetBool(RequireLeaderFlagName),
		}

		recurring := viper.GetInt64(RecurringFlagName)
		if recurring > 0 {
			for {
				if err := c.Sync(syncInput); err != nil {
					log.Fatal(err.Error())
				}
				time.Sleep(time.Duration(recurring) * time.Second)
			}
		} else if err := c.Sync(syncInput); err != nil {
			log.Fatal(err.Error())
		}
	},
}

func init() {
	syncCmd.Flags().StringP(KMSKeyIDFlagName, "k", "", "Optional KMS key ID for encrypting new token IDs")
	syncCmd.Flags().BoolP(InsecureFlagName, "I", false, "Skip encryption when updating SSM with new token IDs")
	syncCmd.Flags().StringP(ConsulTokenParamFlagName, "m", "", "SSM parameter name for Consul management token")
	syncCmd.Flags().BoolP(OverwriteFlagName, "o", false, "Overwrite existing SSM parameter values if they exist")

	AddStringFlag(syncCmd, ACLDefinitionPrefixFlagName, "d", "", "SSM heirarchy prefix to read ACL definitions (required)")
	AddStringFlag(syncCmd, ACLIDPrefixFlagName, "i", "", "SSM heirarchy prefix to read/write ACL token IDs")
	AddInt64Flag(syncCmd, PageSizeFlagName, "p", 0, "Maximum results per SSM query")
	AddBoolFlag(syncCmd, RequireLeaderFlagName, "l", false, "Manage ACLs only if Consul agent is current leader")
	AddInt64Flag(syncCmd, RecurringFlagName, "r", 0, "Make recurring and wait given number of seconds between syncs")
}
