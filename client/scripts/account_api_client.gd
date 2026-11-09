class_name AccountAPIClient
extends Node

signal authentication_succeeded(player_id: String, email: String, session_token: String)
signal request_failed(message: String)

var _base_url: String = ""
var _request: HTTPRequest
var _busy: bool = false


func _ready() -> void:
	_request = HTTPRequest.new()
	_request.timeout = 15.0
	add_child(_request)
	_request.request_completed.connect(_on_request_completed)


func configure(websocket_url: String) -> void:
	_base_url = str(ProjectSettings.get_setting("network/account_api_url", "")).strip_edges()
	var environment_url: String = OS.get_environment("ARCANA_ACCOUNT_API_URL").strip_edges()
	if not environment_url.is_empty():
		_base_url = environment_url
	for argument: String in OS.get_cmdline_user_args():
		if argument.begins_with("--account-api-url="):
			_base_url = argument.trim_prefix("--account-api-url=").strip_edges()
	if _base_url.is_empty():
		_base_url = _derive_base_url(websocket_url)
	_base_url = _base_url.trim_suffix("/")


func register(email: String, password: String, guest_token: String = "") -> bool:
	var payload: Dictionary = {"email": email.strip_edges(), "password": password}
	if not guest_token.is_empty():
		payload["guest_token"] = guest_token
	return _start_request("/api/account/register", payload)


func login(email: String, password: String) -> bool:
	return _start_request("/api/account/login", {
		"email": email.strip_edges(),
		"password": password,
	})


func is_busy() -> bool:
	return _busy


func _start_request(path: String, payload: Dictionary) -> bool:
	if _busy:
		request_failed.emit("An account request is already in progress.")
		return false
	if _base_url.is_empty():
		request_failed.emit("The account service is not configured.")
		return false
	_busy = true
	var result: Error = _request.request(
		_base_url + path,
		PackedStringArray(["Content-Type: application/json", "Accept: application/json"]),
		HTTPClient.METHOD_POST,
		JSON.stringify(payload),
	)
	if result != OK:
		_busy = false
		request_failed.emit("Could not contact the account service: %s" % error_string(result))
		return false
	return true


func _on_request_completed(
	result: int,
	response_code: int,
	_headers: PackedStringArray,
	body: PackedByteArray,
) -> void:
	_busy = false
	if result != HTTPRequest.RESULT_SUCCESS:
		request_failed.emit("The account service could not be reached.")
		return
	var payload: Variant = JSON.parse_string(body.get_string_from_utf8())
	if response_code < 200 or response_code >= 300:
		var message: String = "The account request failed."
		if typeof(payload) == TYPE_DICTIONARY:
			message = str((payload as Dictionary).get("error", message))
		request_failed.emit(message)
		return
	if typeof(payload) != TYPE_DICTIONARY:
		request_failed.emit("The account service returned an invalid response.")
		return
	var response := payload as Dictionary
	var session_token: String = str(response.get("session_token", ""))
	if session_token.is_empty():
		request_failed.emit("The account service did not return a session.")
		return
	authentication_succeeded.emit(
		str(response.get("player_id", "")),
		str(response.get("email", "")),
		session_token,
	)


func _derive_base_url(websocket_url: String) -> String:
	var value: String = websocket_url.strip_edges()
	if value.begins_with("wss://"):
		value = "https://" + value.trim_prefix("wss://")
	elif value.begins_with("ws://"):
		value = "http://" + value.trim_prefix("ws://")
	if value.ends_with("/ws"):
		value = value.trim_suffix("/ws")
	return value
