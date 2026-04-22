extends Node2D

const REFRESH_SECONDS := 2.0
const MAX_RENDERED_QUEUE := 20
const INTERACT_DISTANCE := 84.0

@onready var api: ApiClient = $ApiClient
@onready var player: CharacterBody2D = $Player
@onready var clients_root: Node2D = $Clients
@onready var queue_points_root: Node2D = $QueuePoints
@onready var spawn_point: Node2D = $SpawnPoint
@onready var service_zone: Node2D = $Zones/ServiceZone
@onready var atm_zone: Node2D = $Zones/ATMZone
@onready var vault_zone: Node2D = $Zones/VaultZone

@onready var money_value: Label = %MoneyValue
@onready var vault_value: Label = %VaultValue
@onready var queue_value: Label = %QueueValue
@onready var cashiers_value: Label = %CashiersValue
@onready var atms_value: Label = %AtmsValue
@onready var streak_value: Label = %StreakValue
@onready var prompt_value: Label = %PromptValue
@onready var status_label: Label = %StatusLabel
@onready var log_box: RichTextLabel = %LogBox
@onready var refresh_timer: Timer = $RefreshTimer

var player_id: String = ""
var state := GameState.new()
var busy: bool = false
var queue_clients: Array[Node2D] = []
var queue_points: Array[Vector2] = []

func _ready() -> void:
	player_id = _load_or_create_player_id()
	_collect_queue_points()
	_append_log("Player ID: %s" % player_id)
	_update_status("Connecting...")
	_refresh_state()
	refresh_timer.wait_time = REFRESH_SECONDS
	refresh_timer.start()

func _process(_delta: float) -> void:
	_update_prompt()

func _unhandled_input(event: InputEvent) -> void:
	if event.is_action_pressed("ui_accept"):
		if _is_near(service_zone.global_position):
			_do_action("serve_manual", "Served one client at front desk")
		elif _is_near(vault_zone.global_position):
			_do_action("collect_vault", "Collected money from vault")
		elif _is_near(atm_zone.global_position):
			_do_action("buy_atm", "Purchased and installed ATM")

func _on_refresh_timer_timeout() -> void:
	_refresh_state()

func _on_checkin_button_pressed() -> void:
	if busy:
		return
	busy = true
	_update_status("Daily check-in...")
	var response: Dictionary = await api.checkin(player_id)
	busy = false

	if response.has("error"):
		_append_log("[color=red]Check-in error: %s[/color]" % response["error"])
		_update_status("Error")
		return

	var did_checkin := bool(response.get("didCheckin", false))
	var reward_money := int(response.get("rewardMoney", 0))
	var body: Dictionary = response.get("state", {})
	state.update_from_dict(body)
	_render_state()

	if did_checkin:
		_append_log("🔥 Daily reward: +%d" % reward_money)
	else:
		_append_log("Daily reward already collected")

func _on_buy_cashier_button_pressed() -> void:
	_do_action("buy_cashier", "Hired one cashier")

func _on_upgrade_vault_button_pressed() -> void:
	_do_action("upgrade_vault", "Vault capacity upgraded")

func _do_action(action_type: String, success_log: String) -> void:
	if busy:
		return
	busy = true
	_update_status("Action: %s" % action_type)
	var response: Dictionary = await api.do_action(player_id, action_type)
	busy = false

	if response.has("error"):
		_append_log("[color=red]Action error: %s[/color]" % response["error"])
		_update_status("Error")
		return

	state.update_from_dict(response)
	_render_state()
	_append_log(success_log)
	_update_status("Ready")

func _refresh_state() -> void:
	if busy:
		return
	busy = true
	_update_status("Sync...")
	var response: Dictionary = await api.get_state(player_id)
	busy = false

	if response.has("error"):
		_append_log("[color=red]Sync error: %s[/color]" % response["error"])
		_update_status("Error")
		return

	state.update_from_dict(response)
	_render_state()
	_update_status("Ready")

func _render_state() -> void:
	money_value.text = str(state.money)
	vault_value.text = "%d / %d" % [state.vault_stored, state.vault_cap]
	queue_value.text = str(state.queue)
	cashiers_value.text = str(state.cashiers)
	atms_value.text = str(state.atms)
	streak_value.text = "%d 🔥" % state.streak
	_sync_queue_visual()

func _sync_queue_visual() -> void:
	var target_count := mini(state.queue, MAX_RENDERED_QUEUE)

	while queue_clients.size() < target_count:
		var client := _spawn_client()
		queue_clients.append(client)

	while queue_clients.size() > target_count:
		var c := queue_clients.pop_back()
		c.queue_free()

	for i in range(queue_clients.size()):
		var c: Node2D = queue_clients[i]
		if i < queue_points.size():
			c.set("target_position", queue_points[i])

func _spawn_client() -> Node2D:
	var client_node := Node2D.new()
	client_node.set_script(load("res://scripts/client.gd"))
	client_node.global_position = spawn_point.global_position
	clients_root.add_child(client_node)

	var body := Polygon2D.new()
	body.polygon = PackedVector2Array([
		Vector2(-12, -16),
		Vector2(12, -16),
		Vector2(12, 16),
		Vector2(-12, 16),
	])
	body.color = Color(0.90, 0.74, 0.36, 1.0)
	client_node.add_child(body)

	var head := Polygon2D.new()
	head.polygon = PackedVector2Array([
		Vector2(-8, -28),
		Vector2(8, -28),
		Vector2(8, -16),
		Vector2(-8, -16),
	])
	head.color = Color(0.98, 0.85, 0.70, 1.0)
	client_node.add_child(head)

	return client_node

func _collect_queue_points() -> void:
	queue_points.clear()
	for child in queue_points_root.get_children():
		if child is Node2D:
			queue_points.append((child as Node2D).global_position)

func _update_prompt() -> void:
	if _is_near(service_zone.global_position):
		prompt_value.text = "SPACE/ENTER: Serve client"
	elif _is_near(vault_zone.global_position):
		prompt_value.text = "SPACE/ENTER: Collect vault"
	elif _is_near(atm_zone.global_position):
		prompt_value.text = "SPACE/ENTER: Buy ATM"
	else:
		prompt_value.text = "Move with arrows/WASD"

func _is_near(point: Vector2) -> bool:
	return player.global_position.distance_to(point) <= INTERACT_DISTANCE

func _update_status(text: String) -> void:
	status_label.text = "Status: %s" % text

func _append_log(line: String) -> void:
	log_box.append_text("%s\n" % line)

func _load_or_create_player_id() -> String:
	var cfg := ConfigFile.new()
	var err := cfg.load("user://player.cfg")
	if err == OK:
		var existing := str(cfg.get_value("player", "id", ""))
		if existing != "":
			return existing

	var random := RandomNumberGenerator.new()
	random.randomize()
	var created := "p_%08x_%d" % [random.randi(), Time.get_unix_time_from_system()]
	cfg.set_value("player", "id", created)
	cfg.save("user://player.cfg")
	return created

func mini(a: int, b: int) -> int:
	if a < b:
		return a
	return b

