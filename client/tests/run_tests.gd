extends SceneTree

const AUTH_GATE_SCENE := preload("res://scenes/auth_gate.tscn")
const MAIN_SCENE := preload("res://scenes/main.tscn")
const PLAYER_SCENE := preload("res://scenes/player.tscn")
const XP_TRACKER_SCENE := preload("res://scenes/xp_tracker.tscn")
const MERCHANT_WINDOW_SCENE := preload("res://scenes/merchant_window.tscn")

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
	await _test_chat_message_packet_is_forwarded()
	_test_chat_channel_selection_is_clear()
	await _test_duplicate_login_submission_is_blocked()
	await _test_registration_collects_public_character_name()
	await _test_account_rate_limit_feedback_includes_retry_time()
	await _test_local_movement_reconciliation_ignores_stale_routes()
	await _test_player_pvp_targeting_is_toggleable()
	await _test_equipped_boots_follow_leg_pivots()
	await _test_equipped_cape_sits_behind_player()
	await _test_hit_splats_and_minimap_movement_facing()
	await _test_action_rejection_feedback()
	await _test_xp_tracker_supports_overlapping_drops()
	await _test_consumable_slot_shows_authoritative_cooldown()
	_test_fletching_options_follow_clicked_item_and_inventory()
	_test_skill_guides_show_authoritative_unlock_levels()
	_test_skill_slot_tooltip_and_hover_style()
	_test_light_interface_theme_and_chest_slots()
	_test_inventory_hud_compact_layout()
	await _test_shopkeeper_uses_icon_led_stock_panel()
	_test_equipment_details_only_show_equipment_bonuses()
	await _test_minimap_layout_and_projection()
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


func _test_chat_message_packet_is_forwarded() -> void:
	var network := _isolated_network_client("chat-packet")
	var received: Array[Dictionary] = []
	network.chat_message_received.connect(func(message: Dictionary) -> void: received.append(message))
	network._handle_packet(JSON.stringify({
		"type": "chat_message",
		"chat": {
			"id": 7,
			"channel": "local",
			"sender_id": "player-one",
			"sender_name": "Player One",
			"text": "Hello nearby players",
			"sent_at_unix_ms": 1795000000000,
		},
	}).to_utf8_buffer())
	_check(received.size() == 1, "chat packet is forwarded once")
	_check(
		received.size() == 1 and str(received[0].get("sender_name", "")) == "Player One",
		"chat packet preserves the authoritative sender name",
	)
	_cleanup_network_identity(network)
	network.free()
	await process_frame


func _test_chat_channel_selection_is_clear() -> void:
	var main: Node = MAIN_SCENE.instantiate()
	var game_tab := main.get_node("UI/ChatHUD/Margin/VBox/Channels/Game") as Button
	main.call("_style_chat_channel_tab", game_tab)
	var normal := game_tab.get_theme_stylebox("normal") as StyleBoxFlat
	var selected := game_tab.get_theme_stylebox("pressed") as StyleBoxFlat
	_check(
		selected.bg_color.g > selected.bg_color.r and selected.bg_color != normal.bg_color,
		"selected chat channel uses a distinct green state",
	)
	main.free()


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
	_test_snapshot_deltas_preserve_unchanged_public_state(main)
	_cleanup_main_harness(main, network)
	await process_frame


