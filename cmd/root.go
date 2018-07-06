package cmd

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// ConsulTokenParamFlagName is the flag which sets the
	// SSM parameter name storing a Consul management token
	ConsulTokenParamFlagName = "consul-token-param"

	// KMSKeyIDFlagName is the flag which sets the
	// KMS key id used for encrypting new ACL token IDs
	KMSKeyIDFlagName = "kms-key-id"

	// InsecureFlagName is the flag which sets whether
	// new ACL IDs should be written unencrypted
	InsecureFlagName = "insecure"

	// OverwriteFlagName is the flag which sets whether
	// existing SSM parameters can be overwritten
	OverwriteFlagName = "overwrite"

	// DebugFlagName is the flag which sets whether
	// debug logging will be enabled
	DebugFlagName = "debug"
)

// Formatter is the struct used in the logging package.
type Formatter struct {
}

// the root command of the application
var rootCmd = &cobra.Command{
	Use:   "consulssm",
	Short: "Bootstrap and manage Consul ACLs via AWS SSM parameters",
}

func init() {
	log.SetFormatter(&Formatter{})
	log.SetLevel(log.InfoLevel)

	rootCmd.PersistentFlags().Bool(DebugFlagName, false, "Enable debug logging")
	viper.BindPFlag(DebugFlagName, rootCmd.PersistentFlags().Lookup(DebugFlagName))

	viper.SetEnvPrefix("ssm")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}

// Execute is the main entrypoint function into the cli app
func Execute() {
	rootCmd.AddCommand(bootstrapCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(agentCmd)

	if os.Getenv("AWS_REGION") == "" {
		os.Setenv("AWS_REGION", "us-east-1")
	}

	rootCmd.Execute()
}

// Format builds the log desired log format.
func (c *Formatter) Format(entry *log.Entry) ([]byte, error) {
	return []byte(fmt.Sprintf("%s [%s] %s\n", entry.Time.Format("2006/01/02 15:04:05"), strings.ToUpper(entry.Level.String()), entry.Message)), nil
}

// AddStringFlag adds a string flag to a command.
func AddStringFlag(cmd *cobra.Command, name, shorthand, def, desc string) {
	cmd.Flags().StringP(name, shorthand, def, desc)
	viper.BindPFlag(name, cmd.Flags().Lookup(name))
}

// AddInt64Flag adds an int64 flag to a command.
func AddInt64Flag(cmd *cobra.Command, name, shorthand string, def int64, desc string) {
	cmd.Flags().Int64P(name, shorthand, def, desc)
	viper.BindPFlag(name, cmd.Flags().Lookup(name))
}

// AddBoolFlag adds an integr flag to a command.
func AddBoolFlag(cmd *cobra.Command, name, shorthand string, def bool, desc string) {
	cmd.Flags().BoolP(name, shorthand, def, desc)
	viper.BindPFlag(name, cmd.Flags().Lookup(name))
}

// func bindFlag(cmd *cobra.Command, name string) {
// 	viper.BindPFlag(name, cmd.Flags().Lookup(name))
// }

func bindFlag(cmd *cobra.Command, names ...string) {
	for _, n := range names {
		viper.BindPFlag(n, cmd.Flags().Lookup(n))
	}
}

func bail(err error, code int) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(code)
}

func usageError(cmd *cobra.Command, message string, codeOptional ...int) {
	code := 1
	if len(codeOptional) == 1 {
		code = codeOptional[0]
	}

	cmd.Usage()

	if message != "" {
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, message)
	}

	os.Exit(code)
}
