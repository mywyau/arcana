class_name NetworkClient
extends Node

signal status_changed(status: String)
signal welcome_received(player_id: String)
signal production_recipes_received(recipes: Array)
signal spell_catalog_received(books: Array, spells: Array)
signal prayer_catalog_received(prayers: Array, prayer_sets: Array)
signal snapshot_received(snapshot: Dictionary)
signal server_error(code: String, message: String)

@export var server_url: String = ""

var _socket: WebSocketPeer = WebSocketPeer.new()
var _last_state: WebSocketPeer.State = WebSocketPeer.STATE_CLOSED
var _guest_token: String = ""
var _retry_without_guest_token: bool = false
var _guest_identity_path: String = ClientProfile.config_path("guest_identity.cfg")
var _guest_identity_section: String = ""


func _ready() -> void:
	_configure_server_url()
	_guest_identity_section = ClientProfile.server_identity_section(server_url)
	_load_guest_token()
	_connect_to_server.call_deferred()


func _configure_server_url() -> void:
	if server_url.is_empty():
		server_url = str(ProjectSettings.get_setting("network/server_url", "ws://localhost:8080/ws"))
	var environment_url := OS.get_environment("ARCANA_SERVER_URL").strip_edges()
	if not environment_url.is_empty():
		server_url = environment_url
	for argument: String in OS.get_cmdline_user_args():
		if argument.begins_with("--server-url="):
			server_url = argument.trim_prefix("--server-url=").strip_edges()
	if not server_url.begins_with("ws://") and not server_url.begins_with("wss://"):
		push_error("Configured server URL must begin with ws:// or wss://")
		server_url = "ws://localhost:8080/ws"


func _connect_to_server() -> void:
	_socket.heartbeat_interval = 10.0
	var result: Error = _socket.connect_to_url(server_url)
	if result != OK:
		status_changed.emit("Connection failed: %s" % error_string(result))
		return
	_last_state = WebSocketPeer.STATE_CONNECTING
	status_changed.emit("Connecting to %s" % server_url)


func _process(_delta: float) -> void:
	_socket.poll()
	# A server may send a final error packet immediately before closing. Drain
	# available packets first so an invalid saved guest token can be cleared and
	# retried instead of losing the error when the closed socket is replaced.
	while _socket.get_available_packet_count() > 0:
		_handle_packet(_socket.get_packet())

	var state: WebSocketPeer.State = _socket.get_ready_state()
	if state != _last_state:
		_handle_state_change(state)
		_last_state = state


func send_move(sequence: int, destination: Vector3) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot move while disconnected")
		return false

	var message: Dictionary = {
		"type": "move_to",
		"sequence": sequence,
		"destination": {
			"x": destination.x,
			"y": destination.y,
			"z": destination.z,
		},
	}
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not send movement: %s" % error_string(result))
		return false
	return true


func send_engage_target(sequence: int, target_id: String) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot engage a target while disconnected")
		return false

	var message: Dictionary = {
		"type": "engage_target",
		"sequence": sequence,
		"target_id": target_id,
	}
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not engage target: %s" % error_string(result))
		return false
	return true


func send_disengage(sequence: int) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot disengage while disconnected")
		return false

	var message: Dictionary = {
		"type": "disengage",
		"sequence": sequence,
	}
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not disengage: %s" % error_string(result))
		return false
	return true


func send_experience_lock(sequence: int, locked: bool) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot change XP settings while disconnected")
		return false
	var result: Error = _socket.send_text(JSON.stringify({
		"type": "set_experience_lock",
		"sequence": sequence,
		"active": locked,
	}))
	if result != OK:
		status_changed.emit("Could not change XP settings: %s" % error_string(result))
		return false
	return true


func send_debug_set_skill(sequence: int, skill_id: String, level: int, experience: int = 0, set_experience: bool = false) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot adjust test skills while disconnected")
		return false
	var result: Error = _socket.send_text(JSON.stringify({
		"type": "debug_set_skill",
		"sequence": sequence,
		"skill_id": skill_id,
		"level": level,
		"experience": experience,
		"set_experience": set_experience,
	}))
	if result != OK:
		status_changed.emit("Could not adjust test skill: %s" % error_string(result))
		return false
	return true


func send_equip_item(sequence: int, item_id: String, instance_id: String = "") -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot equip while disconnected")
		return false

	var message: Dictionary = {
		"type": "equip_item",
		"sequence": sequence,
		"item_id": item_id,
	}
	if not instance_id.is_empty():
		message["instance_id"] = instance_id
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not equip item: %s" % error_string(result))
		return false
	return true


