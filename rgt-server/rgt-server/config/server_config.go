package config

import (
	"fmt"
	"os"
	"path/filepath"
	"rgt-server/log"
	"rgt-server/option"
	"rgt-server/util"
	"strings"
	"time"

	"github.com/magiconair/properties"
)

type ServerConfig struct {
	filePathName               string
	options                    *option.Options
	address                    option.TypedOption[string]
	emulationPort              option.TypedOption[uint16]
	adminPort                  option.TypedOption[uint16]
	profilePort                option.TypedOption[uint16]
	teLogLevel                 option.TypedOption[log.LogLevel]
	teLogPathName              option.TypedOption[string]
	appLogLevel                option.TypedOption[log.LogLevel]
	appLogPathName             option.TypedOption[string]
	appLogDaysRetention        option.TypedOption[uint16]
	serverLogLevel             option.TypedOption[log.LogLevel]
	serverLogPathName          option.TypedOption[string]
	sessionIdleTimeout         option.TypedOption[time.Duration]
	sessionsCheckInterval      option.TypedOption[time.Duration]
	orphanProcessCheckInterval option.TypedOption[time.Duration]
	appLackTimeout             option.TypedOption[time.Duration]
	appLaunchTimeout           option.TypedOption[time.Duration]
	appLoginTimeout            option.TypedOption[time.Duration]
	appMinLaunchInterval       option.TypedOption[time.Duration]
	appTransactionTimeout      option.TypedOption[time.Duration]
	standaloneEnabled          option.TypedOption[bool]
	showConsole                option.TypedOption[bool]
	adminTCPReadBufferSize     option.TypedOption[uint32]
	adminTCPWriteBufferSize    option.TypedOption[uint32]
	terminalTCPReadBufferSize  option.TypedOption[uint32]
	terminalTCPWriteBufferSize option.TypedOption[uint32]
	adminFileTransferChunkSize option.TypedOption[uint32]
	healthEnabled              option.TypedOption[bool]
	healthCheckInterval        option.TypedOption[time.Duration]
	healthCpuThreshold         option.TypedOption[float64]
	maxCpuAlerts               option.TypedOption[uint16]
	healthCpuResumeThreshold   option.TypedOption[float64]
	healthMemThreshold         option.TypedOption[float64]
	healthMemResumeThreshold   option.TypedOption[float64]
	maxMemoryAlerts            option.TypedOption[uint16]
	healthDiskThreshold        option.TypedOption[float64]
	healthDiskResumeThreshold  option.TypedOption[float64]
	maxDiskAlerts              option.TypedOption[uint16]
	healthPendingLoginTimeout  option.TypedOption[time.Duration]
	healthMaxPendingLogins     option.TypedOption[uint16]
	maxPendingLoginsAlerts     option.TypedOption[uint16]
	maxTotalAlerts             option.TypedOption[uint16]
	envVars                    map[string]string
	mandatoryOptions           []option.Option
}

const (
	DEFAULT_EMULATION_PORT           uint16 = 7654
	DEFAULT_ADMIN_PORT               uint16 = 7656
	TERMINAL_AUTH_PREFIX             string = "terminal.authentication"
	ADMIN_AUTH_PREFIX                string = "admin.authentication"
	STANDALONE_AUTH_PREFIX           string = "standalone.authentication"
	EMULATION_SERVICE_ID             string = "emulation"
	ADMIN_SERVICE_ID                 string = "admin"
	STANDALONE_CONFIG_ID             string = "standalone"
	DEFAULT_ADMIN_TCP_BUFFER_SIZE    uint32 = 256 * 1024
	DEFAULT_TERMINAL_TCP_BUFFER_SIZE uint32 = 256 * 1024 // 256KB
	ADMIN_TRANSFER_FILE_CHUNK_SIZE   uint32 = 512 * 1024 // 512KB
)

