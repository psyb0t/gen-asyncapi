package codegen

import (
	"bytes"
	"os"

	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/ctxerrors/commerr"
	"gopkg.in/yaml.v3"
)

const defaultOutput = "asyncapi.gen.go"

// Config is the generator configuration file. Its layout follows
// oapi-codegen's: the spec is a command-line argument, and paths are relative
// to the directory the generator runs in.
type Config struct {
	// Package is the Go package name of the generated file.
	Package string `yaml:"package"`
	// Output is the generated file. Empty means asyncapi.gen.go.
	Output string `yaml:"output"`
	// Generate picks what to generate. Omitted, everything is generated.
	Generate *GenerateOptions `yaml:"generate"`
	// OutputOptions tunes the generated code.
	//nolint:tagliatelle // kebab-case keys follow oapi-codegen
	OutputOptions OutputOptions `yaml:"output-options"`
}

// GenerateOptions picks the parts of the generated file.
type GenerateOptions struct {
	// Models generates the schema types, codecs and validation.
	Models bool `yaml:"models"`
	// Client generates message decoding, the Handler interface, Dispatch and
	// the Client. It needs Models.
	Client bool `yaml:"client"`
}

// OutputOptions tunes the generated code.
type OutputOptions struct {
	// MessageMatchers tells incoming messages apart when the spec gives no
	// const or discriminator to do it. It maps a message key to JSON pointers
	// (RFC 6901) into the payload and the value each must hold.
	//nolint:tagliatelle // kebab-case keys follow oapi-codegen
	MessageMatchers map[string]map[string]any `yaml:"message-matchers"`
}

// LoadConfig reads and checks the configuration file at path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, ctxerrors.Wrapf(err, "read config %s", path)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	config := &Config{}
	if err := decoder.Decode(config); err != nil {
		return nil, ctxerrors.Wrapf(
			commerr.ErrParseFailed,
			"decode config %s: %v",
			path,
			err,
		)
	}

	return config, nil
}

// Validate fills defaults and checks required settings. Call it after any
// command-line overrides are applied.
func (c *Config) Validate() error {
	if c.Package == "" {
		return ctxerrors.Wrap(commerr.ErrRequiredConfigValueNotSet, "package")
	}

	if c.Output == "" {
		c.Output = defaultOutput
	}

	if c.Generate == nil {
		c.Generate = &GenerateOptions{Models: true, Client: true}
	}

	if !c.Generate.Models && !c.Generate.Client {
		return ctxerrors.Wrap(
			commerr.ErrInvalidValue,
			"generate: nothing selected, enable models or client",
		)
	}

	if c.Generate.Client && !c.Generate.Models {
		return ctxerrors.Wrap(
			commerr.ErrInvalidValue,
			"generate: client needs models",
		)
	}

	return nil
}