func send_unequip_item(sequence: int, item_id: String, instance_id: String = "") -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot unequip while disconnected")
		return false

	var message: Dictionary = {
		"type": "unequip_item",
		"sequence": sequence,
		"item_id": item_id,
	}
	if not instance_id.is_empty():
		message["instance_id"] = instance_id
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not unequip item: %s" % error_string(result))
		return false
	return true


func send_drop_item(sequence: int, item_id: String, instance_id: String = "") -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot drop items while disconnected")
		return false

	var message: Dictionary = {
		"type": "drop_item",
		"sequence": sequence,
		"item_id": item_id,
	}
	if not instance_id.is_empty():
		message["instance_id"] = instance_id
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not drop item: %s" % error_string(result))
		return false
	return true


func send_move_inventory_item(sequence: int, from_slot: int, to_slot: int) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot move inventory items while disconnected")
		return false

	var message: Dictionary = {
		"type": "move_inventory_item",
		"sequence": sequence,
		"from_slot": from_slot,
		"to_slot": to_slot,
	}
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not move inventory item: %s" % error_string(result))
		return false
	return true


func send_consume_item(sequence: int, item_id: String, inventory_slot: int) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot consume items while disconnected")
		return false

	var message: Dictionary = {
		"type": "consume_item",
		"sequence": sequence,
		"item_id": item_id,
		"inventory_slot": inventory_slot,
	}
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not consume item: %s" % error_string(result))
		return false
	return true


func send_brew(sequence: int, item_ids: Array[String], quantity: int = 1) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot brew while disconnected")
		return false
	var message: Dictionary = {
		"type": "brew",
		"sequence": sequence,
		"item_ids": item_ids,
		"quantity": quantity,
	}
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not brew: %s" % error_string(result))
		return false
	return true


func send_fletch(sequence: int, recipe_id: String, quantity: int = 1) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot fletch while disconnected")
		return false
	var message: Dictionary = {
		"type": "fletch",
		"sequence": sequence,
		"recipe_id": recipe_id,
		"quantity": quantity,
	}
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not fletch: %s" % error_string(result))
		return false
	return true


func send_gather(sequence: int, spot_id: String, recipe_id: String = "", instance_id: String = "", quantity: int = 0) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot gather while disconnected")
		return false

	var message: Dictionary = {
		"type": "gather",
		"sequence": sequence,
		"spot_id": spot_id,
	}
	if not recipe_id.is_empty():
		message["recipe_id"] = recipe_id
	if not instance_id.is_empty():
		message["instance_id"] = instance_id
	if quantity > 0:
		message["quantity"] = quantity
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not start gathering: %s" % error_string(result))
		return false
	return true


func send_bank_deposit(sequence: int, item_id: String, inventory_slot: int, instance_id: String = "") -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot use the bank while disconnected")
		return false
	var message: Dictionary = {
		"type": "bank_deposit",
		"sequence": sequence,
		"item_id": item_id,
		"inventory_slot": inventory_slot,
	}
	if not instance_id.is_empty():
		message["instance_id"] = instance_id
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not deposit item: %s" % error_string(result))
		return false
	return true


func send_bank_withdraw(sequence: int, bank_slot: int, quantity: int = 1) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot use the bank while disconnected")
		return false
	var message: Dictionary = {
		"type": "bank_withdraw",
		"sequence": sequence,
		"bank_slot": bank_slot,
		"quantity": quantity,
	}
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not withdraw item: %s" % error_string(result))
		return false
	return true


func send_bank_expand(sequence: int) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot expand the bank while disconnected")
		return false
	var result: Error = _socket.send_text(JSON.stringify({
		"type": "bank_expand", "sequence": sequence,
	}))
	if result != OK:
		status_changed.emit("Could not expand bank: %s" % error_string(result))
		return false
	return true


func send_merchant_sell(sequence: int, inventory_slot: int) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot use the shop while disconnected")
		return false
	var result: Error = _socket.send_text(JSON.stringify({
		"type": "merchant_sell", "sequence": sequence, "inventory_slot": inventory_slot,
	}))
	if result != OK:
		status_changed.emit("Could not sell item: %s" % error_string(result))
		return false
	return true


