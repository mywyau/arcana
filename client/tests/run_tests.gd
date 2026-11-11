extends SceneTree

const AUTH_GATE_SCENE := preload("res://scenes/auth_gate.tscn")
const MAIN_SCENE := preload("res://scenes/main.tscn")
const PLAYER_SCENE := preload("res://scenes/player.tscn")

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
	await _test_account_rate_limit_feedback_includes_retry_time()
	await _test_local_movement_reconciliation_ignores_stale_routes()
	_test_beta_build_uses_hosted_server()
	# Let queued Window and scene resources finish releasing before the test
	# process exits. This keeps headless runs deterministic on slower machines.
	for _frame: int in range(3):
		await process_frame

	if _failures > 0:
		printerr("Client tests failed: %d of %d checks failed." % [_failures, _checks])
		quit(1)
		return
	print("Client tests passed: %d checks." % _checks)
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
	main.call("_synchronize_command_sequences", {
		"last_command_sequence": 291,
		"last_action_sequence": 48,
		"last_inventory_move_sequence": 17,
		"last_fire_sequence": 6,
	})
	_check(int(main.get("_next_sequence")) == 292, "resumed movement continues after the server sequence")
	_check(int(main.get("_next_action_sequence")) == 49, "resumed combat continues after the server sequence")
	_check(
		int(main.get("_next_inventory_move_sequence")) == 18,
		"resumed inventory movement continues after the server sequence",
	)
	_check(int(main.get("_next_fire_sequence")) == 7, "resumed skilling continues after the server sequence")
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


func _test_account_rate_limit_feedback_includes_retry_time() -> void:
	var api := AccountAPIClient.new()
	root.add_child(api)
	await process_frame
	var failures: Array[String] = []
	api.request_failed.connect(func(message: String) -> void: failures.append(message))
	api._busy = true
	api._on_request_completed(
		HTTPRequest.RESULT_SUCCESS,
		429,
		PackedStringArray(["Content-Type: application/json", "Retry-After: 75"]),
		JSON.stringify({"error": "too many login attempts"}).to_utf8_buffer(),
	)
	_check(not api._busy, "rate-limited account request releases the busy state")
	_check(failures.size() == 1, "rate-limited account request reports one error")
	_check(
		failures.size() == 1 and failures[0].contains("2 minutes"),
		"rate-limit feedback includes a rounded retry time",
	)
	api.queue_free()
	await process_frame


func _test_local_movement_reconciliation_ignores_stale_routes() -> void:
	var player := PLAYER_SCENE.instantiate() as PlayerView
	root.add_child(player)
	await process_frame
	player.set_local_player(true)
	player.apply_authoritative_state(Vector3.ZERO, [], 6.0, 0, 0, false)

	player.global_position = Vector3(1.0, 0.0, 0.0)
	player.apply_authoritative_state(Vector3.ZERO, [], 6.0, 1, 1, false)
	_check(not player._pending_correction.is_zero_approx(), "acknowledged movement drift creates a correction")

	var latest_destination := Vector3(-5.0, 0.0, 0.0)
	player.predict_move(latest_destination)
	_check(player._pending_correction.is_zero_approx(), "new prediction clears the previous correction")

	player.global_position = Vector3(2.0, 0.0, 0.0)
	player.apply_authoritative_state(Vector3.ZERO, [], 6.0, 1, 2, false)
	_check(player._pending_correction.is_zero_approx(), "stale acknowledgement does not pull local prediction backwards")
	_check(
		player._predicted_path == [latest_destination],
		"stale acknowledgement preserves the newest predicted route",
	)

	player.global_position = Vector3(1.2, 0.0, 0.0)
	var acknowledged_path: Array[Vector3] = [
		Vector3(0.5, 0.0, 0.0),
		Vector3(1.0, 0.0, 0.0),
		Vector3(1.5, 0.0, 0.0),
		Vector3(2.0, 0.0, 0.0),
	]
	player.apply_authoritative_state(
		Vector3(0.6, 0.0, 0.0),
		acknowledged_path,
		6.0,
		2,
		2,
		false,
	)
	_check(
		player._predicted_path == [Vector3(1.5, 0.0, 0.0), Vector3(2.0, 0.0, 0.0)],
		"acknowledged route removes waypoints already passed by prediction",
	)

	player.apply_authoritative_state(
		Vector3(0.8, 0.0, 0.0),
		[Vector3(1.0, 0.0, 0.0), Vector3(1.5, 0.0, 0.0), Vector3(2.0, 0.0, 0.0)],
		6.0,
		2,
		2,
		true,
	)
	player._predicted_path = [Vector3(1.6, 0.0, 0.0), Vector3(2.0, 0.0, 0.0)]
	player.apply_authoritative_state(
		Vector3(1.0, 0.0, 0.0),
		[Vector3(1.5, 0.0, 0.0), Vector3(2.0, 0.0, 0.0)],
		6.0,
		2,
		2,
		true,
	)
	_check(
		player._predicted_path == [Vector3(1.6, 0.0, 0.0), Vector3(2.0, 0.0, 0.0)],
		"equivalent server-driven route is not reset every snapshot",
	)
	player.queue_free()
	await process_frame


func _test_beta_build_uses_hosted_server() -> void:
	var local_url: String = NetworkClient.default_server_url(false)
	var beta_url: String = NetworkClient.default_server_url(true)
	_check(local_url.begins_with("ws://localhost:"), "development builds default to localhost")
	_check(beta_url.begins_with("wss://"), "beta builds default to a secure WebSocket endpoint")
	_check(not beta_url.contains("localhost"), "beta builds do not default to localhost")


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