func NewConfigWithName(filePathName string) *ServerConfig {
	config := &ServerConfig{
		filePathName: filePathName,
		options:      option.NewOptions(),
		envVars:      make(map[string]string)}

	config.address = option.NewString("", "server.address", "address")
	config.emulationPort = option.NewUint(DEFAULT_EMULATION_PORT, "server.port", "emulationPort")
	config.adminPort = option.NewUint(DEFAULT_ADMIN_PORT, "server.adminPort", "adminPort")
	config.profilePort = option.NewUint(uint16(0), "server.profile.port", "profilePort")
	config.teLogLevel = option.NewLogLevel(log.WARNING, "terminal.logLevel", "teLogLevel")
	config.teLogPathName = option.NewString("%USERPROFILE%\\rgt-terminal.log", "terminal.logPathName", "teLogPathName")
	config.appLogLevel = option.NewLogLevel(log.WARNING, "application.logLevel", "appLogLevel")
	config.appLogPathName = option.NewString("logs-app\\rgt-app-${pid}_${date}_${time}.log", "application.logPathName", "appLogPathName")
	config.appLogDaysRetention = option.NewUint(uint16(7), "application.logDaysRetention", "appLogDaysRetention")
	config.serverLogLevel = option.NewLogLevel(log.WARNING, "server.logLevel", "serverLogLevel")
	config.serverLogPathName = option.NewString("rgt-server.log", "server.logPathName", "serverLogPathName")
	config.sessionIdleTimeout = option.NewDuration(time.Hour, "server.sessionIdleTimeout", "sessionIdleTimeout")
	config.appLackTimeout = option.NewDuration(30*time.Minute, "application.communicationLackTimeout", "appLackTimeout")
	config.appLaunchTimeout = option.NewDuration(30*time.Second, "application.launchTimeout", "appLaunchTimeout")
	config.appLoginTimeout = option.NewDuration(30*time.Second, "application.loginTimeout", "appLoginTimeout")
	config.appMinLaunchInterval = option.NewDuration(500*time.Millisecond, "application.minLaunchInterval", "appMinLaunchInterval")
	config.appTransactionTimeout = option.NewDuration(15*time.Minute, "application.transactionTimeout", "appTransactionTimeout")
	config.standaloneEnabled = option.NewBool(false, "standalone.enabled", "standaloneEnabled")
	config.sessionsCheckInterval = option.NewDuration(10*time.Second, "server.sessionsCheckInterval", "sessionsCheckInterval")
	config.orphanProcessCheckInterval = option.NewDuration(5*time.Minute, "server.orphanProcessCheckInterval", "orphanProcessCheckInterval")
	config.adminTCPReadBufferSize = option.NewUint(DEFAULT_ADMIN_TCP_BUFFER_SIZE, "admin.TCPReadBufferSize", "adminTCPReadBufferSize")
	config.adminTCPWriteBufferSize = option.NewUint(DEFAULT_ADMIN_TCP_BUFFER_SIZE, "admin.TCPWriteBufferSize", "adminTCPWriteBufferSize")
	config.terminalTCPReadBufferSize = option.NewUint(DEFAULT_TERMINAL_TCP_BUFFER_SIZE, "terminal.TCPReadBufferSize", "teTCPReadBufferSize")
	config.terminalTCPWriteBufferSize = option.NewUint(DEFAULT_TERMINAL_TCP_BUFFER_SIZE, "terminal.TCPWriteBufferSize", "teTCPWriteBufferSize")
	config.adminFileTransferChunkSize = option.NewUint(ADMIN_TRANSFER_FILE_CHUNK_SIZE, "admin.fileTransferChunkSize", "adminFileTransferChunkSize")
	config.showConsole = option.NewBool(false, "application.console.show", "consoleShow")
	config.healthEnabled = option.NewBool(false, "server.health.enabled", "healthEnabled")
	config.healthCheckInterval = option.NewDuration(5*time.Second, "server.health.checkInterval", "healthCheckInterval")
	config.healthCpuThreshold = option.NewFloat(90.0, "server.health.cpuThreshold", "healthCpuThreshold")
	config.healthCpuResumeThreshold = option.NewFloat(80.0, "server.health.cpuResumeThreshold", "healthCpuResumeThreshold")
	config.maxCpuAlerts = option.NewUint(uint16(5), "server.health.maxCpuAlerts", "healthMaxCpuAlerts")
	config.healthMemThreshold = option.NewFloat(90.0, "server.health.memoryThreshold", "healthMemThreshold")
	config.healthMemResumeThreshold = option.NewFloat(80.0, "server.health.memoryResumeThreshold", "healthMemResumeThreshold")
	config.maxMemoryAlerts = option.NewUint(uint16(5), "server.health.maxMemoryAlerts", "healthMaxMemoryAlerts")
	config.healthDiskThreshold = option.NewFloat(95.0, "server.health.diskThreshold", "healthDiskThreshold")
	config.healthDiskResumeThreshold = option.NewFloat(90.0, "server.health.diskResumeThreshold", "healthDiskResumeThreshold")
	config.maxDiskAlerts = option.NewUint(uint16(5), "server.health.maxDiskAlerts", "healthMaxDiskAlerts")
	config.healthPendingLoginTimeout = option.NewDuration(2*time.Minute, "server.health.pendingLoginTimeout", "healthPendingLoginTimeout")
	config.healthMaxPendingLogins = option.NewUint(uint16(10), "server.health.maxPendingLogins", "healthMaxPendingLogins")
	config.maxPendingLoginsAlerts = option.NewUint(uint16(5), "server.health.maxPendingLoginsAlerts", "healthMaxPendingLoginsAlerts")
	config.maxTotalAlerts = option.NewUint(uint16(10), "server.health.maxTotalAlerts", "healthMaxTotalAlerts")

	config.options.Add(config.address)
	config.options.Add(config.emulationPort)
	config.options.Add(config.adminPort)
	config.options.Add(config.profilePort)
	config.options.Add(config.teLogLevel)
	config.options.Add(config.teLogPathName)
	config.options.Add(config.appLogLevel)
	config.options.Add(config.appLogPathName)
	config.options.Add(config.appLogDaysRetention)
	config.options.Add(config.serverLogLevel)
	config.options.Add(config.serverLogPathName)
	config.options.Add(config.sessionIdleTimeout)
	config.options.Add(config.appLackTimeout)
	config.options.Add(config.appLaunchTimeout)
	config.options.Add(config.appLoginTimeout)
	config.options.Add(config.appMinLaunchInterval)
	config.options.Add(config.appTransactionTimeout)
	config.options.Add(config.standaloneEnabled)
	config.options.Add(config.showConsole)
	config.options.Add(config.sessionsCheckInterval)
	config.options.Add(config.orphanProcessCheckInterval)
	config.options.Add(config.adminTCPReadBufferSize)
	config.options.Add(config.adminTCPWriteBufferSize)
	config.options.Add(config.terminalTCPReadBufferSize)
	config.options.Add(config.terminalTCPWriteBufferSize)
	config.options.Add(config.adminFileTransferChunkSize)
	config.options.Add(config.healthEnabled)
	config.options.Add(config.healthCheckInterval)
	config.options.Add(config.healthCpuThreshold)
	config.options.Add(config.healthCpuResumeThreshold)
	config.options.Add(config.maxCpuAlerts)
	config.options.Add(config.healthMemThreshold)
	config.options.Add(config.healthMemResumeThreshold)
	config.options.Add(config.maxMemoryAlerts)
	config.options.Add(config.healthDiskThreshold)
	config.options.Add(config.healthDiskResumeThreshold)
	config.options.Add(config.maxDiskAlerts)
	config.options.Add(config.healthPendingLoginTimeout)
	config.options.Add(config.healthMaxPendingLogins)
	config.options.Add(config.maxPendingLoginsAlerts)
	config.options.Add(config.maxTotalAlerts)
	config.mandatoryOptions = config.options.List()
	return config
}