func send_ammunition_deposit(sequence: int, inventory_slot: int) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot use the ammunition pouch while disconnected")
		return false
	var result: Error = _socket.send_text(JSON.stringify({
		"type": "ammunition_deposit",
		"sequence": sequence,
		"inventory_slot": inventory_slot,
	}))
	if result != OK:
		status_changed.emit("Could not store ammunition: %s" % error_string(result))
		return false
	return true


func send_ammunition_withdraw(sequence: int, ammunition_slot: int) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot use the ammunition pouch while disconnected")
		return false
	var result: Error = _socket.send_text(JSON.stringify({
		"type": "ammunition_withdraw",
		"sequence": sequence,
		"ammunition_slot": ammunition_slot,
	}))
	if result != OK:
		status_changed.emit("Could not withdraw ammunition: %s" % error_string(result))
		return false
	return true


func send_ring_deposit(sequence: int, item_id: String, instance_id: String) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot change rings while disconnected")
		return false
	var result: Error = _socket.send_text(JSON.stringify({
		"type": "ring_deposit", "sequence": sequence,
		"item_id": item_id, "instance_id": instance_id,
	}))
	if result != OK:
		status_changed.emit("Could not equip ring: %s" % error_string(result))
		return false
	return true


func send_ring_withdraw(sequence: int, ring_slot: int) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot change rings while disconnected")
		return false
	var result: Error = _socket.send_text(JSON.stringify({
		"type": "ring_withdraw", "sequence": sequence, "ring_slot": ring_slot,
	}))
	if result != OK:
		status_changed.emit("Could not unequip ring: %s" % error_string(result))
		return false
	return true


func send_rune_deposit(sequence: int, inventory_slot: int) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot use the rune pouch while disconnected")
		return false
	var result: Error = _socket.send_text(JSON.stringify({
		"type": "rune_deposit", "sequence": sequence, "inventory_slot": inventory_slot,
	}))
	if result != OK:
		status_changed.emit("Could not store runes: %s" % error_string(result))
		return false
	return true


func send_rune_withdraw(sequence: int, rune_slot: int) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot use the rune pouch while disconnected")
		return false
	var result: Error = _socket.send_text(JSON.stringify({
		"type": "rune_withdraw", "sequence": sequence, "rune_slot": rune_slot,
	}))
	if result != OK:
		status_changed.emit("Could not withdraw runes: %s" % error_string(result))
		return false
	return true


func send_combat_style(sequence: int, combat_style: String) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot change combat style while disconnected")
		return false

	var message: Dictionary = {
		"type": "set_combat_style",
		"sequence": sequence,
		"combat_style": combat_style,
	}
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not change combat style: %s" % error_string(result))
		return false
	return true


func send_prayer(sequence: int, prayer_id: String, active: bool) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot change prayer while disconnected")
		return false

	var message: Dictionary = {
		"type": "set_prayer",
		"sequence": sequence,
		"prayer_id": prayer_id,
		"active": active,
	}
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not change prayer: %s" % error_string(result))
		return false
	return true


func send_cut_gem(sequence: int, item_id: String) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot cut a gem while disconnected")
		return false
	var result: Error = _socket.send_text(JSON.stringify({
		"type": "cut_gem", "sequence": sequence, "item_id": item_id,
	}))
	if result != OK:
		status_changed.emit("Could not cut gem: %s" % error_string(result))
		return false
	return true


func send_light_fire(sequence: int, log_item: String) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot light a fire while disconnected")
		return false
	var result: Error = _socket.send_text(JSON.stringify({
		"type": "light_fire", "sequence": sequence, "item_id": log_item,
	}))
	if result != OK:
		status_changed.emit("Could not light fire: %s" % error_string(result))
		return false
	return true


func send_spell(sequence: int, spell_id: String) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot select a spell while disconnected")
		return false

	var message: Dictionary = {
		"type": "set_spell",
		"sequence": sequence,
		"spell_id": spell_id,
	}
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not select spell: %s" % error_string(result))
		return false
	return true


func send_pickup_item(sequence: int, drop_id: String) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot pick up loot while disconnected")
		return false

	var message: Dictionary = {
		"type": "pickup_item",
		"sequence": sequence,
		"drop_id": drop_id,
	}
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not pick up loot: %s" % error_string(result))
		return false
	return true


