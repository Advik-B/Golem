package server

import (
    "go.uber.org/zap"
)

var Log *zap.Logger

func InitLogger(development bool) {
    var err error
    if development {
        Log, err = zap.NewDevelopment()
    } else {
        Log, err = zap.NewProduction()
    }
    if err != nil {
        panic("failed to initialize logger: " + err.Error())
    }
}