func NewConfig() *ServerConfig {
	return NewConfigWithName("rgt_config.properties")
}

func (c *ServerConfig) SetFilePathName(filePathName string) {
	c.filePathName = filePathName
}

func (c *ServerConfig) GetFilePathName() string {
	return c.filePathName
}

func (c *ServerConfig) Address() option.TypedOption[string] {
	return c.address
}

func (c *ServerConfig) EmulationPort() option.TypedOption[uint16] {
	return c.emulationPort
}

func (c *ServerConfig) AdminPort() option.TypedOption[uint16] {
	return c.adminPort
}

func (c *ServerConfig) ProfilePort() option.TypedOption[uint16] {
	return c.profilePort
}

func (c *ServerConfig) TeLogLevel() option.TypedOption[log.LogLevel] {
	return c.teLogLevel
}

func (c *ServerConfig) TeLogPathName() option.TypedOption[string] {
	return c.teLogPathName
}

func (c *ServerConfig) AppLogLevel() option.TypedOption[log.LogLevel] {
	return c.appLogLevel
}

func (c *ServerConfig) AppLogPathName() option.TypedOption[string] {
	return c.appLogPathName
}

func (c *ServerConfig) AppLogDaysRetention() option.TypedOption[uint16] {
	return c.appLogDaysRetention
}

func (c *ServerConfig) ServerLogLevel() option.TypedOption[log.LogLevel] {
	return c.serverLogLevel
}

func (c *ServerConfig) ServerLogPathName() option.TypedOption[string] {
	return c.serverLogPathName
}

func (c *ServerConfig) SessionIdleTimeout() option.TypedOption[time.Duration] {
	return c.sessionIdleTimeout
}