func _test_snapshot_deltas_preserve_unchanged_public_state(main: Node) -> void:
	main.set("_local_player_id", "local")
	var initial_players: Array = main.call("_materialize_player_payloads", {
		"players": [{"id": "local", "position": {"x": 0.0, "y": 0.0, "z": 0.0}}],
		"player_updates": [{
			"id": "remote",
			"position": {"x": 2.0, "y": 0.0, "z": 3.0},
			"speed": 6.0,
			"health": 20,
			"maximum_health": 20,
			"equipped_weapon": "iron_sword",
		}],
	}, true)
	_check(initial_players.size() == 2, "full delta materializes local and remote players")
	var changed_players: Array = main.call("_materialize_player_payloads", {
		"players": [{"id": "local", "position": {"x": 0.0, "y": 0.0, "z": 0.0}}],
		"player_updates": [{"id": "remote", "position": {"x": 4.0, "y": 0.0, "z": 3.0}}],
	}, false)
	var remote: Dictionary = changed_players[1]
	_check(float((remote["position"] as Dictionary)["x"]) == 4.0, "player delta applies changed position")
	_check(str(remote["equipped_weapon"]) == "iron_sword", "player delta retains unchanged equipment")
	_check(int(remote["health"]) == 20, "player delta retains unchanged health")

	var initial_targets: Array = main.call("_materialize_target_payloads", {
		"target_updates": [{
			"id": "target", "name": "Target", "kind": "creature",
			"position": {"x": 3.0, "y": 0.0, "z": 1.0}, "health": 10,
		}],
	}, true)
	_check(initial_targets.size() == 1, "full delta materializes targets")
	var changed_targets: Array = main.call("_materialize_target_payloads", {
		"target_updates": [{"id": "target", "health": 0}],
	}, false)
	var target: Dictionary = changed_targets[0]
	_check(int(target["health"]) == 0, "target delta preserves meaningful zero values")
	_check(str(target["name"]) == "Target", "target delta retains unchanged identity")

	main.call("_on_snapshot_received", {
		"tick": 1,
		"action_pulse": 0,
		"full_reconciliation": true,
		"players": [{
			"id": "local",
			"position": {"x": 0.0, "y": 0.0, "z": 0.0},
			"combat_target": "dummy",
			"combat_matchup": {
				"target_id": "dummy", "player_hit_chance": 0.72,
				"player_minimum_damage": 8, "player_maximum_damage": 12,
				"target_hit_chance": 0.45, "target_minimum_damage": 2,
				"target_maximum_damage": 4, "player_attack_range": 2.0,
			},
			"equipped_weapon": "iron_sword",
			"equipped_weapon_instance_id": "sword-instance",
			"equipped_armour": {"head": "iron_helm"},
			"equipped_armour_instances": {"head": "helm-instance"},
			"equipment_revision": 1,
			"skills_revision": 1,
			"skills": [{
				"id": "attack", "experience": 125, "level": 2,
				"current_level_experience": 83, "next_level_experience": 174,
			}],
			"combat_boosts_revision": 1,
			"active_combat_boosts": [{
				"skill": "attack", "levels": 3, "expires_at_pulse": 500,
			}],
			"brewing_revision": 1,
			"last_brew_sequence": 1,
			"brew_remaining": 2,
			"next_brew_pulse": 5,
			"brewing_result": "brewing",
			"fletching_revision": 1,
			"last_fletch_sequence": 1,
			"fletch_remaining": 3,
			"next_fletch_pulse": 6,
			"fletching_result": "fletching",
			"gathering_revision": 1,
			"last_gather_sequence": 1,
			"gathering_spot": "starter_fishing_spot",
			"next_gather_pulse": 7,
			"gathering_result": "fishing",
			"inventory_revision": 1,
			"inventory": {"capacity": 20, "items": [
				{"slot": 0, "item_id": "cooked_minnow", "quantity": 1},
				{"slot": 1, "item_id": "prayer_potion", "quantity": 1},
			]},
			"next_food_consume_pulse": 2,
			"next_potion_consume_pulse": 0,
			"bank_revision": 1,
			"bank": {"capacity": 100, "next_expansion_cost": 1000, "items": [{
				"slot": 0, "item_id": "air_rune", "quantity": 5,
			}]},
			"ammunition_pouch_revision": 1,
			"ammunition_pouch": {"capacity": 10, "unlocked_slots": 2, "items": [{
				"slot": 0, "item_id": "bronze_arrow", "quantity": 10,
			}]},
			"ring_rack_revision": 1,
			"ring_rack": {"capacity": 4, "items": [{
				"slot": 0, "item_id": "ruby_ring", "quantity": 1,
				"instance_id": "ring-instance",
			}]},
			"rune_pouch_revision": 1,
			"rune_pouch": {"capacity": 9, "unlocked_slots": 2, "items": [{
				"slot": 0, "item_id": "air_rune", "quantity": 5,
			}]},
		}],
		"target_updates": [{
			"id": "dummy", "name": "Target Dummy", "kind": "dummy",
			"position": {"x": 2.0, "y": 0.0, "z": 0.0},
			"health": 100, "maximum_health": 100, "combat_level": 1,
		}],
	})
	_check((main.get("player_details_accuracy") as Label).text == "72%", "combat matchup displays player hit chance")
	_check((main.get("player_details_damage") as Label).text == "8–12", "combat matchup displays player damage range")
	main.call("_on_snapshot_received", {
		"tick": 2,
		"action_pulse": 1,
		"players": [{
			"id": "local",
			"position": {"x": 0.0, "y": 0.0, "z": 0.0},
			"combat_target": "dummy",
			"equipped_weapon": "iron_sword",
			"equipped_armour": {"head": "iron_helm"},
			"equipment_revision": 1,
			"skills_revision": 1,
			"combat_boosts_revision": 1,
			"brewing_revision": 1,
			"fletching_revision": 1,
			"gathering_revision": 1,
			"inventory_revision": 1,
			"bank_revision": 1,
			"ammunition_pouch_revision": 1,
			"ring_rack_revision": 1,
			"rune_pouch_revision": 1,
			"next_food_consume_pulse": 2,
			"next_potion_consume_pulse": 0,
		}],
	})
	_check(
		str(main.get("_equipped_weapon_instance_id")) == "sword-instance",
		"unchanged private equipment metadata survives an omitted payload",
	)
	_check(
		str((main.get("_equipped_armour_instances") as Dictionary).get("head", "")) == "helm-instance",
		"unchanged private armour metadata survives an omitted payload",
	)
	_check(
		(main.get("_inventory_items") as Array).size() == 2,
		"unchanged inventory survives an omitted payload",
	)
	_check(
		(main.get("_bank_items") as Array).size() == 1,
		"unchanged bank survives an omitted payload",
	)
	_check(
		(main.get("_ammunition_pouch_items") as Array).size() == 1,
		"unchanged ammunition pouch survives an omitted payload",
	)
	_check(
		(main.get("_ring_rack_items") as Array).size() == 1,
		"unchanged ring rack survives an omitted payload",
	)
	_check(
		(main.get("_rune_pouch_items") as Array).size() == 1,
		"unchanged rune pouch survives an omitted payload",
	)
	_check(
		int(((main.get("_skill_progress") as Dictionary)["attack"] as Dictionary)["experience"]) == 125,
		"unchanged skills survive an omitted payload",
	)
	_check(
		(main.get("_active_combat_boosts") as Array).size() == 1,
		"unchanged combat boosts survive an omitted payload",
	)
	_check(int(main.get("_brew_remaining")) == 2, "unchanged brewing progress survives an omitted payload")
	_check(int(main.get("_fletch_remaining")) == 3, "unchanged fletching progress survives an omitted payload")
	_check(str(main.get("_gathering_spot")) == "starter_fishing_spot", "unchanged gathering progress survives an omitted payload")
	_check((main.get("player_details_accuracy") as Label).text == "72%", "omitted matchup payload preserves displayed hit chance")
	_check((main.get("player_details_damage") as Label).text == "8–12", "omitted matchup payload preserves displayed damage range")
	main.call("_on_snapshot_received", {
		"tick": 3,
		"action_pulse": 1,
		"players": [{
			"id": "local",
			"position": {"x": 0.0, "y": 0.0, "z": 0.0},
			"skills_revision": 2,
			"skills": [{
				"id": "attack", "experience": 174, "level": 3,
				"current_level_experience": 174, "next_level_experience": 276,
			}],
			"combat_boosts_revision": 2,
			"active_combat_boosts": [],
			"next_food_consume_pulse": 2,
			"next_potion_consume_pulse": 0,
		}],
	})
	_check(
		int(((main.get("_skill_progress") as Dictionary)["attack"] as Dictionary)["experience"]) == 174,
		"changed skills replace the cached skill payload",
	)
	_check(
		(main.get("_active_combat_boosts") as Array).is_empty(),
		"an explicit empty combat boost payload clears expired boosts",
	)
	main.call("_on_snapshot_received", {
		"tick": 4,
		"action_pulse": 1,
		"players": [{
			"id": "local",
			"position": {"x": 0.0, "y": 0.0, "z": 0.0},
			"brewing_revision": 2,
			"fletching_revision": 2,
			"gathering_revision": 2,
			"next_food_consume_pulse": 2,
			"next_potion_consume_pulse": 0,
		}],
	})
	_check(int(main.get("_brew_remaining")) == 0, "completed brewing clears progress with an explicit revision")
	_check(int(main.get("_fletch_remaining")) == 0, "completed fletching clears progress with an explicit revision")
	_check(str(main.get("_gathering_spot")).is_empty(), "completed gathering clears progress with an explicit revision")
	var inventory_window: InventoryWindow = main.get("_inventory_window") as InventoryWindow
	var presentations: Dictionary = inventory_window._presentations
	_check(
		is_equal_approx(float((presentations[0] as Dictionary).get("cooldown_seconds", 0.0)), 0.6),
		"food displays the authoritative food cooldown",
	)
	_check(
		is_zero_approx(float((presentations[1] as Dictionary).get("cooldown_seconds", 0.0))),
		"food cooldown does not appear on potions",
	)


