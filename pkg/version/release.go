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
package release

import (
	"context"
	"fmt"

	"github.com/google/go-github/v24/github"
)

const (
	owner = "brightzheng100"
	repo  = "vind"
)

// FindLastRelease searches latest release of the project
func FindLastRelease() (*github.RepositoryRelease, error) {
	githubclient := github.NewClient(nil)
	repoRelease, _, err := githubclient.Repositories.GetLatestRelease(context.Background(), owner, repo)
	if err != nil {
		return nil, fmt.Errorf("Failed to get latest release information")
	}
	return repoRelease, nil
}
