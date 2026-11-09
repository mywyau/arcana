class_name AccountWindow
extends Window

signal registration_requested(email: String, password: String)
signal login_requested(email: String, password: String)
signal logout_requested

@onready var email_input: LineEdit = $Panel/Margin/VBox/Email
@onready var password_input: LineEdit = $Panel/Margin/VBox/Password
@onready var confirm_password_input: LineEdit = $Panel/Margin/VBox/ConfirmPassword
@onready var feedback: Label = $Panel/Margin/VBox/Feedback
@onready var close_button: Button = $Panel/Margin/VBox/Header/Close
@onready var register_button: Button = $Panel/Margin/VBox/Actions/Register
@onready var login_button: Button = $Panel/Margin/VBox/Actions/Login
@onready var logout_button: Button = $Panel/Margin/VBox/Actions/Logout


func _ready() -> void:
	close_requested.connect(hide)
	close_button.pressed.connect(hide)
	register_button.pressed.connect(_request_registration)
	login_button.pressed.connect(_request_login)
	logout_button.pressed.connect(logout_requested.emit)


func open(has_saved_session: bool) -> void:
	feedback.text = (
		"A saved account session is active for this server."
		if has_saved_session
		else "Link this character or recover an existing account."
	)
	feedback.modulate = Color(0.82, 0.86, 0.9)
	set_session_active(has_saved_session)
	password_input.clear()
	confirm_password_input.clear()
	popup_centered()
	email_input.grab_focus()


func set_session_active(active: bool) -> void:
	register_button.disabled = active
	logout_button.disabled = not active
	register_button.tooltip_text = (
		"This profile already has an account session for the current server."
		if active
		else "Link the character currently being played to a new account."
	)
	logout_button.tooltip_text = (
		"Revoke this device session and reconnect as a guest."
		if active
		else "No account session is active for this server."
	)


func show_feedback(message: String, is_error: bool = false, clear_passwords: bool = false) -> void:
	feedback.text = message
	feedback.modulate = Color(1.0, 0.38, 0.35) if is_error else Color(0.4, 0.9, 0.55)
	if visible and (is_error or clear_passwords):
		password_input.clear()
		confirm_password_input.clear()


func _request_registration() -> void:
	if not _validate_inputs(true):
		return
	show_feedback("Creating account…")
	registration_requested.emit(email_input.text.strip_edges(), password_input.text)


func _request_login() -> void:
	if not _validate_inputs(false):
		return
	show_feedback("Logging in…")
	login_requested.emit(email_input.text.strip_edges(), password_input.text)


func _validate_inputs(require_confirmation: bool) -> bool:
	if email_input.text.strip_edges().is_empty():
		show_feedback("Enter your email address.", true)
		return false
	if password_input.text.length() < 8:
		show_feedback("Passwords must contain at least 8 characters.", true)
		return false
	if require_confirmation and password_input.text != confirm_password_input.text:
		show_feedback("The passwords do not match.", true)
		return false
	return true


func _unhandled_key_input(event: InputEvent) -> void:
	if visible and event.is_action_pressed("ui_cancel"):
		hide()
		get_viewport().set_input_as_handled()
