class_name GameState
extends RefCounted

var money: int = 0
var vault_stored: int = 0
var vault_cap: int = 0
var cashiers: int = 0
var atms: int = 0
var queue: int = 0
var streak: int = 0
var last_checkin_day: String = ""
var last_seen_unix: int = 0

func update_from_dict(data: Dictionary) -> void:
	money = int(data.get("money", money))
	vault_stored = int(data.get("vaultStored", vault_stored))
	vault_cap = int(data.get("vaultCap", vault_cap))
	cashiers = int(data.get("cashiers", cashiers))
	atms = int(data.get("atms", atms))
	queue = int(data.get("queue", queue))
	streak = int(data.get("streak", streak))
	last_checkin_day = str(data.get("lastCheckinDay", last_checkin_day))
	last_seen_unix = int(data.get("lastSeenUnix", last_seen_unix))

