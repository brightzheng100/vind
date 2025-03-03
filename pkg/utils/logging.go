/*
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
package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger = logrus.New()

func init() {
	Logger.SetFormatter(&logrus.TextFormatter{})

	// defaults to Info log level
	Logger.SetLevel(logrus.InfoLevel)

	// and log level is configurable by env vaiable $LOG_LEVEL
	config_log_level := os.Getenv("LOG_LEVEL")
	if config_log_level != "" {
		log_level, err := logrus.ParseLevel(config_log_level)
		if err != nil {
			Logger.Warnf("configured LOG_LEVEL is unparsable: %s, ignore and fall back to Info level", config_log_level)
		} else {
			Logger.Infof("log level is set to [%s]", log_level)
			Logger.SetLevel(log_level)
		}
	}
}
