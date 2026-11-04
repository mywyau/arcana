class_name NetworkClient
extends Node

signal status_changed(status: String)
signal welcome_received(player_id: String)
signal snapshot_received(snapshot: Dictionary)
signal server_error(code: String, message: String)

@export var server_url: String = "ws://localhost:8080/ws"

var _socket: WebSocketPeer = WebSocketPeer.new()
var _last_state: WebSocketPeer.State = WebSocketPeer.STATE_CLOSED


func _ready() -> void:
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
	var state: WebSocketPeer.State = _socket.get_ready_state()
	if state != _last_state:
		_handle_state_change(state)
		_last_state = state

	if state != WebSocketPeer.STATE_OPEN:
		return

	while _socket.get_available_packet_count() > 0:
		_handle_packet(_socket.get_packet())


func send_move(sequence: int, destination: Vector3) -> bool:
	if _socket.get_ready_state() != WebSocketPeer.STATE_OPEN:
		status_changed.emit("Cannot move while disconnected")
		return false

	var message: Dictionary = {
		"type": "move_to",
		"sequence": sequence,
		"destination": {
			"x": destination.x,
			"z": destination.z,
		},
	}
	var result: Error = _socket.send_text(JSON.stringify(message))
	if result != OK:
		status_changed.emit("Could not send movement: %s" % error_string(result))
		return false
	return true


func _handle_state_change(state: WebSocketPeer.State) -> void:
	match state:
		WebSocketPeer.STATE_OPEN:
			status_changed.emit("Connected")
		WebSocketPeer.STATE_CLOSING:
			status_changed.emit("Disconnecting")
		WebSocketPeer.STATE_CLOSED:
			var reason: String = _socket.get_close_reason()
			if reason.is_empty():
				reason = "connection closed"
			status_changed.emit("Disconnected: %s" % reason)


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
				server_error.emit(
					str(details.get("code", "server_error")),
					str(details.get("message", "Unknown server error")),
				)
		_:
			server_error.emit("unknown_message", "Unknown server message: %s" % message_type)
