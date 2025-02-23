package konfig

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"
)

type EncodingFormat int

const (
	JSON EncodingFormat = iota
	YAML
	TOML
)

// the options you can pass when creating a new KonfigManager
type KonfigOptions struct {
	Format       EncodingFormat // the format to use for the config file, available: JSON, YAML, TOML
	AutoLoad     bool           // will try to automatically load the data when the manager is created
	AutoSave     bool           // if true, will save the configuration file on SIGINT and SIGTERM, you still need to defer a call to Save() in your main() function
	UseCallbacks bool           // whether to call the OnLoad() and Validate() callbacks on each section
	KonfigPath   string         // the path to the config file, no validation is done on the path, it is up to the user to make sure it is correct
}

type KonfigManager struct {
	opts          KonfigOptions
	sections      map[string]KonfigSection
	formatHnadler FormatHandler
}

func NewKonfigManager(opt KonfigOptions) (*KonfigManager, error) {
	fmtHandler, err := createFormatHandler(opt.Format)
	if err != nil {
		return nil, fmt.Errorf("error creating format handler: %v", err)
	}

	if opt.KonfigPath != "" {
		if _, err := os.Stat(opt.KonfigPath); err != nil {
			if os.IsNotExist(err) {
				f, err := os.Create(opt.KonfigPath)
				if err != nil {
					return nil, fmt.Errorf("error creating configuration file: %v", err)
				}
				f.Close()
			} else {
				return nil, fmt.Errorf("error checking configuration file: %v", err)
			}
		}
	}

	mngr := &KonfigManager{
		opts:          opt,
		sections:      make(map[string]KonfigSection),
		formatHnadler: fmtHandler,
	}

	if opt.AutoLoad {
		if err := mngr.Load(); err != nil {
			return nil, fmt.Errorf("automatic configuration loading failed: %v.\nTo resolve this:\n1. Ensure config file exists at '%s'\n2. Disable AutoLoad option\n3. Check file permissions", err, opt.KonfigPath)
		}
	}

	autoSaveChan := make(chan os.Signal, 1)

	go func() {
		<-autoSaveChan
		mngr.Save()
	}()

	if opt.AutoSave {
		signal.Notify(autoSaveChan, syscall.SIGINT, syscall.SIGTERM)
	}

	return mngr, nil
}

func (m *KonfigManager) RegisterSection(section KonfigSection) error {
	if section == nil {
		return fmt.Errorf("cannot register nil section")
	}

	if reflect.TypeOf(section).Kind() != reflect.Ptr || reflect.TypeOf(section).Elem().Kind() != reflect.Struct {
		return fmt.Errorf("invalid section type: section must be a pointer to a struct that implements KonfigSection. Got: %T", section)
	}

	sectionName := section.Name()
	if _, exists := m.sections[sectionName]; exists {
		return fmt.Errorf("section name conflict: '%s' is already registered. Each section must have a unique name", sectionName)
	}

	m.sections[sectionName] = section
	return nil
}

// simple shorthand for registering new sections directly from structs
func (m *KonfigManager) AddSimpleSection(sectionName string, data interface{}) error {
	return m.RegisterSection(NewKonfigSection(data, WithSectionName(func() string { return sectionName })))
}

