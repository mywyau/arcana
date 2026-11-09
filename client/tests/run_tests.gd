extends SceneTree

const AUTH_GATE_SCENE := preload("res://scenes/auth_gate.tscn")
const MAIN_SCENE := preload("res://scenes/main.tscn")

var _failures: int = 0
var _checks: int = 0


class RetryNetworkClient extends NetworkClient:
	var connection_attempts: int = 0

	func _ready() -> void:
		pass

	func _connect_to_server() -> void:
		connection_attempts += 1


func _init() -> void:
	_run.call_deferred()


func _run() -> void:
	await _test_revoked_session_requires_login()
	await _test_incompatible_client_requires_update()
	await _test_unavailable_server_exposes_retry()
	await _test_retry_starts_connection_immediately()
	await _test_replaced_session_stops_and_can_reconnect()
	await _test_logout_returns_to_account_gate()
	await _test_welcome_hides_account_gate()
	await _test_duplicate_login_submission_is_blocked()
	# Let queued Window and scene resources finish releasing before the test
	# process exits. This keeps headless runs deterministic on slower machines.
	for _frame: int in range(3):
		await process_frame

	if _failures > 0:
		printerr("Auth tests failed: %d of %d checks failed." % [_failures, _checks])
		quit(1)
		return
	print("Auth tests passed: %d checks." % _checks)
	quit(0)


func _test_revoked_session_requires_login() -> void:
	var network := _isolated_network_client("revoked")
	var login_messages: Array[String] = []
	network.account_login_required.connect(func(message: String) -> void: login_messages.append(message))
	network._session_token = "revoked-session"
	network._handle_authentication_close()
	_check(network._session_token.is_empty(), "revoked session token is cleared")
	_check(network._authentication_blocked, "revoked session blocks reconnection")
	_check(login_messages.size() == 1, "revoked session requests the login gate")
	_cleanup_network_identity(network)
	network.free()
	await process_frame


func _test_incompatible_client_requires_update() -> void:
	var harness: Dictionary = await _main_auth_harness()
	var main: Node = harness["main"]
	var network: NetworkClient = harness["network"]
	var gate: AuthGate = harness["gate"]
	network._session_token = "still-valid-after-update"
	network._handle_packet(JSON.stringify({
		"type": "error",
		"error": {
			"code": "incompatible_client",
			"message": "This client is incompatible with the server; update required.",
		},
	}).to_utf8_buffer())
	await process_frame
	_check(gate.visible, "incompatible client displays the account gate")
	_check(gate.title_label.text == "Update required", "incompatible client explains the required action")
	_check(network._authentication_blocked, "incompatible client stops reconnect attempts")
	_check(network._session_token == "still-valid-after-update", "client update does not revoke the account session")
	_cleanup_main_harness(main, network)
	await process_frame


func _test_unavailable_server_exposes_retry() -> void:
	var gate: AuthGate = AUTH_GATE_SCENE.instantiate()
	root.add_child(gate)
	await process_frame
	gate.show_connecting("Restoring your session…")
	gate.show_connection_status("The game server is unavailable.", true, true)
	_check(gate.is_connection_pending(), "unavailable server remains in connection mode")
	_check(gate.retry_button.visible, "unavailable server shows Retry")
	_check(not gate.retry_button.disabled, "Retry is actionable")
	_check(gate.feedback_label.text.contains("unavailable"), "unavailable status is explained")
	gate.queue_free()
	await process_frame


func _test_retry_starts_connection_immediately() -> void:
	var network := RetryNetworkClient.new()
	root.add_child(network)
	await process_frame
	network._authentication_blocked = false
	network._reconnect_at_msec = Time.get_ticks_msec() + 30_000
	var accepted: bool = network.retry_connection()
	await process_frame
	_check(accepted, "manual retry is accepted for an authenticated identity")
	_check(network.connection_attempts == 1, "manual retry starts one connection attempt")
	_check(network._reconnect_at_msec == -1, "manual retry cancels the reconnect delay")
	network.queue_free()
	await process_frame


