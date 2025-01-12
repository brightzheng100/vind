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
	"fmt"
	"strings"

	release "github.com/brightzheng100/vind/pkg/version"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print vind version",
	Run:   showVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var version = "git"
var commit = ""
var date = ""

func showVersion(cmd *cobra.Command, args []string) {
	fmt.Println("version:", version, "commit:", commit, "date:", date)
	if version == "git" {
		return
	}
	release, err := release.FindLastRelease()
	if err != nil {
		fmt.Println("version: failed to check for new versions. You may want to check it out at https://github.com/brightzheng100/vind/releases.")
		return
	}
	if strings.Compare(version, *release.TagName) != 0 {
		fmt.Printf("New version %v is available. More information at: %v\n", *release.TagName, *release.HTMLURL)
	}
}
