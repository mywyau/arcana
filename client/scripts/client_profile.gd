class_name ClientProfile
extends RefCounted

const DEFAULT_NAME: String = "default"
const MAXIMUM_NAME_LENGTH: int = 32

static var _selected_name: String = ""


static func selected_name() -> String:
	if not _selected_name.is_empty():
		return _selected_name
	var requested_name: String = DEFAULT_NAME
	for argument: String in OS.get_cmdline_user_args():
		if argument.begins_with("--profile="):
			requested_name = argument.trim_prefix("--profile=")
			break
	_selected_name = _sanitize_name(requested_name)
	return _selected_name


static func config_path(base_filename: String) -> String:
	var profile_name: String = selected_name()
	if profile_name == DEFAULT_NAME:
		return "user://%s" % base_filename
	var extension: String = base_filename.get_extension()
	var stem: String = base_filename.trim_suffix(".%s" % extension)
	return "user://%s.%s.%s" % [stem, profile_name, extension]


static func _sanitize_name(requested_name: String) -> String:
	var normalized: String = requested_name.strip_edges().to_lower().left(MAXIMUM_NAME_LENGTH)
	var sanitized: String = ""
	for index: int in range(normalized.length()):
		var code: int = normalized.unicode_at(index)
		if (
			(code >= 97 and code <= 122)
			or (code >= 48 and code <= 57)
			or code == 45
			or code == 95
		):
			sanitized += normalized[index]
		else:
			sanitized += "_"
	return sanitized if not sanitized.is_empty() else DEFAULT_NAME