func _test_duplicate_login_submission_is_blocked() -> void:
	var gate: AuthGate = AUTH_GATE_SCENE.instantiate()
	root.add_child(gate)
	await process_frame
	var state: Dictionary = {"submissions": 0}
	gate.login_requested.connect(
		func(_email: String, _password: String) -> void: state["submissions"] += 1
	)
	gate.show_login()
	gate.email_input.text = "Arcane Hero"
	gate.password_input.text = "strong-password"
	gate._submit()
	gate._submit()
	_check(int(state["submissions"]) == 1, "duplicate login submission is ignored while busy")
	_check(gate.submit_button.disabled, "login controls remain locked during the request")
	gate.queue_free()
	await process_frame


func _test_registration_collects_public_character_name() -> void:
	var gate: AuthGate = AUTH_GATE_SCENE.instantiate()
	root.add_child(gate)
	await process_frame
	var registration: Dictionary = {"character_name": "", "email": ""}
	gate.registration_requested.connect(func(
		character_name: String, email: String, _password: String, _link_guest: bool,
	) -> void:
		registration["character_name"] = character_name
		registration["email"] = email
	)
	gate._set_registering(true)
	gate.character_name_input.text = "Arcane Hero"
	gate.email_input.text = "player@example.com"
	gate.password_input.text = "strong-password"
	gate.confirmation_input.text = "strong-password"
	gate._submit()
	_check(registration.get("character_name", "") == "Arcane Hero", "registration submits the public character name")
	_check(registration.get("email", "") == "player@example.com", "registration keeps email as the private account identifier")
	_check(gate._character_name_is_valid("Hero_42"), "registration accepts supported character name characters")
	_check(not gate._character_name_is_valid("-invalid"), "registration rejects names beginning with punctuation")
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


