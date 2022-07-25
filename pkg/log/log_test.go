package log

import (
	"bytes"
	"testing"

	"github.com/dnitsch/configmanager/internal/testutils"
	"github.com/rs/zerolog"
)

func Test_LogInfo(t *testing.T) {
	tests := []struct {
		name      string
		level     zerolog.Level
		logMethod func(msg string)
		message   string
		expect    string
	}{
		{
			name:      "info at debug",
			level:     zerolog.DebugLevel,
			logMethod: Info,
			message:   "write me out...",
			expect: `{"level":"info","message":"write me out..."}
`,
		},
		{
			name:      "warn at debug",
			level:     zerolog.DebugLevel,
			logMethod: Warn,
			message:   "write me out...",
			expect: `{"level":"warn","message":"write me out..."}
`,
		},
		{
			name:      "debug at debug",
			level:     zerolog.DebugLevel,
			logMethod: Debug,
			message:   "write me out...",
			expect: `{"level":"debug","message":"write me out..."}
`,
		},
		{
			name:      "debug at info",
			level:     zerolog.InfoLevel,
			logMethod: Debug,
			message:   "write me out...",
			expect:    ``,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// overwrite writer for testing
			var buf bytes.Buffer
			// overwrite logger for testing without timestamp
			Logger = zerolog.New(&buf).With().Logger().Level(tt.level)
			tt.logMethod(tt.message)
			s := buf.String()
			if s != tt.expect {
				t.Errorf(testutils.TestPhrase, tt.expect, s)
			}
		})
	}
}
