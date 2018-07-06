package cmd

import (
	"fmt"
	"os"

	"github.com/bdclark/consulssm/acl"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// HideBootstrapFlagName is the flag which sets whether
	// the bootrap token will be hidden from standard output
	HideBootstrapFlagName = "hide"
)

// command to bootstrap Consul ACLs
var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap Consul ACLs and save token to an SSM parameter",
	PreRun: func(cmd *cobra.Command, args []string) {
		// bind commonly-named flags only when command is executed
		// https://github.com/spf13/viper/issues/233
		bindFlag(cmd, KMSKeyIDFlagName, InsecureFlagName, ConsulTokenParamFlagName, OverwriteFlagName, HideBootstrapFlagName)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetBool(DebugFlagName) {
			log.SetLevel(log.DebugLevel)
		}

		consulTokenParam := viper.GetString(ConsulTokenParamFlagName)
		if consulTokenParam == "" {
			usageError(cmd, "SSM parameter name to write Consul bootstrap token ID is required", 1)
		}

		c, err := acl.NewClientSet(&acl.ClientSetInput{
			KMSKeyID:  viper.GetString(KMSKeyIDFlagName),
			Overwrite: viper.GetBool(OverwriteFlagName),
			Insecure:  viper.GetBool(InsecureFlagName),
		})
		if err != nil {
			bail(err, 1)
		}

		id, err := c.Bootstrap(consulTokenParam)

		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			if id == "" {
				os.Exit(1)
			} else {
				fmt.Println(id)
				os.Exit(255)
			}
		}
		if viper.GetBool(HideBootstrapFlagName) {
			fmt.Println(id)
		}
	},
}

func init() {
	bootstrapCmd.Flags().StringP(KMSKeyIDFlagName, "k", "", "Optional KMS key ID for encrypting bootstrap token ID")
	bootstrapCmd.Flags().BoolP(InsecureFlagName, "I", false, "Skip encryption when writing token to SSM")
	bootstrapCmd.Flags().StringP(ConsulTokenParamFlagName, "m", "", "SSM parameter name to write Consul bootstrap token ID")
	bootstrapCmd.Flags().BoolP(OverwriteFlagName, "o", false, "Overwrite existing SSM parameter value if it exists")
	bootstrapCmd.Flags().Bool(HideBootstrapFlagName, false, "Hide bootstrap token from standard output")
}