func (c *ServerConfig) AppLackTimeout() option.TypedOption[time.Duration] {
	return c.appLackTimeout
}

func (c *ServerConfig) AppLoginTimeout() option.TypedOption[time.Duration] {
	return c.appLoginTimeout
}

func (c *ServerConfig) AppLaunchTimeout() option.TypedOption[time.Duration] {
	return c.appLaunchTimeout
}

func (c *ServerConfig) AppMinLaunchInterval() option.TypedOption[time.Duration] {
	return c.appMinLaunchInterval
}

func (c *ServerConfig) AppTransactionTimeout() option.TypedOption[time.Duration] {
	return c.appTransactionTimeout
}

func (c *ServerConfig) StandaloneEnabled() option.TypedOption[bool] {
	return c.standaloneEnabled
}

func (c *ServerConfig) ShowConsole() option.TypedOption[bool] {
	return c.showConsole
}

func (c *ServerConfig) StandaloneAuthConf() map[string]option.Option {
	return c.GetOptionsPrefix(STANDALONE_AUTH_PREFIX)
}

func (c *ServerConfig) SessionsCheckInterval() option.TypedOption[time.Duration] {
	return c.sessionsCheckInterval
}

func (c *ServerConfig) OrphanProcessCheckInterval() option.TypedOption[time.Duration] {
	return c.orphanProcessCheckInterval
}

func (c *ServerConfig) AdminTCPReadBufferSize() option.TypedOption[uint32] {
	return c.adminTCPReadBufferSize
}

func (c *ServerConfig) AdminTCPWriteBufferSize() option.TypedOption[uint32] {
	return c.adminTCPWriteBufferSize
}

func (c *ServerConfig) TerminalTCPReadBufferSize() option.TypedOption[uint32] {
	return c.terminalTCPReadBufferSize
}

func (c *ServerConfig) TerminalTCPWriteBufferSize() option.TypedOption[uint32] {
	return c.terminalTCPWriteBufferSize
}

func (c *ServerConfig) AdminFileTransferChunkSize() option.TypedOption[uint32] {
	return c.adminFileTransferChunkSize
}

func (c *ServerConfig) HealthEnabled() option.TypedOption[bool] {
	return c.healthEnabled
}

func (c *ServerConfig) HealthCheckInterval() option.TypedOption[time.Duration] {
	return c.healthCheckInterval
}

func (c *ServerConfig) HealthCpuThreshold() option.TypedOption[float64] {
	return c.healthCpuThreshold
}

func (c *ServerConfig) HealthCpuResumeThreshold() option.TypedOption[float64] {
	return c.healthCpuResumeThreshold
}

func (c *ServerConfig) MaxCpuAlerts() option.TypedOption[uint16] {
	return c.maxCpuAlerts
}

func (c *ServerConfig) HealthMemThreshold() option.TypedOption[float64] {
	return c.healthMemThreshold
}

func (c *ServerConfig) HealthMemResumeThreshold() option.TypedOption[float64] {
	return c.healthMemResumeThreshold
}

func (c *ServerConfig) MaxMemoryAlerts() option.TypedOption[uint16] {
	return c.maxMemoryAlerts
}

func (c *ServerConfig) HealthDiskThreshold() option.TypedOption[float64] {
	return c.healthDiskThreshold
}

func (c *ServerConfig) HealthDiskResumeThreshold() option.TypedOption[float64] {
	return c.healthDiskResumeThreshold
}

func (c *ServerConfig) MaxDiskAlerts() option.TypedOption[uint16] {
	return c.maxDiskAlerts
}

func (c *ServerConfig) HealthPendingLoginTimeout() option.TypedOption[time.Duration] {
	return c.healthPendingLoginTimeout
}

func (c *ServerConfig) HealthMaxPendingLogins() option.TypedOption[uint16] {
	return c.healthMaxPendingLogins
}

func (c *ServerConfig) MaxPendingLoginsAlerts() option.TypedOption[uint16] {
	return c.maxPendingLoginsAlerts
}

func (c *ServerConfig) MaxTotalAlerts() option.TypedOption[uint16] {
	return c.maxTotalAlerts
}

func (c *ServerConfig) GetOptionsPrefix(prefix string) map[string]option.Option {
	options := make(map[string]option.Option)
	for _, op := range c.options.List() {
		for _, name := range op.Names() {
			if strings.HasPrefix(name, prefix) {
				options[name] = op
			}
		}
	}
	return options
}

