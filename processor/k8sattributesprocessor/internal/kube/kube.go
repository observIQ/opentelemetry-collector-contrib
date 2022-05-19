// Copyright 2020 OpenTelemetry Authors
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

package kube // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor/internal/kube"

import (
	"fmt"
	"regexp"
	"time"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig"
)

const (
	podNodeField            = "spec.nodeName"
	ignoreAnnotation string = "opentelemetry.io/k8s-processor/ignore"
	tagNodeName             = "k8s.node.name"
	tagStartTime            = "k8s.pod.start_time"
	// MetadataFromPod is used to specify to extract metadata/labels/annotations from pod
	MetadataFromPod = "pod"
	// MetadataFromNamespace is used to specify to extract metadata/labels/annotations from namespace
	MetadataFromNamespace = "namespace"
)

// PodIdentifier is a custom type to represent IP Address or Pod UID
type PodIdentifier string

var (
	// TODO: move these to config with default values
	defaultPodDeleteGracePeriod = time.Second * 120
	watchSyncPeriod             = time.Minute * 5
)

// Client defines the main interface that allows querying pods by metadata.
type Client interface {
	GetPod(PodIdentifier) (*Pod, bool)
	GetNamespace(string) (*Namespace, bool)
	Start()
	Stop()
}

// ClientProvider defines a func type that returns a new Client.
type ClientProvider func(*zap.Logger, k8sconfig.APIConfig, ExtractionRules, Filters, []Association, Excludes, APIClientsetProvider, InformerProvider, InformerProviderNamespace) (Client, error)

// APIClientsetProvider defines a func type that initializes and return a new kubernetes
// Clientset object.
type APIClientsetProvider func(config k8sconfig.APIConfig) (kubernetes.Interface, error)

// Pod represents a kubernetes pod.
type Pod struct {
	Name       string
	Address    string
	PodUID     string
	Attributes map[string]string
	StartTime  *metav1.Time
	Ignore     bool
	Namespace  string

	// Containers is a map of container name to Container struct.
	Containers map[string]*Container

	DeletedAt time.Time
}

// Container stores resource attributes for a specific container defined by k8s pod spec.
type Container struct {
	ImageName string
	ImageTag  string

	// Statuses is a map of container k8s.container.restart_count attribute to ContainerStatus struct.
	Statuses map[int]ContainerStatus
}

// ContainerStatus stores resource attributes for a particular container run defined by k8s pod status.
type ContainerStatus struct {
	ContainerID string
}

// Namespace represents a kubernetes namespace.
type Namespace struct {
	Name         string
	NamespaceUID string
	Attributes   map[string]string
	StartTime    metav1.Time
	DeletedAt    time.Time
}

type deleteRequest struct {
	// id is identifier (IP address or Pod UID) of pod to remove from pods map
	id PodIdentifier
	// name contains name of pod to remove from pods map
	podName string
	ts      time.Time
}

// Filters is used to instruct the client on how to filter out k8s pods.
// Right now only filters supported are the ones supported by k8s API itself
// for performance reasons. We can support adding additional custom filters
// in future if there is a real need.
type Filters struct {
	Node      string
	Namespace string
	Fields    []FieldFilter
	Labels    []FieldFilter
}

// FieldFilter represents exactly one filter by field rule.
type FieldFilter struct {
	// Key matches the field name.
	Key string
	// Value matches the field value.
	Value string
	// Op determines the matching operation.
	// Currently only two operations are supported,
	//  - Equals
	//  - NotEquals
	Op selection.Operator
}

// ExtractionRules is used to specify the information that needs to be extracted
// from pods and added to the spans as tags.
type ExtractionRules struct {
	Deployment         bool
	Namespace          bool
	PodName            bool
	PodUID             bool
	Node               bool
	StartTime          bool
	ContainerID        bool
	ContainerImageName bool
	ContainerImageTag  bool

	Annotations []FieldExtractionRule
	Labels      []FieldExtractionRule
}

// FieldExtractionRule is used to specify which fields to extract from pod fields
// and inject into spans as attributes.
type FieldExtractionRule struct {
	// Name is used to as the Span tag name.
	Name string
	// Key is used to lookup k8s pod fields.
	Key string
	// KeyRegex is a regular expression used to extract a Key that matches the regex.
	KeyRegex             *regexp.Regexp
	HasKeyRegexReference bool
	// Regex is a regular expression used to extract a sub-part of a field value.
	// Full value is extracted when no regexp is provided.
	Regex *regexp.Regexp
	// From determines the kubernetes object the field should be retrieved from.
	// Currently only two values are supported,
	//  - pod
	//  - namespace
	From string
}

func (r *FieldExtractionRule) extractFromPodMetadata(metadata map[string]string, tags map[string]string, formatter string) {
	// By default if the From field is not set for labels and annotations we want to extract them from pod
	if r.From == MetadataFromPod || r.From == "" {
		r.extractFromMetadata(metadata, tags, formatter)
	}
}

func (r *FieldExtractionRule) extractFromNamespaceMetadata(metadata map[string]string, tags map[string]string, formatter string) {
	if r.From == MetadataFromNamespace {
		r.extractFromMetadata(metadata, tags, formatter)
	}
}

func (r *FieldExtractionRule) extractFromMetadata(metadata map[string]string, tags map[string]string, formatter string) {
	if r.KeyRegex != nil {
		for k, v := range metadata {
			if r.KeyRegex.MatchString(k) && v != "" {
				var name string
				if r.HasKeyRegexReference {
					result := []byte{}
					name = string(r.KeyRegex.ExpandString(result, r.Name, k, r.KeyRegex.FindStringSubmatchIndex(k)))
				} else {
					name = fmt.Sprintf(formatter, k)
				}
				tags[name] = v
			}
		}
	} else if v, ok := metadata[r.Key]; ok {
		tags[r.Name] = r.extractField(v)
	}
}

func (r *FieldExtractionRule) extractField(v string) string {
	// Check if a subset of the field should be extracted with a regular expression
	// instead of the whole field.
	if r.Regex == nil {
		return v
	}

	matches := r.Regex.FindStringSubmatch(v)
	if len(matches) == 2 {
		return matches[1]
	}
	return ""
}

// Associations represent a list of rules for Pod metadata associations with resources
type Associations struct {
	Associations []Association
}

// Association represents one association rule
type Association struct {
	From string
	Name string
}

// Excludes represent a list of Pods to ignore
type Excludes struct {
	Pods []ExcludePods
}

// ExcludePods represent a Pod name to ignore
type ExcludePods struct {
	Name *regexp.Regexp
}
