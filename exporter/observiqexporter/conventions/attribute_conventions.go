package conventions

// Conventions for "Attributes"

// EntryTypeField is the field containing the type of entry
const EntryTypeField = "log_type"

// FileNameField is the field containing the file from which the entry came from
//  (if it did come from a file)
const FileNameField = "file_name"

// FilePathField is the field containing the director of the file from which the entry came from
//  (if it did come from a file)
const FilePathField = "file_path"

// PluginIDField is the id of the plugin that generated this entry (from stanza receiver)
const PluginIDField = "plugin_id"

// PluginNameField is the name of the plugin that generated this entry (from stanza receiver)
const PluginNameField = "plugin_name"

// PluginVersionField is the version of the plugin that generated this entry (from stanza receiver)
const PluginVersionField = "plugin_version"

// PluginTypeField is the type of the plugin that generated this entry (from stanza receiver)
const PluginTypeField = "plugin_type"
