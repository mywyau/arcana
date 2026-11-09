class_name AccountWindow
extends Window

signal session_action_requested

@onready var description: Label = $Panel/Margin/VBox/Description
@onready var identity: Label = $Panel/Margin/VBox/Identity
@onready var feedback: Label = $Panel/Margin/VBox/Feedback
@onready var close_button: Button = $Panel/Margin/VBox/Header/Close
@onready var session_action_button: Button = $Panel/Margin/VBox/SessionAction


func _ready() -> void:
	close_requested.connect(hide)
	close_button.pressed.connect(hide)
	session_action_button.pressed.connect(session_action_requested.emit)


func open(has_account_session: bool, account_email: String = "") -> void:
	set_session(has_account_session, account_email)
	popup_centered()


func set_session(has_account_session: bool, account_email: String = "") -> void:
	feedback.text = ""
	if has_account_session:
		description.text = "This character is protected by an account."
		identity.text = (
			"Signed in as %s" % account_email
			if not account_email.is_empty()
			else "Signed in with a saved account session"
		)
		session_action_button.text = "Log out"
		session_action_button.tooltip_text = (
			"Revoke this device session and return to the login screen."
		)
	else:
		description.text = "You are playing with a local development guest."
		identity.text = "Guest characters can be linked when creating an account."
		session_action_button.text = "Return to login"
		session_action_button.tooltip_text = "Leave the world without deleting this guest character."


func show_feedback(message: String, is_error: bool = false) -> void:
	feedback.text = message
	feedback.modulate = Color(1.0, 0.38, 0.35) if is_error else Color(0.4, 0.9, 0.55)


func _unhandled_key_input(event: InputEvent) -> void:
	if visible and event.is_action_pressed("ui_cancel"):
		hide()
		get_viewport().set_input_as_handled()