func (m *KonfigManager) Load() error {
	filePath := m.opts.KonfigPath
	fmtHandler := m.formatHnadler

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("configuration file not found: %s does not exist", filePath)
		}
		return fmt.Errorf("configuration file can not be read: %s", filePath)
	}

	// fuck yaml for being different like this, basically this whole function needs to be duplicated to acomodate yaml
	var configData interface{}
	if m.opts.Format == YAML {
		// YAML needs map[interface{}]interface{}
		yamlData := make(map[interface{}]interface{})
		if len(data) > 0 {
			if err := fmtHandler.Unmarshal(data, &yamlData); err != nil {
				return fmt.Errorf("error unmarshalling YAML configuration data: %v", err)
			}
		}
		configData = yamlData
	} else {
		// JSON and TOML can use map[string]interface{}
		stringData := make(map[string]interface{})
		if len(data) > 0 {
			if err := fmtHandler.Unmarshal(data, &stringData); err != nil {
				return fmt.Errorf("error unmarshalling configuration data: %v", err)
			}
		}
		configData = stringData
	}

	// Process sections based on format
	if m.opts.Format == YAML {
		yamlConfig := configData.(map[interface{}]interface{})
		for sectionName, section := range m.sections {
			var sectionDataFromConfig interface{}
			var sectionExists bool

			for k, v := range yamlConfig {
				if strKey, ok := k.(string); ok && strKey == sectionName {
					sectionDataFromConfig = v
					sectionExists = true
					break
				}
			}

			if !sectionExists {
				fmt.Printf("Section %s not found in configuration file, skipping\n", sectionName)
				continue
			} else {
				sectionImpl, ok := section.(*konfigSectionImpl)
				if !ok {
					return fmt.Errorf("registered section %s is not a *konfigSectionImpl, invalid registration", sectionName)
				}
				dataStructPtr := sectionImpl.data

				mapData, ok := sectionDataFromConfig.(map[interface{}]interface{})
				if !ok {
					return fmt.Errorf("YAML section data is not map[interface{}]interface{}")
				}

				sectionBytes, err := fmtHandler.Marshal(mapData)
				if err != nil {
					return fmt.Errorf("error marshalling section data into specific data type: %v, err: %v", reflect.TypeOf(dataStructPtr).Name(), err)
				}

				if err := fmtHandler.Unmarshal(sectionBytes, dataStructPtr); err != nil {
					return fmt.Errorf("error unmarshalling section data into specific data type: %v, err: %v", reflect.TypeOf(dataStructPtr).Name(), err)
				}
			}

			if m.opts.UseCallbacks {
				if err := section.Validate(); err != nil {
					return fmt.Errorf("error validating section %s: %v", sectionName, err)
				}

				if err := section.OnLoad(); err != nil {
					return fmt.Errorf("error running onload action for section %s: %v", sectionName, err)
				}
			}
		}
	} else {
		stringConfig := configData.(map[string]interface{})
		for sectionName, section := range m.sections {
			sectionDataFromConfig, sectionExists := stringConfig[sectionName]

			if !sectionExists {
				fmt.Printf("Section %s not found in configuration file, skipping\n", sectionName)
				continue
			} else {
				sectionImpl, ok := section.(*konfigSectionImpl)
				if !ok {
					return fmt.Errorf("registered section %s is not a *konfigSectionImpl, invalid registration", sectionName)
				}
				dataStructPtr := sectionImpl.data

				mapData, ok := sectionDataFromConfig.(map[string]interface{})
				if !ok {
					return fmt.Errorf("%v section data is not map[string]interface{}", m.opts.Format)
				}

				sectionBytes, err := fmtHandler.Marshal(mapData)
				if err != nil {
					return fmt.Errorf("error marshalling section data into specific data type: %v, err: %v", reflect.TypeOf(dataStructPtr).Name(), err)
				}

				if err := fmtHandler.Unmarshal(sectionBytes, dataStructPtr); err != nil {
					return fmt.Errorf("error unmarshalling section data into specific data type: %v, err: %v", reflect.TypeOf(dataStructPtr).Name(), err)
				}
			}

			if m.opts.UseCallbacks {
				if err := section.Validate(); err != nil {
					return fmt.Errorf("error validating section %s: %v", sectionName, err)
				}

				if err := section.OnLoad(); err != nil {
					return fmt.Errorf("error running onload action for section %s: %v", sectionName, err)
				}
			}
		}
	}

	return nil
}

func (m *KonfigManager) Save() error {
	filePath := m.opts.KonfigPath
	fmtHandler := m.formatHnadler

	configData := make(map[string]interface{})

	for sectionName, section := range m.sections {
		sectionImpl, ok := section.(*konfigSectionImpl)
		if !ok {
			return fmt.Errorf("registered section %s is not a *konfigSectionImpl, invalid registration", sectionName)
		}
		configData[sectionName] = sectionImpl.data
	}

	data, err := fmtHandler.Marshal(configData)
	if err != nil {
		return fmt.Errorf("error marshalling configuration data: %v", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("error writing configuration file: %v", err)
	}

	return nil
}
