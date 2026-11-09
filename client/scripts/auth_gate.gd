class_name AuthGate
extends Control

signal login_requested(email: String, password: String)
signal registration_requested(email: String, password: String, link_guest: bool)
signal guest_requested

@onready var title_label: Label = $Backdrop/Center/Card/Margin/Content/Title
@onready var login_mode_button: Button = $Backdrop/Center/Card/Margin/Content/Modes/Login
@onready var register_mode_button: Button = $Backdrop/Center/Card/Margin/Content/Modes/Register
@onready var email_input: LineEdit = $Backdrop/Center/Card/Margin/Content/Email
@onready var password_input: LineEdit = $Backdrop/Center/Card/Margin/Content/Password
@onready var confirmation_input: LineEdit = $Backdrop/Center/Card/Margin/Content/Confirmation
@onready var link_guest_checkbox: CheckBox = $Backdrop/Center/Card/Margin/Content/LinkGuest
@onready var submit_button: Button = $Backdrop/Center/Card/Margin/Content/Submit
@onready var feedback_label: Label = $Backdrop/Center/Card/Margin/Content/Feedback
@onready var guest_separator: HSeparator = $Backdrop/Center/Card/Margin/Content/GuestSeparator
@onready var guest_button: Button = $Backdrop/Center/Card/Margin/Content/Guest
@onready var guest_note: Label = $Backdrop/Center/Card/Margin/Content/GuestNote

var _registering: bool = false
var _busy: bool = false
var _has_existing_guest: bool = false


func _ready() -> void:
	login_mode_button.pressed.connect(_set_registering.bind(false))
	register_mode_button.pressed.connect(_set_registering.bind(true))
	submit_button.pressed.connect(_submit)
	guest_button.pressed.connect(_request_guest)
	email_input.text_submitted.connect(_submit_from_text)
	password_input.text_submitted.connect(_submit_from_text)
	confirmation_input.text_submitted.connect(_submit_from_text)
	_set_registering(false)


func configure(allow_development_guest: bool, has_existing_guest: bool) -> void:
	_has_existing_guest = has_existing_guest
	guest_separator.visible = allow_development_guest
	guest_button.visible = allow_development_guest
	guest_note.visible = allow_development_guest
	guest_button.text = (
		"Continue with existing development guest"
		if has_existing_guest
		else "Continue as development guest"
	)
	_set_registering(_registering)


func show_login(message: String = "") -> void:
	visible = true
	_busy = false
	_set_controls_enabled(true)
	show_feedback(message, not message.is_empty())
	password_input.clear()
	confirmation_input.clear()
	email_input.grab_focus.call_deferred()


func show_connecting(message: String = "Connecting to the game server…") -> void:
	visible = true
	_busy = true
	_set_controls_enabled(false)
	show_feedback(message)


func show_error(message: String) -> void:
	visible = true
	_busy = false
	_set_controls_enabled(true)
	show_feedback(message, true)
	password_input.clear()
	confirmation_input.clear()
	password_input.grab_focus.call_deferred()


func dismiss() -> void:
	visible = false
	_busy = false
	password_input.clear()
	confirmation_input.clear()


func show_feedback(message: String, is_error: bool = false) -> void:
	feedback_label.text = message
	feedback_label.modulate = Color(1.0, 0.42, 0.42) if is_error else Color(0.72, 0.82, 0.92)


func _set_registering(registering: bool) -> void:
	if _busy:
		return
	_registering = registering
	login_mode_button.button_pressed = not registering
	register_mode_button.button_pressed = registering
	title_label.text = "Create your account" if registering else "Welcome back"
	confirmation_input.visible = registering
	link_guest_checkbox.visible = registering and _has_existing_guest
	submit_button.text = "Create account" if registering else "Log in"
	show_feedback(
		"Your existing local guest character will be linked to this account."
		if registering and _has_existing_guest
		else ""
	)


func _submit_from_text(_text: String) -> void:
	_submit()


func _submit() -> void:
	if _busy:
		return
	var email: String = email_input.text.strip_edges()
	var password: String = password_input.text
	if email.is_empty() or not email.contains("@"):
		show_error("Enter a valid email address.")
		return
	if password.length() < 8:
		show_error("Password must be at least 8 characters.")
		return
	if _registering and password != confirmation_input.text:
		show_error("The passwords do not match.")
		return
	_busy = true
	_set_controls_enabled(false)
	show_feedback("Creating account…" if _registering else "Logging in…")
	if _registering:
		registration_requested.emit(email, password, link_guest_checkbox.button_pressed)
	else:
		login_requested.emit(email, password)


func _request_guest() -> void:
	if _busy:
		return
	_busy = true
	_set_controls_enabled(false)
	show_feedback("Connecting to the local development server…")
	guest_requested.emit()


func _set_controls_enabled(enabled: bool) -> void:
	login_mode_button.disabled = not enabled
	register_mode_button.disabled = not enabled
	email_input.editable = enabled
	password_input.editable = enabled
	confirmation_input.editable = enabled
	link_guest_checkbox.disabled = not enabled
	submit_button.disabled = not enabled
	guest_button.disabled = not enabled
