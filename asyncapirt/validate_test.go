package asyncapirt_test

import (
	"encoding/json"
	"testing"

	"github.com/psyb0t/ctxerrors/commerr"
	"github.com/psyb0t/gen-asyncapi/asyncapirt"
	"github.com/stretchr/testify/require"
)

func TestChecks(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		check   func() error
		wantErr bool
	}{
		{
			name: "minLength counts characters",
			check: func() error {
				return asyncapirt.CheckMinLength("f", "éé", 2)
			},
		},
		{
			name: "minLength short",
			check: func() error {
				return asyncapirt.CheckMinLength("f", "a", 2)
			},
			wantErr: true,
		},
		{
			name: "maxLength counts characters",
			check: func() error {
				return asyncapirt.CheckMaxLength("f", "ééé", 3)
			},
		},
		{
			name: "maxLength long",
			check: func() error {
				return asyncapirt.CheckMaxLength("f", "abcd", 3)
			},
			wantErr: true,
		},
		{
			name: "pattern match",
			check: func() error {
				return asyncapirt.CheckPattern("f", "ab12", "^[a-z]+[0-9]+$")
			},
		},
		{
			name: "pattern miss",
			check: func() error {
				return asyncapirt.CheckPattern("f", "12ab", "^[a-z]+$")
			},
			wantErr: true,
		},
		{
			name: "pattern invalid",
			check: func() error {
				return asyncapirt.CheckPattern("f", "x", "(")
			},
			wantErr: true,
		},
		{
			name:  "minimum equal",
			check: func() error { return asyncapirt.CheckMinimum("f", 1, 1) },
		},
		{
			name: "minimum below",
			check: func() error {
				return asyncapirt.CheckMinimum("f", 0.5, 1)
			},
			wantErr: true,
		},
		{
			name:  "maximum equal",
			check: func() error { return asyncapirt.CheckMaximum("f", 1, 1) },
		},
		{
			name:    "maximum above",
			check:   func() error { return asyncapirt.CheckMaximum("f", 2, 1) },
			wantErr: true,
		},
		{
			name: "exclusive minimum equal",
			check: func() error {
				return asyncapirt.CheckExclusiveMinimum("f", 1, 1)
			},
			wantErr: true,
		},
		{
			name: "exclusive minimum above",
			check: func() error {
				return asyncapirt.CheckExclusiveMinimum("f", 1.1, 1)
			},
		},
		{
			name: "exclusive maximum equal",
			check: func() error {
				return asyncapirt.CheckExclusiveMaximum("f", 1, 1)
			},
			wantErr: true,
		},
		{
			name: "exclusive maximum below",
			check: func() error {
				return asyncapirt.CheckExclusiveMaximum("f", 0.9, 1)
			},
		},
		{
			name:  "minItems ok",
			check: func() error { return asyncapirt.CheckMinItems("f", 2, 2) },
		},
		{
			name: "minItems short",
			check: func() error {
				return asyncapirt.CheckMinItems("f", 1, 2)
			},
			wantErr: true,
		},
		{
			name:  "maxItems ok",
			check: func() error { return asyncapirt.CheckMaxItems("f", 2, 2) },
		},
		{
			name: "maxItems long",
			check: func() error {
				return asyncapirt.CheckMaxItems("f", 3, 2)
			},
			wantErr: true,
		},
		{
			name: "multipleOf exact",
			check: func() error {
				return asyncapirt.CheckMultipleOf("f", 9.99, 0.01)
			},
		},
		{
			name: "multipleOf integer",
			check: func() error {
				return asyncapirt.CheckMultipleOf("f", 15, 5)
			},
		},
		{
			name: "multipleOf off",
			check: func() error {
				return asyncapirt.CheckMultipleOf("f", 9.999, 0.01)
			},
			wantErr: true,
		},
		{
			name: "multipleOf zero divisor",
			check: func() error {
				return asyncapirt.CheckMultipleOf("f", 1, 0)
			},
			wantErr: true,
		},
		{
			name: "uniqueItems ok",
			check: func() error {
				return asyncapirt.CheckUniqueItems("f", []string{"a", "b"})
			},
		},
		{
			name: "uniqueItems duplicate",
			check: func() error {
				return asyncapirt.CheckUniqueItems("f", []int{1, 2, 1})
			},
			wantErr: true,
		},
		{
			name: "uniqueItems compares objects by JSON",
			check: func() error {
				return asyncapirt.CheckUniqueItems(
					"f",
					[]map[string]int{{"a": 1}, {"a": 1}},
				)
			},
			wantErr: true,
		},
		{
			name: "uniqueItems unencodable",
			check: func() error {
				return asyncapirt.CheckUniqueItems("f", []any{make(chan int)})
			},
			wantErr: true,
		},
		{
			name: "minProperties ok",
			check: func() error {
				return asyncapirt.CheckMinProperties("f", 2, 2)
			},
		},
		{
			name: "minProperties short",
			check: func() error {
				return asyncapirt.CheckMinProperties("f", 1, 2)
			},
			wantErr: true,
		},
		{
			name: "maxProperties ok",
			check: func() error {
				return asyncapirt.CheckMaxProperties("f", 2, 2)
			},
		},
		{
			name: "maxProperties long",
			check: func() error {
				return asyncapirt.CheckMaxProperties("f", 3, 2)
			},
			wantErr: true,
		},
		{
			name:  "const equal",
			check: func() error { return asyncapirt.CheckConst("f", "a", "a") },
		},
		{
			name: "const different",
			check: func() error {
				return asyncapirt.CheckConst("f", int64(1), 2)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.check()
			if tc.wantErr {
				require.ErrorIs(t, err, commerr.ErrValidationFailed)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestObjectKeyChecks(t *testing.T) {
	t.Parallel()

	members, err := asyncapirt.DecodeObject([]byte(`{"a":1,"b":2}`), "T")
	require.NoError(t, err)

	require.NoError(t, asyncapirt.RequireKeys("T", members, "a", "b"))
	require.ErrorIs(
		t,
		asyncapirt.RequireKeys("T", members, "c"),
		commerr.ErrRequiredFieldNotSet,
	)

	require.NoError(
		t,
		asyncapirt.RejectUnknownKeys("T", members, "a", "b", "c"),
	)
	require.ErrorIs(
		t,
		asyncapirt.RejectUnknownKeys("T", members, "a"),
		commerr.ErrValidationFailed,
	)

	require.NoError(t, asyncapirt.RequireDependent("T", members, "a", "b"))
	require.NoError(
		t,
		asyncapirt.RequireDependent("T", members, "missing", "c"),
	)
	require.ErrorIs(
		t,
		asyncapirt.RequireDependent("T", members, "a", "c"),
		commerr.ErrRequiredFieldNotSet,
	)

	_, err = asyncapirt.DecodeObject([]byte(`null`), "T")
	require.ErrorIs(t, err, commerr.ErrCouldNotDecode)

	_, err = asyncapirt.DecodeObject([]byte(`[1]`), "T")
	require.ErrorIs(t, err, commerr.ErrCouldNotDecode)
}

func TestValidateValue(t *testing.T) {
	t.Parallel()

	require.NoError(t, asyncapirt.ValidateValue(json.RawMessage(`1`)))
	require.ErrorIs(
		t,
		asyncapirt.ValidateValue(alwaysInvalid{}),
		commerr.ErrValidationFailed,
	)
}
