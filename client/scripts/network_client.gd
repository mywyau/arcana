class_name NetworkClient
extends Node

const GUEST_IDENTITY_PATH: String = "user://guest_identity.cfg"

signal status_changed(status: String)
signal welcome_received(player_id: String)
signal snapshot_received(snapshot: Dictionary)
signal server_error(code: String, message: String)

@export var server_url: String = "ws://localhost:8080/ws"

var _socket: WebSocketPeer = WebSocketPeer.new()
var _last_state: WebSocketPeer.State = WebSocketPeer.STATE_CLOSED
var _guest_token: String = ""
var _retry_without_guest_token: bool = false


func _ready() -> void:
	_load_guest_token()
	_connect_to_server.call_deferred()


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


func send_equip_item(sequence: int, item_id: String) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot equip while disconnected")
		return false

	var message: Dictionary = {
		"type": "equip_item",
		"sequence": sequence,
		"item_id": item_id,
	}
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not equip item: %s" % error_string(result))
		return false
	return true


func send_unequip_item(sequence: int, item_id: String) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot unequip while disconnected")
		return false

	var message: Dictionary = {
		"type": "unequip_item",
		"sequence": sequence,
		"item_id": item_id,
	}
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not unequip item: %s" % error_string(result))
		return false
	return true


func send_drop_item(sequence: int, item_id: String) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot drop items while disconnected")
		return false

	var message: Dictionary = {
		"type": "drop_item",
		"sequence": sequence,
		"item_id": item_id,
	}
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


func send_consume_item(sequence: int, item_id: String) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot consume items while disconnected")
		return false

	var message: Dictionary = {
		"type": "consume_item",
		"sequence": sequence,
		"item_id": item_id,
	}
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not consume item: %s" % error_string(result))
		return false
	return true


func send_gather(sequence: int, spot_id: String) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot gather while disconnected")
		return false

	var message: Dictionary = {
		"type": "gather",
		"sequence": sequence,
		"spot_id": spot_id,
	}
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not start gathering: %s" % error_string(result))
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
	if config.load(GUEST_IDENTITY_PATH) != OK:
		return
	_guest_token = str(config.get_value("identity", "guest_token", ""))


func _save_guest_token() -> void:
	var config := ConfigFile.new()
	config.set_value("identity", "guest_token", _guest_token)
	var result: Error = config.save(GUEST_IDENTITY_PATH)
	if result != OK:
		status_changed.emit("Could not save guest identity: %s" % error_string(result))
