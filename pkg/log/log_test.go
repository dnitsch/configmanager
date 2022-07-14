package log

import (
	"bytes"
	"testing"

	"github.com/dnitsch/configmanager/internal/testutils"
	"github.com/rs/zerolog"
)

func Test_LogInfo(t *testing.T) {
	tests := []struct {
		name    string
		level   zerolog.Level
		message string
		expect  string
	}{
		{
			name:    "info at debug",
			level:   zerolog.DebugLevel,
			message: "write me out...",
			expect: `{"level":"info","message":"write me out..."}
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// overwrite writer for testing
			var buf bytes.Buffer
			// overwrite logger for testing without timestamp
			Logger = zerolog.New(&buf).With().Logger().Level(tt.level)
			Info(tt.message)
			s := buf.String()
			if s != tt.expect {
				t.Errorf(testutils.TestPhrase, tt.expect, s)
			}
		})
	}
}