func _test_equipped_boots_follow_leg_pivots() -> void:
	var player := PLAYER_SCENE.instantiate() as PlayerView
	root.add_child(player)
	await process_frame
	player.set_armour({"feet": "swift_boots"})
	_check(player.swift_left_boot.visible and player.swift_right_boot.visible, "equipped boots are visible")
	_check(player.swift_left_boot.get_parent() == player.left_leg, "left boot is attached to the left leg pivot")
	_check(player.swift_right_boot.get_parent() == player.right_leg, "right boot is attached to the right leg pivot")
	player.queue_free()
	await process_frame


func _test_player_pvp_targeting_is_toggleable() -> void:
	var player := PLAYER_SCENE.instantiate() as PlayerView
	root.add_child(player)
	await process_frame
	var target_area := player.get_node("PvPTargetArea") as Area3D
	player.set_combat_target_id("opponent")
	player.set_pvp_targetable(true)
	_check(target_area.collision_layer == 8, "PvP opponent enables the combat target collision layer")
	_check(str(target_area.get_meta(&"combat_target_id")) == "opponent", "PvP target exposes its player ID")
	player.set_pvp_targetable(false)
	_check(target_area.collision_layer == 0, "leaving PvP disables player targeting")
	player.queue_free()
	await process_frame


func _test_equipped_cape_sits_behind_player() -> void:
	var player := PLAYER_SCENE.instantiate() as PlayerView
	root.add_child(player)
	await process_frame
	player.set_armour({"cape": "melee_cape"})
	_check(player.cape.visible, "equipped cape is visible")
	_check(player.cape.position.z < 0.0, "equipped cape is attached behind the player")
	player.queue_free()
	await process_frame


func _test_hit_splats_and_minimap_movement_facing() -> void:
	var player := PLAYER_SCENE.instantiate() as PlayerView
	root.add_child(player)
	await process_frame
	var hit_splat := player.get_node("HitSplat") as Node3D
	var value_label := hit_splat.get_node("Value") as Label3D
	player.play_hit(17, Color(0.9, 0.2, 0.15, 1.0))
	_check(hit_splat.visible, "damage displays a hit splat on the struck character")
	_check(value_label.text == "17", "damage hit splat displays the dealt amount without a minus prefix")
	_check(not hit_splat.has_node("Background"), "hit splat uses coloured text without a background")
	_check(value_label.font_size == 60, "damage hit splat uses large text")
	_check(value_label.modulate.r > 0.8, "damage amount inherits its combat feedback colour")
	_check(hit_splat.position.y < 1.6, "hit splat is positioned over the character rather than above it")
	player.play_dodge("MISS")
	_check(value_label.text == "MISS", "successful avoidance is presented as a miss")
	_check(value_label.font_size == 42, "miss feedback remains prominent without a background")
	player.global_position = Vector3.RIGHT
	player.call("_update_movement_pose", 0.1, Vector3.ZERO, Vector3.RIGHT, true)
	_check(player.minimap_facing().x > 0.9, "minimap arrow follows the player's movement direction")
	player.queue_free()
	await process_frame


func _test_action_rejection_feedback() -> void:
	var main: Node = MAIN_SCENE.instantiate()
	_check(bool(main.call("_gathering_result_is_rejection", "missing_axe")), "missing tools trigger rejected-action feedback")
	_check(bool(main.call("_gathering_result_is_rejection", "requires_woodcutting_level_5")), "level requirements trigger rejected-action feedback")
	_check(not bool(main.call("_gathering_result_is_rejection", "chopped_oak_logs")), "successful gathering does not trigger rejected-action feedback")
	var path_error := "find path: position { 1, 2, 3 } is not on a walkable layer"
	_check(
		str(main.call("_friendly_server_error", "command_rejected", path_error)) == "You cannot walk there.",
		"navigation diagnostics use a concise player-facing message",
	)
	main.set("_combat_debug_hud_visible", true)
	_check(
		str(main.call("_friendly_server_error", "command_rejected", path_error)) == path_error,
		"navigation diagnostics remain available in debug mode",
	)
	main.free()

	var effects := SoundEffects.new()
	root.add_child(effects)
	await process_frame
	effects.play_feedback(&"action_rejected")
	var first_feedback_msec: int = int(effects.get("_last_rejection_sound_msec"))
	effects.play_feedback(&"action_rejected")
	_check(first_feedback_msec >= 0, "rejected-action sound is available")
	_check(int(effects.get("_last_rejection_sound_msec")) == first_feedback_msec, "rejected-action sound is throttled against duplicate messages")
	for child: Node in effects.get_children():
		if child is AudioStreamPlayer:
			(child as AudioStreamPlayer).stop()
			(child as AudioStreamPlayer).stream = null
	effects.free()
	await process_frame


