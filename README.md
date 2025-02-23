# konfig

A simple, configuration management library for Go. Offloads worrying about hooking all configuration into some cental file from you onto konfig.

## Features

- Allows for unrelated configuration to live in the same config file
- Automatically manages configuration loading and saving
- Supports JSON, YAML, and TOML
- Supports validation and onLoad callbacks
- Allows for completely custom flow, since you can decide when the data is saved or loaded

## Planned features

- Support more file formats
- Allow the user to supply their own format handler
- Automatic saving when data changes in a section(not sure if this one is going to be possible in the go version)
- *Have a feature suggestion? Open an issue!*

## Installation

```bash
go get github.com/kociumba/konfig
```

## Quick Start

This is a very simple example, for more complex ones look in [examples](https://github.com/kociumba/konfig-go/tree/main/examples).

```go
package main

import (
    "github.com/kociumba/konfig"
    "log"
)

// define your configuration struct (use struct tags to provide names for the format you want to use)
type DatabaseConfig struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Username string `json:"username"`
    Password string `json:"password"`
}

// you define defaults, if you want
var dbConfig = DatabaseConfig{
    Host:     "localhost",
    Port:     5432,
    Username: "admin",
    Password: "default_password",
}

func main() {
    // Create a configuration manager
    mngr, err := konfig.NewKonfigManager(konfig.KonfigOptions{
        Format:       konfig.JSON,        // Use JSON format
        AutoLoad:     true,               // Load configs when the manager is created
        AutoSave:     true,               // Save on SIGINT/SIGTERM
        UseCallbacks: true,               // Enable validation callbacks
        KonfigPath:   "config.json",      // Config file path (absolute or relative to wd)
    })
    if err != nil {
        log.Fatalf("Failed to create config manager: %v", err)
    }

    // Register configuration section
    section := konfig.NewKonfigSection(&dbConfig,
        konfig.WithSectionName(func() string { return "database" }),
        konfig.WithOnLoad(func() error {
            log.Println("Database configuration loaded")
            return nil
        }),
    )

    if err := mngr.RegisterSection(section); err != nil {
        log.Fatalf("Failed to register section: %v", err)
    }

    // ensure configuration is saved on exit
    defer func() {
        if err := mngr.Save(); err != nil {
            log.Printf("Error saving configuration: %v", err)
        }
    }()

    // access your config straight from the struct itself
    fmt.Printf("Database host: %s\n", dbConfig.Host)

    // rest of your application code here...
}
```

## Detailed Usage

### Configuration Manager Options

The `KonfigOptions` struct provides several options to customize the behavior of the configuration manager:

```go
type KonfigOptions struct {
    Format       EncodingFormat // JSON, YAML, or TOML
    AutoLoad     bool          // Load config automatically on startup
    AutoSave     bool          // Save on program termination
    UseCallbacks bool          // Enable validation/callbacks
    KonfigPath   string        // Path to config file
}
```

### Creating Configuration Sections

1. Define your configuration struct with appropriate tags (of course only include the tags for the format you want to use):
```go
type MyConfig struct {
    Setting1 string        `json:"setting1" yaml:"setting1" toml:"setting1"`
    Setting2 int           `json:"setting2" yaml:"setting2" toml:"setting2"`
    Timeout time.Duration  `json:"timeout" yaml:"timeout" toml:"timeout"`
}
```

2. Create and register a section:
```go
myConfig := MyConfig{
    Setting1: "default",
    Setting2: 42,
    Timeout: time.Second * 30,
}

section := konfig.NewKonfigSection(&myConfig,
    konfig.WithSectionName(func() string { return "my_section" }),
    konfig.WithOnLoad(func() error {
        // Called after loading
        return nil
    }),
    konfig.WithValidation(func() error {
        // Validate configuration values
        if myConfig.Setting2 < 0 {
            return fmt.Errorf("setting2 must be positive")
        }
        return nil
    }),
)

// or use the shorthand for a basic registration
mngr.AddSimpleSection("my_section", &myConfig)

mngr.RegisterSection(section)
```

### File Formats

Konfig supports three configuration file formats:

1. **JSON** (`konfig.JSON`):
```json
{
    "my_section": {
        "setting1": "value",
        "setting2": 42,
        "timeout": 30000000000
    }
}
```

2. **YAML** (`konfig.YAML`):
```yaml
my_section:
  setting1: value
  setting2: 42
  timeout: 30000000000
```

3. **TOML** (`konfig.TOML`):
```toml
[my_section]
setting1 = "value"
setting2 = 42
timeout = 30000000000
```

### Best Practices

1. **Always defer Save() in your main()**
```go
defer mngr.Save()
```

2. **Use validation callbacks**
```go
// Validation and OnLoad are functionally the same but they provide a nice way to separate concerns
konfig.WithValidation(func() error {
    if config.Port < 1 || config.Port > 65535 {
        return fmt.Errorf("port must be between 1 and 65535")
    }
    return nil
})
```

3. **Try to always save after modifying state**
```go
// doing this ensures no data will be lost by overwriting it with old values whe calling Load(), the one exception is when you have a dedicated settings menu with a save button
myConfig.Setting1 = "new value"
myConfig.Setting2 = 123
if err := mngr.Save(); err != nil {
    log.Printf("Error saving configuration: %v", err)
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Author

Created by [@kociumba](https://github.com/kociumba)
