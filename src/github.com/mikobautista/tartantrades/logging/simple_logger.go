package logger

import (
    "fmt"
    "log"
    "os"
    "runtime/debug"
)

const (
    name        = "log-proto.txt"
    flag        = os.O_RDWR | os.O_CREATE
    perm        = os.FileMode(0666)
    LOG_TO_FILE = false
)

type Logger struct {
    VERBOSE bool
    logv    *log.Logger
    loge    *log.Logger
    file    *os.File
}

func NewLogger(verbose bool) *Logger {
    if LOG_TO_FILE {
        file, err := os.OpenFile(name, flag, perm)
        if err != nil {
            os.Exit(10)
        }

        return &Logger{
            VERBOSE: verbose,
            file:    file,
            logv:    log.New(file, "VERBOSE ", log.Lshortfile|log.Lmicroseconds),
            loge:    log.New(file, "ERROR ", log.Lshortfile|log.Lmicroseconds),
        }

    }

    return &Logger{
        VERBOSE: verbose,
        logv:    log.New(os.Stdout, "VERBOSE ", log.Lmicroseconds),
        loge:    log.New(os.Stderr, "ERROR ", log.Lmicroseconds),
    }
}

var ST = debug.PrintStack

func (logger *Logger) LogVerbose(format string, v ...interface{}) {
    if logger.VERBOSE {
        logger.logv.Output(2, fmt.Sprintf(format, v...))
    }
}

func (logger *Logger) LogError(format string, v ...interface{}) {
    logger.loge.Output(2, fmt.Sprintf(format, v...))

}

func (logger *Logger) CheckForError(err error, shouldClose bool) {
    if err != nil {
        logger.loge.Printf("Error: %s", err.Error())
        debug.PrintStack()
        if shouldClose {
            os.Exit(1)
        }
    }
}

func (logger *Logger) closeFile() {
    if LOG_TO_FILE {
        logger.file.Close()
    }
}
