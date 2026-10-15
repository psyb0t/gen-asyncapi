package codegen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/psyb0t/ctxerrors/commerr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const goNameInstrumentRates = "InstrumentRates"

func TestGoName(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input string
		want  string
	}{
		{input: "instrument_rates", want: goNameInstrumentRates},
		{input: "instrumentRates", want: goNameInstrumentRates},
		{input: "instrument-rates", want: goNameInstrumentRates},
		{input: "userId", want: "UserID"},
		{input: "HTTPServer", want: "HTTPServer"},
		{input: "message.posted", want: "MessagePosted"},
		{input: "2fa", want: "N2fa"},
		{input: "wss", want: "WSS"},
		{input: "", want: "Value"},
		{input: "---", want: "Value"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, GoName(tc.input))
		})
	}
}

func TestLowerFirst(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input string
		want  string
	}{
		{input: "InstrumentID", want: "instrumentID"},
		{input: "ID", want: "id"},
		{input: "Type", want: "typeValue"},
		{input: "", want: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, lowerFirst(tc.input))
		})
	}
}

func TestNameSetClaim(t *testing.T) {
	t.Parallel()

	names := newNameSet("Client")
	assert.Equal(t, "Client2", names.claim("Client"))
	assert.Equal(t, "Room", names.claim("Room"))
	assert.Equal(t, "Room2", names.claim("Room"))
	assert.Equal(t, "Room3", names.claim("Room"))
}

// TestGeneratedExamplesAreCurrent regenerates every example and compares it
// with the committed output, so a generator change that alters output fails
// until the examples are regenerated and their tests rerun.
func TestGeneratedExamplesAreCurrent(t *testing.T) {
	t.Parallel()

	configs, err := filepath.Glob(
		filepath.Join("..", "..", "examples", "*", "config.yaml"),
	)
	require.NoError(t, err)

	official, err := filepath.Glob(
		filepath.Join("..", "..", "examples", "official", "*", "config.yaml"),
	)
	require.NoError(t, err)

	configs = append(configs, official...)
	require.GreaterOrEqual(t, len(configs), 4)

	for _, configPath := range configs {
		dir := filepath.Dir(configPath)

		for _, run := range generateRuns(t, dir) {
			name := filepath.Base(dir) + "/" + string(run.language)

			t.Run(name, func(t *testing.T) {
				t.Parallel()

				config, err := LoadConfig(filepath.Join(dir, run.config))
				require.NoError(t, err)

				if run.language != "" {
					config.Language = run.language
				}

				if run.output != "" {
					config.Output = run.output
				}

				require.NoError(t, config.Validate())

				generated, err := Generate(config, filepath.Join(dir, run.spec))
				require.NoError(t, err)

				committed, err := os.ReadFile(filepath.Join(dir, config.Output))
				require.NoError(t, err)

				assert.Equal(t, string(committed), string(generated))
			})
		}
	}
}

func chatSpec() string {
	return filepath.Join("..", "..", "examples", "chat", "chat.yaml")
}

// generateRun is the arguments of one go:generate line of an example.
type generateRun struct {
	config   string
	language Language
	output   string
	spec     string
}

// generateRuns returns every go:generate line of the example's gen.go, so the
// test regenerates exactly what go generate does.
func generateRuns(t *testing.T, dir string) []generateRun {
	t.Helper()

	source, err := os.ReadFile(filepath.Join(dir, "gen.go"))
	require.NoError(t, err)

	var runs []generateRun

	for line := range strings.SplitSeq(string(source), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != "//go:generate" {
			continue
		}

		// Skips "//go:generate go tool gen-asyncapi".
		const commandWords = 4

		var run generateRun

		for _, argument := range fields[commandWords:] {
			if value, ok := strings.CutPrefix(argument, "--config="); ok {
				run.config = value

				continue
			}

			if value, ok := strings.CutPrefix(argument, "-language="); ok {
				run.language = Language(value)

				continue
			}

			if value, ok := strings.CutPrefix(argument, "-o="); ok {
				run.output = value

				continue
			}

			run.spec = argument
		}

		runs = append(runs, run)
	}

	require.NotEmpty(t, runs, "no go:generate line in %s", dir)

	return runs
}

