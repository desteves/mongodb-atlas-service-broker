package broker

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/cloudfoundry-incubator/candiedyaml"
)

type Config struct {
	AtlasConfiguration              ServiceConfiguration `yaml:"atlas"`
	AuthConfiguration               AuthConfiguration    `yaml:"auth"`
	Host                            string               `yaml:"backend_host"`
	Port                            string               `yaml:"backend_port"`
	MonitExecutablePath             string               `yaml:"monit_executable_path"`
	AtlasServerExecutablePath       string               `yaml:"atlas_server_executable_path"`
	AgentPort                       string               `yaml:"agent_port"`
	ConsistencyVerificationInterval int                  `yaml:"consistency_check_interval_seconds"`
}

type AuthConfiguration struct {
	Password string `yaml:"password"`
	Username string `yaml:"username"`
}

type ServiceConfiguration struct {
	ServiceName                 string    `yaml:"service_name"`
	ServiceID                   string    `yaml:"service_id"`
	DedicatedVMPlanID           string    `yaml:"dedicated_vm_plan_id"`
	SharedVMPlanID              string    `yaml:"shared_vm_plan_id"`
	Host                        string    `yaml:"host"`
	DefaultConfigPath           string    `yaml:"atlas_conf_path"`
	ProcessCheckIntervalSeconds int       `yaml:"process_check_interval"`
	StartAtlasTimeoutSeconds    int       `yaml:"start_atlas_timeout"`
	InstanceDataDirectory       string    `yaml:"data_directory"`
	PidfileDirectory            string    `yaml:"pidfile_directory"`
	InstanceLogDirectory        string    `yaml:"log_directory"`
	ServiceInstanceLimit        int       `yaml:"service_instance_limit"`
	Dedicated                   Dedicated `yaml:"dedicated"`
	Description                 string    `yaml:"description"`
	LongDescription             string    `yaml:"long_description"`
	ProviderDisplayName         string    `yaml:"provider_display_name"`
	DocumentationURL            string    `yaml:"documentation_url"`
	SupportURL                  string    `yaml:"support_url"`
	DisplayName                 string    `yaml:"display_name"`
	IconImage                   string    `yaml:"icon_image"`
}

type Dedicated struct {
	Nodes         []string `yaml:"nodes"`
	Port          int      `yaml:"port"`
	StatefilePath string   `yaml:"statefile_path"`
}

func (config *Config) DedicatedEnabled() bool {
	return len(config.AtlasConfiguration.Dedicated.Nodes) > 0
}

func (config *Config) SharedEnabled() bool {
	return config.AtlasConfiguration.ServiceInstanceLimit > 0
}

func ParseConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := candiedyaml.NewDecoder(file).Decode(&config); err != nil {
		return Config{}, err
	}

	return config, ValidateConfig(config.AtlasConfiguration)
}

func ValidateConfig(config ServiceConfiguration) error {
	err := checkPathExists(config.DefaultConfigPath, "AtlasConfig.DefaultAtlasConfPath")
	if err != nil {
		return err
	}

	err = checkPathExists(config.InstanceDataDirectory, "AtlasConfig.InstanceDataDirectory")
	if err != nil {
		return err
	}

	err = checkPathExists(config.InstanceLogDirectory, "AtlasConfig.InstanceLogDirectory")
	if err != nil {
		return err
	}

	err = checkDedicatedNodesAreIPs(config.Dedicated.Nodes)
	if err != nil {
		return err
	}

	return nil
}

func checkPathExists(path string, description string) error {
	_, err := os.Stat(path)
	if err != nil {
		errMessage := fmt.Sprintf(
			"File '%s' (%s) not found",
			path,
			description)
		return errors.New(errMessage)
	}
	return nil
}

func checkDedicatedNodesAreIPs(dedicatedNodes []string) error {
	validIPField := "(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)"
	ipRegex := fmt.Sprintf("^(%[1]s\\.){3}%[1]s$", validIPField)

	for _, nodeAddress := range dedicatedNodes {
		match, _ := regexp.MatchString(ipRegex, strings.TrimSpace(nodeAddress))
		if !match {
			return errors.New("The broker only supports IP addresses for dedicated nodes")
		}
	}
	return nil
}
