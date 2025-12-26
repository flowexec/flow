package filesystem

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/types/config"
)

const FlowConfigDirEnvVar = "FLOW_CONFIG_DIR"

func ConfigDirPath() string {
	if dir := os.Getenv(FlowConfigDirEnvVar); dir != "" {
		return dir
	}

	dirname, err := os.UserConfigDir()
	if err != nil {
		panic(errors.Wrap(err, "unable to get config directory"))
	}
	return filepath.Join(dirname, dataDirName)
}

func UserConfigFilePath() string {
	return ConfigDirPath() + "/config.yaml"
}

func EnsureConfigDir() error {
	if _, err := os.Stat(ConfigDirPath()); os.IsNotExist(err) {
		err = os.MkdirAll(ConfigDirPath(), 0750)
		if err != nil {
			return errors.Wrap(err, "unable to create config directory")
		}
	} else if err != nil {
		return errors.Wrap(err, "unable to check for config directory")
	}
	return nil
}

func InitConfig() error {
	if err := EnsureDefaultWorkspace(); err != nil {
		return errors.Wrap(err, "unable to ensure default workspace")
	}
	defaultWsName := "default"
	if _, err := os.Stat(filepath.Join(DefaultWorkspaceDir(), WorkspaceConfigFileName)); os.IsNotExist(err) {
		if err := InitWorkspaceConfig(defaultWsName, DefaultWorkspaceDir()); err != nil {
			return errors.Wrap(err, "unable to initialize default workspace config")
		}
	} else if err != nil {
		return errors.Wrap(err, "unable to check for default workspace config")
	}

	defaultCfg := &config.Config{
		Workspaces:       map[string]string{defaultWsName: DefaultWorkspaceDir()},
		CurrentWorkspace: defaultWsName,
		WorkspaceMode:    config.ConfigWorkspaceModeDynamic,
		Interactive: &config.Interactive{
			Enabled: true,
		},
		DefaultLogMode: "logfmt",
	}

	_, err := os.Create(UserConfigFilePath())
	if err != nil {
		return errors.Wrap(err, "unable to create config file")
	}
	err = WriteConfig(defaultCfg)
	if err != nil {
		return errors.Wrap(err, "unable to write default config")
	}
	return nil
}

func WriteConfig(config *config.Config) error {
	file, err := os.OpenFile(filepath.Clean(UserConfigFilePath()), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return errors.Wrap(err, "unable to open config file")
	}
	defer file.Close()

	if err := file.Truncate(0); err != nil {
		return errors.Wrap(err, "unable to truncate config file")
	}

	err = yaml.NewEncoder(file).Encode(config)
	if err != nil {
		return errors.Wrap(err, "unable to encode config file")
	}

	return nil
}

func LoadConfig() (*config.Config, error) {
	if err := EnsureConfigDir(); err != nil {
		return nil, errors.Wrap(err, "unable to ensure existence of config directory")
	}

	if _, err := os.Stat(UserConfigFilePath()); os.IsNotExist(err) {
		if err := InitConfig(); err != nil {
			return nil, errors.Wrap(err, "unable to initialize config file")
		}
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to stat config file")
	}

	file, err := os.Open(UserConfigFilePath())
	if err != nil {
		return nil, errors.Wrap(err, "unable to open config file")
	}
	defer file.Close()

	userCfg := &config.Config{}
	err = yaml.NewDecoder(file).Decode(userCfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode config file")
	}

	if err := userCfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "encountered validation error")
	}

	return userCfg, nil
}
