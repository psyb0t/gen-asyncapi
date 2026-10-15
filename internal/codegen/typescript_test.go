package codegen

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeSpec(t *testing.T, spec string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "asyncapi.yaml")
	require.NoError(t, os.WriteFile(path, []byte(spec), 0o600))

	return path
}

func TestTSIndent(t *testing.T) {
	t.Parallel()

	lines := []string{
		"switch (x) {",
		`case "{": {`,
		"rtOneOf(path, [",
		"(): void => {",
		"},",
		"]);",
		"}",
		"}",
		"done();",
	}

	want := []string{
		"  switch (x) {",
		`    case "{": {`,
		"      rtOneOf(path, [",
		"        (): void => {",
		"        },",
		"      ]);",
		"    }",
		"  }",
		"  done();",
	}

	assert.Equal(t, want, tsIndent(lines, 1))
}

func TestTSBraces(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		line              string
		wantOpens         int
		wantLeadingCloses int
		wantCloses        int
	}{
		{line: "if (x) {", wantOpens: 2, wantCloses: 1},
		{line: "}", wantLeadingCloses: 1, wantCloses: 1},
		{line: "]);", wantLeadingCloses: 2, wantCloses: 2},
		{line: `rtPattern(p, v, "^[a-z]{2}$");`, wantOpens: 1, wantCloses: 1},
		{line: `x("\"{");`, wantOpens: 1, wantCloses: 1},
	}

	for _, tc := range testCases {
		t.Run(tc.line, func(t *testing.T) {
			t.Parallel()

			opens, leadingCloses, closes := tsBraces(tc.line)
			assert.Equal(t, tc.wantOpens, opens)
			assert.Equal(t, tc.wantLeadingCloses, leadingCloses)
			assert.Equal(t, tc.wantCloses, closes)
		})
	}
}

func TestTSDoc(t *testing.T) {
	t.Parallel()

	assert.Empty(t, tsDoc(nil, ""))
	assert.Equal(
		t,
		"  /**\n   * a *\\/ b\n   *\n   */\n",
		tsDoc([]string{"a */ b", ""}, "  "),
	)
}

func TestTSLiterals(t *testing.T) {
	t.Parallel()

	assert.Equal(t, `"<a&b>"`, tsQuote("<a&b>"))
	assert.Equal(t, `"tab\there"`, tsFromGoString(`"tab\there"`))
	assert.Equal(t, "not-quoted", tsFromGoString("not-quoted"))
	assert.Equal(t, `"x"`, tsFromGoLiteral(`"x"`))
	assert.Equal(t, "-1.5", tsFromGoLiteral("-1.5"))
}

func TestTSNames(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "room", tsPropertyName("room"))
	assert.Equal(t, `"x-trace-id"`, tsPropertyName("x-trace-id"))
	assert.Equal(t, `"2fa"`, tsPropertyName("2fa"))
	assert.Equal(t, "classValue", tsParamName("class"))
	assert.Equal(t, "userID", tsParamName("userID"))
}

func TestTSType(t *testing.T) {
	t.Parallel()

	stringRef := &typeRef{kind: refPrimitive, jsonType: jsonTypeString}

	testCases := []struct {
		name string
		ref  *typeRef
		want string
	}{
		{name: "string", ref: stringRef, want: "string"},
		{
			name: "integer",
			ref:  &typeRef{kind: refPrimitive, jsonType: jsonTypeInteger},
			want: "number",
		},
		{
			name: "untyped primitive",
			ref:  &typeRef{kind: refPrimitive},
			want: "unknown",
		},
		{
			name: "array",
			ref:  &typeRef{kind: refArray, elem: stringRef},
			want: "Array<string>",
		},
		{
			name: "map",
			ref:  &typeRef{kind: refMap, elem: stringRef},
			want: "Record<string, string>",
		},
		{
			name: "object without properties",
			ref:  &typeRef{kind: refAny, expr: goTypeAnyMap},
			want: "Record<string, unknown>",
		},
		{
			name: "raw JSON",
			ref:  &typeRef{kind: refAny, expr: goTypeRawJSON},
			want: "unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, tsType(tc.ref))
		})
	}
}

func TestTSRuntimeIncludesOnlyWhatIsCalled(t *testing.T) {
	t.Parallel()

	runtime, err := tsRuntime([]byte("rtDateTime(value, path); rtNope();"))
	require.NoError(t, err)

	assert.Contains(t, runtime, "function rtDateTime(")
	assert.Contains(t, runtime, "function rtString(", "a dependency")
	assert.Contains(t, runtime, "function rtFail(", "a dependency")
	assert.NotContains(t, runtime, "function rtArray(")
	assert.NotContains(t, runtime, runtimeMarker)

	empty, err := tsRuntime([]byte("nothing here"))
	require.NoError(t, err)
	assert.Empty(t, empty)
}

func TestGenerateTypeScriptModelsOnly(t *testing.T) {
	t.Parallel()

	config := &Config{
		Language: LanguageTypeScript,
		Generate: &GenerateOptions{Models: true},
	}
	require.NoError(t, config.Validate())

	generated, err := Generate(config, chatSpec())
	require.NoError(t, err)

	source := string(generated)
	assert.Contains(t, source, "export interface JoinRequest {")
	assert.Contains(t, source, "export function validateJoinRequest(")
	assert.NotContains(t, source, "export class Client")
	assert.NotContains(t, source, "export interface Handler")
	assert.NotContains(t, source, "UnknownMessageError")
}

func TestGenerateTypeScriptReservesItsNames(t *testing.T) {
	t.Parallel()

	spec := writeSpec(t, `asyncapi: 3.0.0
info: {title: t, version: "1"}
components:
  schemas:
    Client:
      type: object
      properties:
        class: {type: string}
    Ratio:
      type: number
      enum: [0.5, 1]
`)

	config := &Config{Language: LanguageTypeScript}
	require.NoError(t, config.Validate())

	generated, err := Generate(config, spec)
	require.NoError(t, err)

	source := string(generated)
	assert.Contains(t, source, "export interface Client2 {")
	assert.Contains(t, source, "export type Ratio = 0.5 | 1;")
}
