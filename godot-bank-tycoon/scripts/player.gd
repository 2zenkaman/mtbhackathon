extends CharacterBody2D

@export var move_speed: float = 260.0

func _physics_process(_delta: float) -> void:
	var input_vec := Input.get_vector("ui_left", "ui_right", "ui_up", "ui_down")
	velocity = input_vec * move_speed
	move_and_slide()

