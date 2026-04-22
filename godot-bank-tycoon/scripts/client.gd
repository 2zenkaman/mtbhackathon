extends Node2D

@export var move_speed: float = 140.0
var target_position: Vector2

func _ready() -> void:
	target_position = global_position

func _process(delta: float) -> void:
	global_position = global_position.move_toward(target_position, move_speed * delta)

