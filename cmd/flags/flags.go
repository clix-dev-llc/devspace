package flags

import (
	"github.com/devspace-cloud/devspace/pkg/devspace/config/generated"
	"github.com/devspace-cloud/devspace/pkg/devspace/config/loader"
	"github.com/devspace-cloud/devspace/pkg/util/log"
	"github.com/mgutz/ansi"

	flag "github.com/spf13/pflag"
)

// GlobalFlags is the flags that contains the global flags
type GlobalFlags struct {
	Silent bool
	NoWarn bool
	Debug  bool

	Namespace      string
	KubeContext    string
	Profile        string
	ProfileRefresh bool
	ConfigPath     string
	Vars           []string

	SwitchContext bool

	Flags *flag.FlagSet
}

// UseLastContext uses the last context
func (gf *GlobalFlags) UseLastContext(generatedConfig *generated.Config, log log.Logger) error {
	if gf.KubeContext == "" && gf.Namespace == "" && gf.SwitchContext == true {
		if generatedConfig == nil || generatedConfig.GetActive().LastContext == nil {
			log.Warn("There is no last context to use. Only use the '--switch-context / -s' flag if you already have deployed the project before")
		} else {
			gf.KubeContext = generatedConfig.GetActive().LastContext.Context
			gf.Namespace = generatedConfig.GetActive().LastContext.Namespace

			log.Infof("Switching to context '%s' and namespace '%s'", ansi.Color(gf.KubeContext, "white+b"), ansi.Color(gf.Namespace, "white+b"))
			return nil
		}
	}

	gf.SwitchContext = false
	return nil
}

// ToConfigOptions converts the globalFlags into config options
func (gf *GlobalFlags) ToConfigOptions() *loader.ConfigOptions {
	return &loader.ConfigOptions{
		Profile:        gf.Profile,
		ProfileRefresh: gf.ProfileRefresh,
		ConfigPath:     gf.ConfigPath,
		KubeContext:    gf.KubeContext,
		Namespace:      gf.Namespace,
		Vars:           gf.Vars,
	}
}

// SetGlobalFlags applies the global flags
func SetGlobalFlags(flags *flag.FlagSet) *GlobalFlags {
	globalFlags := &GlobalFlags{
		Vars:  []string{},
		Flags: flags,
	}

	flags.BoolVar(&globalFlags.NoWarn, "no-warn", false, "If true does not show any warning when deploying into a different namespace or kube-context than before")
	flags.BoolVar(&globalFlags.Debug, "debug", false, "Prints the stack trace if an error occurs")
	flags.BoolVar(&globalFlags.Silent, "silent", false, "Run in silent mode and prevents any devspace log output except panics & fatals")

	flags.StringVar(&globalFlags.ConfigPath, "config", "", "The devspace config file to use")
	flags.StringVarP(&globalFlags.Profile, "profile", "p", "", "The devspace profile to use (if there is any)")
	flags.BoolVar(&globalFlags.ProfileRefresh, "profile-refresh", false, "If true will pull and re-download profile parent sources")
	flags.StringVarP(&globalFlags.Namespace, "namespace", "n", "", "The kubernetes namespace to use")
	flags.StringVar(&globalFlags.KubeContext, "kube-context", "", "The kubernetes context to use")
	flags.BoolVarP(&globalFlags.SwitchContext, "switch-context", "s", false, "Switches and uses the last kube context and namespace that was used to deploy the DevSpace project")
	flags.StringSliceVar(&globalFlags.Vars, "var", []string{}, "Variables to override during execution (e.g. --var=MYVAR=MYVALUE)")

	return globalFlags
}
