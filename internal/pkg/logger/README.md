# Logger
This package implements a logger using zerolog that should be used for application logging.

## Usage

### Importing the package
```go
import github.com/nawiridigital/pkg/logger
```

### Creating logger
Simplest way to instantiate a new logger is

```go
l := logger.New(nil)
```

The `New()` constructor function takes in a logger configuration object
```go
type Config struct {
	LogMode string
	Service string // unique name 
}
```

`LogMode` can be any of
```go
	LModeSilent = "SILENT"
	LModeTrace  = "TRACE"
	LModeInfo   = "INFO"
	LModeWarn   = "WARN"
	LModeError  = "ERROR"
	LModeDebug  = "DEBUG"
```

### Setting default logger
Once you create a logger instance you can set it as the default logger in your application

```go
// create a logger instance
l := logger.New(nil)
// set it as default logger instance
logger.SetDefaultLogger(l)

// fetch the default logger instance
l = logger.Get()
```

### Using the logger
Usages of the logger should be similar to a zerolog.Logger instance

```go
l := logger.New(nil)
l.Info().Msg("example log")
l.Err().Error(err).Msg("database error")
```