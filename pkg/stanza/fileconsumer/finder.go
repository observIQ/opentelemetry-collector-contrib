// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package fileconsumer // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/fileconsumer"

import (
	"github.com/bmatcuk/doublestar/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Finder struct {
	Include []string `mapstructure:"include,omitempty"`
	Exclude []string `mapstructure:"exclude,omitempty"`
	logger  *zap.Logger
}

// FindFiles gets a list of paths given an array of glob patterns to include and exclude
//
// Deprecated: [v0.80.0] This will be made internal in a future release, tentatively v0.82.0.
func (f Finder) FindFiles() []string {
	all := make([]string, 0, len(f.Include))
	for _, include := range f.Include {
		var matches []string

		if f.logger != nil && f.logger.Level().Enabled(zapcore.DebugLevel) {
			// For debugging, we should log IO errors if they occur.
			var err error
			matches, err = doublestar.FilepathGlob(include, doublestar.WithFilesOnly(), doublestar.WithFailOnIOErrors())
			if err != nil {
				if f.logger != nil {
					f.logger.Debug("Error when globbing.", zap.String("glob", include), zap.Error(err))
				}
				matches, _ = doublestar.FilepathGlob(include, doublestar.WithFilesOnly()) // compile error checked in build
			}
		} else {
			matches, _ = doublestar.FilepathGlob(include, doublestar.WithFilesOnly()) // compile error checked in build
		}

	INCLUDE:
		for _, match := range matches {
			for _, exclude := range f.Exclude {
				if itMatches, _ := doublestar.PathMatch(exclude, match); itMatches {
					continue INCLUDE
				}
			}

			for _, existing := range all {
				if existing == match {
					continue INCLUDE
				}
			}

			all = append(all, match)
		}
	}

	return all
}