func _test_xp_tracker_supports_overlapping_drops() -> void:
	var tracker := XP_TRACKER_SCENE.instantiate() as XPTracker
	root.add_child(tracker)
	await process_frame
	var total_font_size: int = tracker.total_label.get_theme_font_size("font_size")
	var gain_font_size: int = tracker.gain_label.get_theme_font_size("font_size")
	var icon_size: Vector2 = tracker.skill_icon.custom_minimum_size
	_check(tracker is PanelContainer, "XP tracker has its own HUD background panel")
	_check(
		gain_font_size < total_font_size and icon_size.x < float(total_font_size),
		"XP gain text and icon are smaller than the persistent total XP counter",
	)
	tracker.show_gain("Attack", null, 12, 120)
	tracker.show_gain("Strength", null, 9, 84)
	var drops := tracker.get_node("DropLayer").get_children()
	_check(drops.size() == 2, "XP tracker keeps overlapping XP drops visible")
	_check(tracker.total_label.text == "84 XP", "XP tracker shows a compact value-first skill total")
	_check(tracker.total_skill_icon.tooltip_text == "Strength", "XP tracker identifies the total with its skill icon")
	_check((drops[0].get_node("Gain") as Label).text == "+12 XP", "XP drop keeps its own gained amount")
	_check(drops[1].position.y > drops[0].position.y, "rapid XP drops begin in separate visual lanes")
	tracker.queue_free()
	await process_frame


func _test_consumable_slot_shows_authoritative_cooldown() -> void:
	var slot := InventorySlot.new()
	root.add_child(slot)
	await process_frame
	slot.set_consumable_cooldown(1.2)
	_check(slot._cooldown_overlay.visible, "consumable slot displays its cooldown overlay")
	_check(slot._cooldown_label.text == "1.2s", "consumable slot displays cooldown seconds")
	slot.set_consumable_cooldown(0.0)
	_check(not slot._cooldown_overlay.visible, "consumable cooldown overlay clears when ready")
	slot.queue_free()
	await process_frame


func _test_shopkeeper_uses_icon_led_stock_panel() -> void:
	var shop := MERCHANT_WINDOW_SCENE.instantiate() as MerchantWindow
	root.add_child(shop)
	await process_frame
	shop.render_stock([{
		"item_id": "air_rune", "quantity": 7, "buy_price": 12,
		"instance_id": "", "reinforcement": 0,
	}])
	await process_frame
	var title := shop.get_node("Panel/Margin/VBox/Header/Title") as Label
	var grid := shop.get_node("Panel/Margin/VBox/Scroll/Grid") as GridContainer
	var stock_button := grid.get_child(0) as Button if grid.get_child_count() > 0 else null
	_check(title.text == "Shopkeeper", "shop window avoids buy-back terminology")
	_check(not shop.has_node("Panel/Margin/VBox/Hint"), "shop window removes permanent instruction text")
	_check(grid.columns == 5, "shop stock uses a compact five-column icon grid")
	_check(stock_button != null and stock_button.get_node("Stock").text == "×7", "shop slot displays its stock quantity")
	_check(
		stock_button != null and stock_button.get_theme_stylebox("normal") is StyleBoxEmpty,
		"shop item slots have no permanent square background",
	)
	_check(stock_button != null and not stock_button.tooltip_text.contains("12 coins"), "shop slot hides value until inspection")
	var inspected: Array = []
	shop.value_inspected.connect(func(item_id: String, price: int, quantity: int) -> void:
		inspected.append([item_id, price, quantity])
	)
	shop.set("_selected_item_id", "air_rune")
	shop.set("_selected_price", 12)
	shop.set("_selected_quantity", 7)
	shop.call("_on_context_action", 3)
	_check(inspected == [["air_rune", 12, 7]], "shop right-click inspection exposes value and stock")
	shop.queue_free()
	await process_frame


