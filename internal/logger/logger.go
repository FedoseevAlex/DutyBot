package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

const logFilePermissions = 0o666

var Log zerolog.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

func InitLogger(LogPath string) {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// Unfortunately i did't find good way to close log file after use.
	// Log file could be closed by GC: https://pkg.go.dev/os#File.Fd
	logfile, err := os.OpenFile(LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY|os.O_SYNC, logFilePermissions)
	if err != nil {
		Log.Error().
			Err(err).
			Msgf("Unable to open a log file: %s. Writing to STDOUT.", LogPath)
	} else {
		Log = Log.Output(logfile).With().Timestamp().Logger()
	}
}

func GetConsoleLogger() zerolog.Logger {
	return zerolog.New(zerolog.NewConsoleWriter())
}
