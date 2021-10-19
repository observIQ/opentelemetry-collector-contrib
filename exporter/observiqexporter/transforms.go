package observiqexporter

import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/observiqexporter/conventions"

// doExtraTransforms transforms the observIQLogEntry by 'hoisting' certain fields off of Labels and Data to the top level observIQLogEntry
func doExtraTransforms(le *observIQLogEntry) {
	pullMessageFromData(le)
	fillEntryType(le)
	fillLogInfo(le)
	fillSourceInfo(le)
}

// pullMessageFromData moves the "message" key-value pair off the Body and to the top-level message field.
func pullMessageFromData(le *observIQLogEntry) {
	if message, ok := le.Data[conventions.MessageField]; ok {
		if messageAsString, ok := message.(string); ok {
			le.Message = messageAsString
			delete(le.Data, conventions.MessageField)
			return
		}
	}

	if message, ok := le.Data[conventions.MessageShortField]; ok {
		if messageAsString, ok := message.(string); ok {
			le.Message = messageAsString
			delete(le.Data, conventions.MessageShortField)
			return
		}
	}
}

const fileInputType = "file_input"

func fillEntryType(le *observIQLogEntry) {
	entryType := ""
	if logTypeLabel, ok := le.Labels[conventions.EntryTypeField]; ok {
		delete(le.Labels, conventions.EntryTypeField)
		entryType = logTypeLabel

		if entryType == fileInputType {
			entryType = "file"
		}
	}

	if entryType == "" {
		if eType, ok := le.Data[conventions.EntryTypeAltField]; ok {
			if eTypeString, ok := eType.(string); ok {
				delete(le.Data, conventions.EntryTypeAltField)
				entryType = eTypeString
			}
		}
	}

	le.EntryType = entryType
}

//Get log info from the Entry
func fillLogInfo(le *observIQLogEntry) {
	var fileName string
	if fn, ok := le.Labels[conventions.FileNameField]; ok {
		fileName = fn
		delete(le.Labels, conventions.FileNameField)
	}

	var filePath string
	if fn, ok := le.Labels[conventions.FilePathField]; ok {
		filePath = fn
		delete(le.Labels, conventions.FilePathField)
	}
	if fileName != "" || filePath != "" {
		le.LogInfo = &observIQLogInfo{
			Name: fileName,
			Path: filePath,
		}
	}
}

func fillSourceInfo(le *observIQLogEntry) {
	var sourcePluginID string
	if pluginID, ok := le.Labels[conventions.PluginIDField]; ok {
		sourcePluginID = pluginID
		delete(le.Labels, conventions.PluginIDField)
	}

	var sourcePluginName string
	if pluginName, ok := le.Labels[conventions.PluginNameField]; ok {
		sourcePluginName = pluginName
		delete(le.Labels, conventions.PluginNameField)
	}

	var sourcePluginVersion string
	if pluginVersion, ok := le.Labels[conventions.PluginVersionField]; ok {
		sourcePluginVersion = pluginVersion
		delete(le.Labels, conventions.PluginVersionField)
	}

	var sourcePluginType string
	if pluginType, ok := le.Labels[conventions.PluginTypeField]; ok {
		sourcePluginType = pluginType
		delete(le.Labels, conventions.PluginTypeField)
	}

	if sourcePluginID != "" || sourcePluginName != "" ||
		sourcePluginType != "" || sourcePluginVersion != "" {
		le.SourceInfo = &observIQLogEntrySource{
			ID:         sourcePluginID,
			SourceType: sourcePluginType,
			Name:       sourcePluginName,
			Version:    sourcePluginVersion,
		}
	}
}