func _test_fletching_options_follow_clicked_item_and_inventory() -> void:
	var fletching := FletchingWindow.new()
	var recipes: Array = [
		{
			"id": "arrow_shafts", "required_level": 1, "required_tool": "knife",
			"costs": [{"item_id": "normal_logs", "quantity": 1}],
		},
		{
			"id": "unstrung_shortbow", "required_level": 1, "required_tool": "knife",
			"costs": [{"item_id": "normal_logs", "quantity": 1}],
		},
		{
			"id": "unstrung_oak_shortbow", "required_level": 2, "required_tool": "knife",
			"costs": [{"item_id": "oak_logs", "quantity": 1}],
		},
		{
			"id": "shortbow", "required_level": 1,
			"costs": [
				{"item_id": "unstrung_shortbow", "quantity": 1},
				{"item_id": "bowstring", "quantity": 1},
			],
		},
		{
			"id": "iron_arrows", "required_level": 2,
			"costs": [
				{"item_id": "arrow_shafts", "quantity": 20},
				{"item_id": "feathers", "quantity": 20},
				{"item_id": "iron_arrow_tips", "quantity": 20},
			],
		},
	]
	fletching.update_data(recipes, [
		{"item_id": "knife", "quantity": 1},
		{"item_id": "normal_logs", "quantity": 2},
	], 1, false, 0)
	var log_options := fletching.available_recipe_ids_for_item("normal_logs")
	_check(
		log_options.size() == 2 and log_options.has("arrow_shafts") and log_options.has("unstrung_shortbow"),
		"normal logs show only shafts and their unstrung bow",
	)

	fletching.update_data(recipes, [
		{"item_id": "unstrung_shortbow", "quantity": 1},
		{"item_id": "bowstring", "quantity": 1},
	], 2, false, 0)
	var bow_options := fletching.available_recipe_ids_for_item("unstrung_shortbow")
	_check(bow_options == ["shortbow"], "an unstrung bow shows only its matching stringing recipe")

	fletching.update_data(recipes, [
		{"item_id": "arrow_shafts", "quantity": 20},
		{"item_id": "feathers", "quantity": 20},
		{"item_id": "iron_arrow_tips", "quantity": 20},
	], 2, false, 0)
	var arrow_options := fletching.available_recipe_ids_for_item("arrow_shafts")
	_check(arrow_options == ["iron_arrows"], "arrow components show only fully makeable arrow recipes")
	fletching.free()


func _test_beta_build_uses_hosted_server() -> void:
	var local_url: String = NetworkClient.default_server_url(false)
	var beta_url: String = NetworkClient.default_server_url(true)
	_check(local_url.begins_with("ws://localhost:"), "development builds default to localhost")
	_check(beta_url.begins_with("wss://"), "beta builds default to a secure WebSocket endpoint")
	_check(not beta_url.contains("localhost"), "beta builds do not default to localhost")


func _test_inventory_hud_compact_layout() -> void:
	var main: Node = MAIN_SCENE.instantiate()
	var inventory_hud := main.get_node("UI/InventoryHUD") as Control
	var inventory_grid := main.get_node("UI/InventoryHUD/Margin/VBox/ContentScroll/Grid") as GridContainer
	var upper_tab := main.get_node("UI/MenuTabsHUD/Margin/Row/Inventory") as Button
	var lower_tab := main.get_node("UI/MenuUtilityTabsHUD/Margin/Row/Settings") as Button
	_check(
		inventory_hud.size.x <= 520.0 and inventory_hud.size.y <= 500.0,
		"inventory HUD has a smaller default footprint",
	)
	var shared_panel_size := inventory_hud.size
	var shared_panels_match := true
	for panel_name: String in [
		"CombatHUD", "SkillsHUD", "EquipmentHUD", "MagicHUD", "PrayerHUD", "StylesHUD", "SettingsHUD",
	]:
		shared_panels_match = shared_panels_match and (
			(main.get_node("UI/" + panel_name) as Control).size == shared_panel_size
		)
	_check(shared_panels_match, "all shared menu tabs use the same fixed panel size")
	_check(
		inventory_grid.get_theme_constant("h_separation") == 4
			and inventory_grid.get_theme_constant("v_separation") == 4,
		"inventory slots use compact row and column spacing",
	)
	_check(
		upper_tab.custom_minimum_size.y >= 62.0 and lower_tab.custom_minimum_size.y >= 62.0,
		"upper and lower menu tabs have larger click targets",
	)
	_check(
		float(main.call("_minimum_hud_scale", "menu")) <= 0.55
			and float(main.call("_minimum_hud_scale", "chat")) >= 0.75,
		"menu HUD can scale smaller without reducing other HUD readability limits",
	)
	main.free()


