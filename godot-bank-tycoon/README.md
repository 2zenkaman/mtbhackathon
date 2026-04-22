## Godot Bank Tycoon (skeleton)

Каркас клиента на Godot 4 для демо-игры.

### Что внутри
- `project.godot` — проект.
- `scenes/world.tscn` — игровая 2D-сцена банка.
- `scripts/world.gd` — игровой цикл, интеракции и синхронизация с backend.
- `scripts/player.gd` — движение персонажа.
- `scripts/client.gd` — движение NPC-клиентов в очереди.
- `scripts/api_client.gd` — HTTP-запросы к backend.
- `scripts/game_state.gd` — модель состояния игрока.

### Запуск
1. Открой папку `godot-bank-tycoon` в Godot 4.2+.
2. Перед запуском подними backend:
   - `cd ../game-demo`
   - `go run ./cmd/server`
3. Нажми Play в Godot.

По умолчанию клиент ходит на `http://127.0.0.1:8088`.

### Управление
- Стрелки/WASD — движение персонажа.
- `Space`/`Enter` рядом с зонами:
  - касса — `Serve Manual`
  - сейф — `Collect Vault`
  - банкомат — `Buy ATM`
- Кнопки HUD:
  - `Daily Check-in`
  - `Buy Cashier`
  - `Upgrade Vault`