func (c *ServerConfig) AdminAuthConf() map[string]option.Option {
	return c.GetOptionsPrefix(ADMIN_AUTH_PREFIX)
}

func (c *ServerConfig) TeAuthConf() map[string]option.Option {
	return c.GetOptionsPrefix(TERMINAL_AUTH_PREFIX)
}

func LoadConfig() (*ServerConfig, error) {
	return LoadConfigFromFile("rgt_config.properties")
}

func LoadConfigFromFile(filePathname string) (*ServerConfig, error) {
	props, err := properties.LoadFile(filePathname, properties.UTF8)
	if err != nil {
		return nil, err
	}
	conf := NewConfigWithName(filePathname)
	conf.SetFromMap(props.Map())
	createLogsDirectories(conf)
	return conf, nil
}

func createLogsDirectories(conf *ServerConfig) {
	os.MkdirAll(filepath.Dir(util.RelativePathToAbsolute(conf.serverLogPathName.Get())), os.ModePerm)
	os.MkdirAll(filepath.Dir(util.RelativePathToAbsolute(conf.appLogPathName.Get())), os.ModePerm)
}

func (conf *ServerConfig) SetFromMap(values map[string]string) {
	for key, value := range values {
		optionName := strings.TrimSpace(key)
		optionValue := strings.TrimSpace(value)
		op := conf.options.Get(optionName)
		if op != nil {
			op.SetString(strings.TrimSpace(value))
		} else if strings.HasPrefix(optionName, "env.") {
			conf.envVars[optionName[4:]] = optionValue
		} else {
			conf.options.Add(option.NewString(optionValue, optionName))
		}
	}
}

func (conf *ServerConfig) Reload() error {
	props, err := properties.LoadFile(conf.filePathName, properties.UTF8)
	if err != nil {
		return err
	}
	conf.SetFromMap(props.Map())
	return nil
}

func (conf *ServerConfig) ToProperties() *properties.Properties {
	saved := make(map[string]bool)
	props := properties.NewProperties()
	for _, op := range conf.options.List() {
		_, found := saved[op.Name()]
		if !found {
			props.Set(op.Name(), op.GetString())
			saved[op.Name()] = true
		}
	}
	for key, value := range conf.envVars {
		props.Set("env."+key, value)
	}
	return props
}

func (conf *ServerConfig) ToMap() map[string]string {
	return conf.ToProperties().Map()
}

func (conf *ServerConfig) Save() error {
	fileHandle, err := os.OpenFile(conf.filePathName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	p := conf.ToProperties()
	p.Sort()
	_, err = p.Write(fileHandle, properties.UTF8)
	return err
}

func (c *ServerConfig) Set(name string, value string) bool {
	optionName := strings.TrimSpace(name)
	optionValue := strings.TrimSpace(value)
	op := c.options.Get(optionName)
	if op != nil {
		op.SetString(optionValue)
	} else if strings.HasPrefix(optionName, "env.") {
		c.envVars[optionName[4:]] = optionValue
	} else {
		c.options.Add(option.NewString(optionValue, optionName))
	}
	return true
}

func (c *ServerConfig) Get(name string) option.Option {
	return c.options.Get(strings.TrimSpace(name))
}

func (c *ServerConfig) GetValue(name string) string {
	if strings.HasPrefix(name, "env.") {
		return c.envVars[name[4:]]
	}
	op := c.options.Get(strings.TrimSpace(name))
	if op != nil {
		return op.GetString()
	}
	return ""
}

func (c *ServerConfig) GetEnvVars() map[string]string {
	return c.envVars
}

func (c *ServerConfig) GetEnv(envName string) string {
	optionName := strings.TrimSpace(envName)
	if strings.HasPrefix(optionName, "env.") {
		return c.envVars[optionName[4:]]
	} else {
		return c.envVars[optionName]
	}
}

func (c *ServerConfig) Delete(name string) (bool, error) {
	optionName := strings.TrimSpace(name)
	for _, o := range c.mandatoryOptions {
		if o.HasName(optionName) {
			return false, fmt.Errorf("cannot remove mandatory otion: %s", name)
		}
	}

	if strings.HasPrefix(optionName, "env.") {
		_, found := c.envVars[optionName[4:]]
		if found {
			delete(c.envVars, optionName[4:])
			return true, nil
		}
		return false, nil
	}
	op := c.options.Delete(optionName)
	if op != nil {
		return true, nil
	}
	_, found := c.envVars[optionName]
	if found {
		delete(c.envVars, optionName)
		return true, nil
	}
	return false, nil
}