func _test_equipment_details_only_show_equipment_bonuses() -> void:
	var main: Node = MAIN_SCENE.instantiate()
	main.set("_combat_stats", {
		"weapon": "steel_sword",
		"style": "slash",
		"accuracy_level": 42,
		"damage_level": 37,
		"agility_level": 29,
		"weapon_accuracy_bonus": 8,
		"strength_bonus": 6,
		"ammunition_strength_bonus": 2,
		"attack_interval_modifier": -1,
		"equipment_offence": {
			"melee_accuracy": 5, "ranged_accuracy": 3, "magic_accuracy": 1,
			"melee_strength": 4, "ranged_strength": 2, "magic_strength": 1,
		},
		"defence_bonuses": {"stab": 10, "slash": 11, "crush": 9, "ranged": 8, "magic": 4},
		"equipment_evasion_bonus": 3,
		"maximum_health_bonus": 20,
		"armour_damage_reduction": 0.125,
	})
	var offence: String = str(main.call("_detailed_offence_text"))
	var defence: String = str(main.call("_detailed_defence_text"))
	var details_panel := main.get_node("UI/CombatStatsHUD") as Control
	_check(
		offence.contains("[table=3]") and offence.contains("GEAR ACCURACY")
			and offence.contains("GEAR POWER") and offence.contains("WEAPON & AMMO"),
		"equipment offence details use the available width across three equipment columns",
	)
	_check(
		not offence.contains("Lv") and not offence.contains("Style")
			and not offence.contains("TOTAL") and not offence.contains("DAMAGE RANGE"),
		"equipment offence details omit skill levels, attack styles, and derived combat totals",
	)
	_check(
		defence.contains("[table=3]") and defence.contains("EQUIPMENT UTILITY")
			and not defence.contains("Agility") and not defence.contains("Prayer"),
		"equipment defence details retain equipment bonuses without character levels or prayers",
	)
	_check(details_panel.size.x >= 900.0, "equipment details window uses a wider compact layout")
	main.free()


func _test_minimap_layout_and_projection() -> void:
	var main: Node = MAIN_SCENE.instantiate()
	var minimap: Control = main.get_node("UI/MinimapHUD") as Control
	var xp_tracker := main.get_node("UI/XPTrackerHUD") as Control
	_check(minimap != null, "main scene includes the minimap")
	_check(
		is_equal_approx(
			minimap.offset_right - minimap.offset_left,
			minimap.offset_bottom - minimap.offset_top,
		),
		"minimap has a square circular-drawing area",
	)
	_check(xp_tracker.offset_right <= minimap.offset_left - 8.0, "XP counter defaults to the left of the minimap")
	minimap.get_parent().remove_child(minimap)
	minimap.owner = null
	root.add_child(minimap)
	minimap.size = Vector2(196.0, 196.0)
	minimap.visible = true
	var markers: Array[Dictionary] = [
		{"position": Vector3(20.0, 2.0, 20.0), "kind": "target"},
		{"position": Vector3(10.0, 5.0, 10.0), "kind": "resource"},
	]
	minimap.call("set_world_state", Vector3(10.0, 2.0, 20.0), Vector2.UP, markers)
	var east_offset: Vector2 = minimap.call("map_offset_for", Vector3(46.0, 2.0, 20.0)) as Vector2
	var north_offset: Vector2 = minimap.call("map_offset_for", Vector3(10.0, 2.0, -16.0)) as Vector2
	_check(east_offset.x > 0.0 and is_zero_approx(east_offset.y), "minimap projects world east to screen right")
	_check(north_offset.y < 0.0 and is_zero_approx(north_offset.x), "minimap keeps world north at the top")
	await process_frame
	minimap.queue_free()
	main.free()
	await process_frame


