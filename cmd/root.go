/*
Copyright © 2019-2023 footloose developers
Copyright © 2024-2025 Bright Zheng <bright.zheng@outlook.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var log = logrus.New()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vind",
	Short: "A tool to create containers that look and work like virtual machines, on Docker.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var cfgFile struct {
	config string
}

func init() {
	// init logging framework
	initLog()

	rootCmd.PersistentFlags().StringVarP(&cfgFile.config, "config", "c", "", "Cluster configuration file")
}

func initLog() {
	log.SetFormatter(&logrus.TextFormatter{})
	log.SetOutput(os.Stdout)

	// defaults to Info log level
	log.SetLevel(logrus.InfoLevel)

	// and log level is configurable by env vaiable $LOG_LEVEL
	config_log_level := os.Getenv("LOG_LEVEL")
	if config_log_level != "" {
		log_level, err := logrus.ParseLevel(config_log_level)
		if err != nil {
			log.Warnf("configured LOG_LEVEL is unparsable: %s, ignore and fall back to Info level", config_log_level)
		} else {
			log.Infof("log level is set to [%s]", log_level)
			log.SetLevel(log_level)
		}
	}
}