func _handle_state_change(state: WebSocketPeer.State) -> void:
	match state:
		WebSocketPeer.STATE_OPEN:
			if _send_authentication():
				status_changed.emit("Connected — authenticating")
			else:
				_socket.close(1008, "could not authenticate")
		WebSocketPeer.STATE_CLOSING:
			status_changed.emit("Disconnecting")
		WebSocketPeer.STATE_CLOSED:
			var reason: String = _socket.get_close_reason()
			if reason.is_empty():
				reason = "connection closed"
			status_changed.emit("Disconnected: %s" % reason)
			if _retry_without_guest_token:
				_retry_without_guest_token = false
				_socket = WebSocketPeer.new()
				_connect_to_server.call_deferred()


func _handle_packet(packet: PackedByteArray) -> void:
	var parser := JSON.new()
	var result: Error = parser.parse(packet.get_string_from_utf8())
	if result != OK:
		server_error.emit("invalid_json", parser.get_error_message())
		return

	var parsed: Variant = parser.data
	if typeof(parsed) != TYPE_DICTIONARY:
		server_error.emit("invalid_message", "Server message must be an object")
		return

	var message: Dictionary = parsed
	var message_type: String = str(message.get("type", ""))
	match message_type:
		"welcome":
			var issued_token: String = str(message.get("guest_token", ""))
			if not issued_token.is_empty():
				_guest_token = issued_token
				_save_guest_token()
			var production_recipes: Variant = message.get("production_recipes", [])
			if typeof(production_recipes) == TYPE_ARRAY:
				production_recipes_received.emit(production_recipes as Array)
			var spellbooks: Variant = message.get("spellbooks", [])
			var spells: Variant = message.get("spells", [])
			if typeof(spellbooks) == TYPE_ARRAY and typeof(spells) == TYPE_ARRAY:
				spell_catalog_received.emit(spellbooks as Array, spells as Array)
			var prayers: Variant = message.get("prayers", [])
			var prayer_sets: Variant = message.get("prayer_sets", [])
			if typeof(prayers) == TYPE_ARRAY and typeof(prayer_sets) == TYPE_ARRAY:
				prayer_catalog_received.emit(prayers as Array, prayer_sets as Array)
			welcome_received.emit(str(message.get("player_id", "")))
		"snapshot":
			var snapshot: Variant = message.get("snapshot")
			if typeof(snapshot) == TYPE_DICTIONARY:
				snapshot_received.emit(snapshot as Dictionary)
			else:
				server_error.emit("invalid_snapshot", "Snapshot payload is missing")
		"error":
			var error_payload: Variant = message.get("error", {})
			if typeof(error_payload) == TYPE_DICTIONARY:
				var details: Dictionary = error_payload
				var error_code: String = str(details.get("code", "server_error"))
				if error_code == "invalid_guest_token":
					_guest_token = ""
					_save_guest_token()
					_retry_without_guest_token = true
				server_error.emit(
					error_code,
					str(details.get("message", "Unknown server error")),
				)
		_:
			server_error.emit("unknown_message", "Unknown server message: %s" % message_type)


func _send_authentication() -> bool:
	var message: Dictionary = {"type": "authenticate"}
	if not _guest_token.is_empty():
		message["guest_token"] = _guest_token
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not authenticate: %s" % error_string(result))
		return false
	return true


func _load_guest_token() -> void:
	var config := ConfigFile.new()
	if config.load(_guest_identity_path) != OK:
		return
	_guest_token = str(config.get_value(_guest_identity_section, "guest_token", ""))
	if (
		_guest_token.is_empty()
		and ClientProfile.is_local_server(server_url)
		and config.has_section_key(ClientProfile.LEGACY_IDENTITY_SECTION, "guest_token")
	):
		# Guest tokens previously belonged implicitly to localhost. Migrate that
		# value only for local connections so it is never presented to a host.
		_guest_token = str(
			config.get_value(ClientProfile.LEGACY_IDENTITY_SECTION, "guest_token", "")
		)
		if not _guest_token.is_empty():
			_save_guest_token()


func _save_guest_token() -> void:
	var config := ConfigFile.new()
	var load_result: Error = config.load(_guest_identity_path)
	if load_result != OK and load_result != ERR_FILE_NOT_FOUND:
		status_changed.emit("Could not read guest identities: %s" % error_string(load_result))
		return
	config.set_value(_guest_identity_section, "server_url", server_url)
	config.set_value(_guest_identity_section, "guest_token", _guest_token)
	var result: Error = config.save(_guest_identity_path)
	if result != OK:
		status_changed.emit("Could not save guest identity: %s" % error_string(result))
