package health

type AlertType string

type MetricChecker interface {
	Type() AlertType
	Check() bool
	Unhealth() bool
	Alerts() uint
}

type metricChecker struct {
	config HealthConfig
	alerts uint
}