func _test_replaced_session_stops_and_can_reconnect() -> void:
	var harness: Dictionary = await _main_auth_harness()
	var main: Node = harness["main"]
	var network: NetworkClient = harness["network"]
	var gate: AuthGate = harness["gate"]
	network._session_token = "replacement-session"
	network._handle_packet(JSON.stringify({
		"type": "error",
		"error": {
			"code": "session_replaced",
			"message": "This character was logged in from another client.",
		},
	}).to_utf8_buffer())
	await process_frame
	_check(gate.visible, "replaced session displays the account gate")
	_check(gate.title_label.text == "Logged in elsewhere", "replaced session explains the disconnect")
	_check(gate.retry_button.visible, "replaced session offers deliberate reconnection")
	_check(network._authentication_blocked, "replaced session stops automatic reconnects")
	_check(network._session_token == "replacement-session", "replacement keeps the account session")
	_cleanup_main_harness(main, network)
	await process_frame

	var retry_network := RetryNetworkClient.new()
	root.add_child(retry_network)
	await process_frame
	retry_network._session_token = "replacement-session"
	retry_network._session_was_replaced = true
	retry_network._authentication_blocked = true
	var accepted: bool = retry_network.reconnect_replaced_session()
	await process_frame
	_check(accepted, "deliberate replaced-session reconnect is accepted")
	_check(not retry_network._authentication_blocked, "deliberate reconnect unblocks authentication")
	_check(retry_network.connection_attempts == 1, "deliberate reconnect starts one connection")
	retry_network.queue_free()
	await process_frame


func _test_logout_returns_to_account_gate() -> void:
	var harness: Dictionary = await _main_auth_harness()
	var main: Node = harness["main"]
	var network: NetworkClient = harness["network"]
	var gate: AuthGate = harness["gate"]
	gate.dismiss()
	network._session_token = "active-session"
	network._handle_packet(JSON.stringify({"type": "account_logged_out"}).to_utf8_buffer())
	await process_frame
	_check(gate.visible, "logout displays the account gate")
	_check(not gate.is_connection_pending(), "logout displays an editable login form")
	_check(network._session_token.is_empty(), "logout clears the saved session")
	_cleanup_main_harness(main, network)
	await process_frame


func _test_welcome_hides_account_gate() -> void:
	var harness: Dictionary = await _main_auth_harness()
	var main: Node = harness["main"]
	var network: NetworkClient = harness["network"]
	var gate: AuthGate = harness["gate"]
	gate.show_connecting()
	network.welcome_received.emit("test-player")
	await process_frame
	_check(not gate.visible, "successful welcome hides the account gate")
	_cleanup_main_harness(main, network)
	await process_frame


func _test_duplicate_login_submission_is_blocked() -> void:
	var gate: AuthGate = AUTH_GATE_SCENE.instantiate()
	root.add_child(gate)
	await process_frame
	var state: Dictionary = {"submissions": 0}
	gate.login_requested.connect(
		func(_email: String, _password: String) -> void: state["submissions"] += 1
	)
	gate.show_login()
	gate.email_input.text = "player@example.com"
	gate.password_input.text = "strong-password"
	gate._submit()
	gate._submit()
	_check(int(state["submissions"]) == 1, "duplicate login submission is ignored while busy")
	_check(gate.submit_button.disabled, "login controls remain locked during the request")
	gate.queue_free()
	await process_frame


func _main_auth_harness() -> Dictionary:
	var main: Node = MAIN_SCENE.instantiate()
	var network: NetworkClient = main.get_node("NetworkClient")
	network._guest_identity_path = _temporary_identity_path("main")
	root.add_child(main)
	network._authentication_blocked = true
	network._reconnect_at_msec = -1
	await process_frame
	return {
		"main": main,
		"network": network,
		"gate": main.get_node("UI/AuthGate"),
	}


func _cleanup_main_harness(main: Node, network: NetworkClient) -> void:
	_cleanup_network_identity(network)
	main.queue_free()


func _isolated_network_client(test_name: String) -> NetworkClient:
	var network := NetworkClient.new()
	network._guest_identity_path = _temporary_identity_path(test_name)
	network._guest_identity_section = "auth_test"
	return network


func _temporary_identity_path(test_name: String) -> String:
	return "user://auth_test_%s_%d.cfg" % [test_name, Time.get_ticks_usec()]


func _cleanup_network_identity(network: NetworkClient) -> void:
	var absolute_path: String = ProjectSettings.globalize_path(network._guest_identity_path)
	if FileAccess.file_exists(absolute_path):
		DirAccess.remove_absolute(absolute_path)


func _check(condition: bool, description: String) -> void:
	_checks += 1
	if condition:
		print("PASS: %s" % description)
		return
	_failures += 1
	printerr("FAIL: %s" % description)
