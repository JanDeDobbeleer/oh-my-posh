package sqlite

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testdata(t *testing.T) []byte {
	t.Helper()

	data, err := os.ReadFile("testdata/access_tokens.db")
	require.NoError(t, err)

	return data
}

func TestFindRowShortValue(t *testing.T) {
	data := testdata(t)

	row, found, err := FindRow(data, "access_tokens", "account_id", "user@example.com")
	require.NoError(t, err)
	require.True(t, found)

	assert.Equal(t, "ya29.short-token", row["access_token"])
	assert.Equal(t, "2024-01-01 12:00:00.000000", row["token_expiry"])
	assert.Equal(t, "", row["rapt_token"])
}

func TestFindRowOverflowValue(t *testing.T) {
	data := testdata(t)

	row, found, err := FindRow(data, "access_tokens", "account_id", "big@example.com")
	require.NoError(t, err)
	require.True(t, found)

	assert.Equal(t, strings.Repeat("x", 5000), row["access_token"])
	assert.Equal(t, "2099-01-01 00:00:00.000000", row["token_expiry"])
	assert.Equal(t, "rapt-abc", row["rapt_token"])
}

func TestFindRowNotFound(t *testing.T) {
	data := testdata(t)

	_, found, err := FindRow(data, "access_tokens", "account_id", "missing@example.com")
	require.NoError(t, err)
	assert.False(t, found)
}

func TestFindRowUnknownTable(t *testing.T) {
	data := testdata(t)

	_, _, err := FindRow(data, "does_not_exist", "account_id", "user@example.com")
	assert.Error(t, err)
}

func TestFindRowUnknownColumn(t *testing.T) {
	data := testdata(t)

	_, _, err := FindRow(data, "access_tokens", "nope", "user@example.com")
	assert.Error(t, err)
}

func TestFindRowNotASQLiteFile(t *testing.T) {
	_, _, err := FindRow([]byte("not a database"), "access_tokens", "account_id", "user@example.com")
	assert.Error(t, err)
}

func TestParseColumns(t *testing.T) {
	cases := []struct {
		Case     string
		SQL      string
		Expected []string
	}{
		{
			Case:     "simple",
			SQL:      "CREATE TABLE t (a TEXT, b INTEGER)",
			Expected: []string{"a", "b"},
		},
		{
			Case:     "primary key constraint skipped",
			SQL:      "CREATE TABLE access_tokens (account_id TEXT PRIMARY KEY, access_token TEXT, token_expiry TIMESTAMP, rapt_token TEXT, id_token TEXT)",
			Expected: []string{"account_id", "access_token", "token_expiry", "rapt_token", "id_token"},
		},
		{
			Case:     "quoted identifiers",
			SQL:      `CREATE TABLE t ("a" TEXT, [b] INTEGER)`,
			Expected: []string{"a", "b"},
		},
		{
			Case:     "nested parens in type and table constraint",
			SQL:      "CREATE TABLE t (a DECIMAL(10,2), b TEXT, PRIMARY KEY (a, b))",
			Expected: []string{"a", "b"},
		},
	}

	for _, tc := range cases {
		columns, err := parseColumns(tc.SQL)
		require.NoError(t, err, tc.Case)
		assert.Equal(t, tc.Expected, columns, tc.Case)
	}
}

func TestReadVarint(t *testing.T) {
	cases := []struct {
		Case     string
		Buf      []byte
		Expected int64
		Used     int
	}{
		{Case: "single byte", Buf: []byte{0x05}, Expected: 5, Used: 1},
		{Case: "two bytes", Buf: []byte{0x81, 0x01}, Expected: 129, Used: 2},
		{Case: "zero", Buf: []byte{0x00}, Expected: 0, Used: 1},
	}

	for _, tc := range cases {
		v, used := readVarint(tc.Buf)
		assert.Equal(t, tc.Expected, v, tc.Case)
		assert.Equal(t, tc.Used, used, tc.Case)
	}
}