func _test_skill_guides_show_authoritative_unlock_levels() -> void:
	var main: Node = MAIN_SCENE.instantiate()
	main.call("_on_skill_guide_received", [
		{"id": "oak", "skill": "woodcutting", "category": "gathering", "action": "Cut", "name": "Oak Tree", "required_level": 5},
		{"id": "staff", "skill": "magic", "category": "weapons", "action": "Wield", "name": "Mage Staff", "required_level": 10},
		{"id": "robe", "skill": "magic", "category": "armour", "action": "Wear", "name": "Mage Robe", "required_level": 5},
	])
	main.set("_production_recipes", {"fletching": [
		{"name": "Unstrung Shortbow", "required_level": 1, "reward_item": "unstrung_shortbow"},
		{"name": "Shortbow", "required_level": 1, "reward_item": "bow"},
		{"name": "Unstrung Oak Shortbow", "required_level": 2, "reward_item": "unstrung_oak_shortbow"},
		{"name": "Oak Shortbow", "required_level": 2, "reward_item": "oak_shortbow"},
		{"name": "20 Iron Arrows", "required_level": 2, "reward_item": "iron_arrow"},
		{"name": "20 Arrow Shafts", "required_level": 1, "reward_item": "arrow_shafts"},
		{"name": "20 Iron Bolts", "required_level": 3, "reward_item": "iron_bolt"},
	]})
	var test_spells: Array[Dictionary] = [{"name": "Fire Bolt", "required_magic_level": 7}]
	main.set("_spells", test_spells)
	var woodcutting_guide: String = str(main.call("_skill_guide_text", "woodcutting", 3))
	var magic_guide: String = str(main.call("_skill_guide_text", "magic", 8))
	var magic_spells: String = str(main.call("_skill_guide_text", "magic", 8, "spells"))
	var magic_weapons: String = str(main.call("_skill_guide_text", "magic", 8, "weapons"))
	var magic_armour: String = str(main.call("_skill_guide_text", "magic", 8, "armour"))
	var fletching_bows: String = str(main.call("_skill_guide_text", "fletching", 8, "bows"))
	var fletching_arrows: String = str(main.call("_skill_guide_text", "fletching", 8, "arrows"))
	var fletching_bolts: String = str(main.call("_skill_guide_text", "fletching", 8, "bolts"))
	_check(woodcutting_guide.contains("Lv 5") and woodcutting_guide.contains("Oak Tree"), "skill guide shows gathering unlock levels")
	_check(not woodcutting_guide.contains("Cut Oak Tree"), "skill guide rows omit redundant action verbs")
	_check(
		magic_guide.contains("[table=2]") and magic_guide.contains("[right]") and magic_guide.contains("[cell]  "),
		"skill guide uses an aligned level column with padded unlock names",
	)
	_check(magic_guide.contains("Lv 7") and magic_guide.contains("Fire Bolt"), "skill guide shows spell unlock levels")
	_check(magic_guide.contains("Lv 10") and magic_guide.contains("Mage Staff"), "skill guide shows equipment requirements")
	_check(
		magic_spells.contains("Fire Bolt") and not magic_spells.contains("Mage Staff"),
		"skill guide spell tab filters out equipment unlocks",
	)
	_check(
		magic_weapons.contains("Mage Staff") and not magic_weapons.contains("Mage Robe"),
		"combat skill guide separates weapons from armour",
	)
	_check(
		magic_armour.contains("Mage Robe") and not magic_armour.contains("Mage Staff"),
		"combat skill guide separates armour from weapons",
	)
	_check(
		fletching_bows.count("Shortbow") == 2
			and fletching_bows.contains("Lv 1") and fletching_bows.contains("Lv 2")
			and not fletching_bows.contains("Unstrung") and not fletching_bows.contains("Iron Arrows"),
		"fletching guide summarizes each complete bow tier once",
	)
	_check(
		fletching_arrows.contains("Iron Arrows") and not fletching_arrows.contains("Oak Shortbow"),
		"fletching guide has an arrows category",
	)
	_check(
		fletching_bolts.contains("Iron Bolts") and not fletching_bolts.contains("Arrow Shafts"),
		"fletching guide supports a bolts category",
	)
	main.free()


func _test_skill_slot_tooltip_and_hover_style() -> void:
	var main: Node = MAIN_SCENE.instantiate()
	var tooltip: String = str(main.call("_skill_tooltip_text", 125, 200))
	_check(tooltip.contains("XP to next level: 75"), "skill tooltip shows remaining XP to the next level")
	var skill_tile := Button.new()
	main.call("_remove_skill_tile_hover_effect", skill_tile)
	_check(
		skill_tile.get_theme_stylebox("hover") == skill_tile.get_theme_stylebox("normal"),
		"skill slot hover uses the normal background without a light-up effect",
	)
	var progress_bar := ProgressBar.new()
	main.call("_style_skill_progress_bar", progress_bar)
	var fill_style := progress_bar.get_theme_stylebox("fill") as StyleBoxFlat
	_check(fill_style.bg_color.g > fill_style.bg_color.r, "skill guide XP progress uses a green fill")
	progress_bar.free()
	skill_tile.free()
	main.free()


func _test_light_interface_theme_and_chest_slots() -> void:
	var debug_window := DebugSkillWindow.new()
	_check(debug_window.borderless, "test skill shrine suppresses the duplicate native close control")
	debug_window.free()

	var interface_theme := InterfaceStyle.create_theme()
	_check(
		interface_theme.get_color("font_color", "Label") == InterfaceStyle.TEXT_COLOR,
		"light interfaces use the shared dark navy text colour",
	)
	_check(
		InterfaceStyle.PANEL_COLOR.get_luminance() > 0.7,
		"interface panels use the lighter pastel palette",
	)

	var bank_window := BankWindow.new()
	var chest_slot := Button.new()
	bank_window.call("_style_chest_slot", chest_slot)
	_check(
		chest_slot.get_theme_stylebox("normal") is StyleBoxEmpty,
		"chest item slots have no permanent square background",
	)
	_check(
		chest_slot.get_theme_stylebox("hover") is StyleBoxFlat,
		"chest item slots retain visible hover feedback",
	)
	chest_slot.free()
	bank_window.free()


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
