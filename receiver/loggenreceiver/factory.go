// Copyright  OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package loggenreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/loggenreceiver"

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
)

const (
	// Value of "type" key in configuration.
	typeStr = "loggen"
)

// NewFactory creates a factory for k8s_cluster receiver.
func NewFactory() component.ReceiverFactory {
	return component.NewReceiverFactory(
		typeStr,
		createDefaultConfig,
		component.WithLogsReceiver(createLogsReceiver))
}

func createDefaultConfig() config.Receiver {
	return &Config{
		ReceiverSettings: config.NewReceiverSettings(config.NewComponentID(typeStr)),
		LogsPerSec:       10_000,
		EmitInterval:     100 * time.Millisecond,
		LogLine:          `INFO [IndexSummaryManager:1] 2021-10-07 12:57:05,003 rwQuUEyQyUZTNeZEIhXqfXQWBaAAZinRtrLdvUqVNbTcYv4ZXqmIDp5s4Q8XtbhiqfI5gv3P1TGsqB9nlNXBVs2gaDsupJL4tOX28Y4rxjTvv8lGYS1Lg30lf4mcwu4N6xU2KPgCGC2eJ3SHO2FGAqQCPmC8jc9KkEGuUbspwU4xzovgoAxL3N65PaKlT2cNuwGIUKsg7rEAUcST4wsOIZgUIEHXzYr3cvGM0rOn8RVU4umcasbeW2DuMDKP2exMXQ0Z8BP9IJGqk3S6fyovezmYw8dgYcWVs0JpjDuxZzstjKZjYZTIv3uy2tUKn7RUVpEYfRXFdfuEr4lvhAj2VSnShAMycvfeZF4kQ2sKHFNGUIKYKxFZHKBK5dUnp0m5fyd6AeWK6X6ujM5Jnu0UtT7ctaST2S4kNgzeFdxG3wluR45xMDm2pS8kdzMgKWgt1XwzTsd09Sjnqd5i9aNeaJD3ASZLHZwIzreV8NhcfhP5LzownTDZ0RHpT62g6PMq20iKC7WjWRGxGHkq1M0Qznl0TNragVeAuI3NKMyWNzZcJ3v96KrCdJrbRSXosjAUjHWVqpgUUYHtMmDc8QNAQdAeCWj68cWVK1fpECnwzcOJxAsAPISWZ5dRxIk8C5TKRdhC6xpYhzaeF3Qp8lTXOon6cshObYcmLEADQP3Dbm5ifnrEeZ9fE6C96DZLNjJEWLkM10EnfMep7safKqZabZJEkxOvj21DKpE3RNTZwnpi0943IIA6K9kX5xBooHQARA76AajVQaJ90pwqwWFtfojy5vXEY8dyrGU8qtOW70SKHf4MpFAybRVniN6ZNWQINCVpV2gp1LjdykpfJQJD623f0RZB1mUGgDVpdGuRsgMKl8Zyt4L0diybEB3y1ZPmtO7YzUDrlighN8UBJ0a96PSwTZHFxVdfrfgJpmNS6X0zS1ZNdYaO1Xmeuz1EQJdOtei1IrIaWhp`,
	}
}

func createLogsReceiver(
	_ context.Context,
	params component.ReceiverCreateSettings,
	cfg config.Receiver,
	consumer consumer.Logs,
) (component.LogsReceiver, error) {
	rCfg := cfg.(*Config)

	return newReceiver(params, rCfg, consumer)
}
