package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	ProjectID string
	Dir       string
	Debug     bool
	Verbose   bool
}

func Run() error {
	ctx := context.TODO()

	cfg, err := NewConfig()
	if err != nil {
		return errors.WithStack(err)
	}

	l, err := NewLogger(cfg)
	if err != nil {
		return errors.WithStack(err)
	}
	defer l.Sync()

	zap.ReplaceGlobals(l)
	cfgJSON, _ := json.Marshal(cfg)
	zap.L().Debug("config", zap.String("config", string(cfgJSON)))

	cmd, err := InitializeCmd(ctx, cfg)
	if err != nil {
		return errors.WithStack(err)
	}

	if err := cmd.Execute(); err != nil {
		zap.L().Debug("error", zap.String("stack trace", fmt.Sprintf("%+v\n", err)))
	}
	return nil
}

func NewLogger(cfg Config) (*zap.Logger, error) {
	zcfg := zap.NewProductionConfig()
	if cfg.Debug {
		zcfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	if cfg.Verbose {
		zcfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	l, err := zcfg.Build()
	return l, errors.WithStack(err)
}

func NewConfig() (Config, error) {
	viper.AutomaticEnv()
	viper.BindPFlags(pflag.CommandLine)

	var cfg Config
	pflag.Parse()
	err := viper.Unmarshal(&cfg)
	return cfg, errors.WithStack(err)
}
