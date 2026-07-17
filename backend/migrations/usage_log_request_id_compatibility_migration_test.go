package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigration183RestoresUsageLogRequestIDCompatibility(t *testing.T) {
	content, err := FS.ReadFile("183_fix_usage_log_request_id_compatibility.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "ALTER TABLE usage_logs ALTER COLUMN client_request_id DROP NOT NULL")
	require.Contains(t, sql, "ALTER TABLE usage_logs ALTER COLUMN request_id DROP NOT NULL")
	require.NotContains(t, sql, "ALTER TABLE request_logs")
}
