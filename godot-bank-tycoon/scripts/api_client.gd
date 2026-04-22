class_name ApiClient
extends Node

@export var api_base_url: String = "http://127.0.0.1:8088"

func get_state(player_id: String) -> Dictionary:
	return await _request_json(HTTPClient.METHOD_GET, "/api/state", {}, player_id)

func do_action(player_id: String, action_type: String) -> Dictionary:
	var payload := {"type": action_type}
	return await _request_json(HTTPClient.METHOD_POST, "/api/action", payload, player_id)

func checkin(player_id: String) -> Dictionary:
	return await _request_json(HTTPClient.METHOD_POST, "/api/checkin", {}, player_id)

func _request_json(method: int, path: String, payload: Dictionary, player_id: String) -> Dictionary:
	var request := HTTPRequest.new()
	add_child(request)

	var headers := PackedStringArray([
		"Content-Type: application/json",
		"X-Player-Id: %s" % player_id
	])

	var body := ""
	if method != HTTPClient.METHOD_GET:
		body = JSON.stringify(payload)

	var err := request.request("%s%s" % [api_base_url, path], headers, method, body)
	if err != OK:
		request.queue_free()
		return {"error": "Request start failed: %s" % error_string(err)}

	var result: Array = await request.request_completed
	request.queue_free()

	var response_code: int = int(result[1])
	var raw_body: PackedByteArray = result[3]
	var text := raw_body.get_string_from_utf8()

	var parsed: Variant = JSON.parse_string(text)
	if typeof(parsed) != TYPE_DICTIONARY:
		return {"error": "Invalid JSON from server", "status": response_code}

	var data: Dictionary = parsed
	if response_code < 200 or response_code >= 300:
		if not data.has("error"):
			data["error"] = "HTTP %d" % response_code
		return data

	return data

