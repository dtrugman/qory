package config

// Origin describes where a configuration value came from.
type Origin int

const (
	OriginNotSet  Origin = iota // value is not configured
	OriginDefault               // built-in default is being used
	OriginUser                  // explicitly set in the config file
	OriginEnv                   // sourced from an environment variable
)

func (o Origin) String() string {
	switch o {
	case OriginNotSet:
		return "Not set"
	case OriginDefault:
		return "Default"
	case OriginUser:
		return "Set by user"
	case OriginEnv:
		return "From env"
	default:
		return "Unknown"
	}
}