func writeConfig(t *testing.T, config string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(config), 0o600))

	return path
}

func TestConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		config       string
		wantLoadErr  error
		wantErr      error
		wantOutput   string
		wantGenerate GenerateOptions
	}{
		{
			name:         "defaults output and generates everything",
			config:       "package: chat\n",
			wantOutput:   defaultOutput,
			wantGenerate: GenerateOptions{Models: true, Client: true},
		},
		{
			name:         "models only",
			config:       "package: chat\noutput: x.go\ngenerate:\n  models: true\n",
			wantOutput:   "x.go",
			wantGenerate: GenerateOptions{Models: true},
		},
		{
			name:    "missing package",
			config:  "output: x.go\n",
			wantErr: commerr.ErrRequiredConfigValueNotSet,
		},
		{
			name:    "nothing generated",
			config:  "package: chat\ngenerate: {}\n",
			wantErr: commerr.ErrInvalidValue,
		},
		{
			name:    "client without models",
			config:  "package: chat\ngenerate:\n  client: true\n",
			wantErr: commerr.ErrInvalidValue,
		},
		{
			name:        "unknown key",
			config:      "package: chat\nbogus: 1\n",
			wantLoadErr: commerr.ErrParseFailed,
		},
		{
			name:         "typescript needs no package",
			config:       "language: typescript\n",
			wantOutput:   defaultTypeScriptOutput,
			wantGenerate: GenerateOptions{Models: true, Client: true},
		},
		{
			name:         "explicit go",
			config:       "language: go\npackage: chat\n",
			wantOutput:   defaultOutput,
			wantGenerate: GenerateOptions{Models: true, Client: true},
		},
		{
			name:    "unknown language",
			config:  "language: rust\npackage: chat\n",
			wantErr: commerr.ErrInvalidValue,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config, err := LoadConfig(writeConfig(t, tc.config))
			if tc.wantLoadErr != nil {
				require.ErrorIs(t, err, tc.wantLoadErr)

				return
			}

			require.NoError(t, err)

			err = config.Validate()
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantOutput, config.Output)
			assert.Equal(t, tc.wantGenerate, *config.Generate)
		})
	}

	_, err := LoadConfig(filepath.Join(t.TempDir(), "missing.yaml"))
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestGenerateRejectsUnknownMatcherMessage(t *testing.T) {
	t.Parallel()

	config, err := LoadConfig(
		writeConfig(
			t,
			"package: chat\noutput-options:\n  message-matchers:\n    Nope:\n      /a: b\n",
		),
	)
	require.NoError(t, err)
	require.NoError(t, config.Validate())

	_, err = Generate(config, chatSpec())
	require.ErrorIs(t, err, commerr.ErrNotFound)
}

func TestGenerateRejectsEmptyMatcherList(t *testing.T) {
	t.Parallel()

	config, err := LoadConfig(
		writeConfig(
			t,
			"package: chat\noutput-options:\n  message-matchers:\n    Notice:\n      /messages/0/type: []\n",
		),
	)
	require.NoError(t, err)
	require.NoError(t, config.Validate())

	_, err = Generate(config, chatSpec())
	require.ErrorIs(t, err, commerr.ErrInvalidValue)
}

func TestGenerateModelsOnly(t *testing.T) {
	t.Parallel()

	config := &Config{Package: "chat", Generate: &GenerateOptions{Models: true}}
	require.NoError(t, config.Validate())

	generated, err := Generate(config, chatSpec())
	require.NoError(t, err)
	assert.Contains(t, string(generated), "type JoinRequest struct")
	assert.NotContains(t, string(generated), "type Client struct")
	assert.NotContains(t, string(generated), "type Handler interface")
}

func TestGenerateFailsOnMissingSpec(t *testing.T) {
	t.Parallel()

	config := &Config{Package: "x"}
	require.NoError(t, config.Validate())

	_, err := Generate(config, filepath.Join(t.TempDir(), "nope.yaml"))
	require.ErrorIs(t, err, os.ErrNotExist)
}
