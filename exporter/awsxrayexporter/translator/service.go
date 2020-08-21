// Copyright 2019, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package translator

import (
	"github.com/aws/aws-sdk-go/aws"
	"go.opentelemetry.io/collector/consumer/pdata"
	semconventions "go.opentelemetry.io/collector/translator/conventions"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/common/awsxray"
)

func makeService(resource pdata.Resource) *awsxray.ServiceData {
	var (
		service *awsxray.ServiceData
	)
	if resource.IsNil() {
		return service
	}
	verStr, ok := resource.Attributes().Get(semconventions.AttributeServiceVersion)
	if !ok {
		verStr, ok = resource.Attributes().Get(semconventions.AttributeContainerTag)
	}
	if ok {
		service = &awsxray.ServiceData{
			Version: aws.String(verStr.StringVal()),
		}
	}
	return service
}
