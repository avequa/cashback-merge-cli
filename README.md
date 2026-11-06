# Домашнее задание №2: Cashback Merge

Описание задачи: [TASK.md](TASK.md)

Консольная мини-игра в духе 2048 на поле 4×4 с плитками кэшбэка (`1% + 1% = 2%`, ..., `16% + 16% = 32%`)
и batch-симуляция кампании: сколько игроков доберётся до каждого уровня награды и укладывается ли доля
максимальной награды в допустимый лимит

## Команды

| Команда                   | Что делает                                                     | Без make                                                   |
| ------------------------- | -------------------------------------------------------------- | ---------------------------------------------------------- |
| `make install`            | скачать зависимости и локальный линтер                         | `go mod download`                                          |
| `make build`              | собрать исполняемый файл `bin/cashback-merge`                  | `go build -trimpath -o bin/cashback-merge ./cmd/cashback-merge` |
| `make run ARGS="..."`     | запустить приложение с аргументами                             | `go run ./cmd/cashback-merge ...`                          |
| `make test`               | запустить тесты, создать `coverage.out` и `coverage.html`      | `go test -coverprofile=coverage.out ./...`                 |
| `make test-race`          | тесты с race detector (нужен CGO и C-компилятор)               | `go test -race ./...`                                      |
| `make check`              | проверить `go.mod`, форматирование и golangci-lint             | см. `Makefile`                                             |
| `make format`             | отформатировать код (gofumpt, goimports)                       | `go tool golangci-lint fmt`                                |

Первый запуск:

```bash
make install
make build
make test
```

После `make build` приложение можно запускать напрямую: `./bin/cashback-merge <команда> [флаги]`.

## Запуск

Справка по командам:

```bash
make run ARGS="--help"
```

### `new` - начальное поле

```bash
make run ARGS="new --config campaign.json --seed 42"
```

| Флаг       | Обязательный | По умолчанию | Описание                    |
| ---------- | ------------ | ------------ | --------------------------- |
| `--config` | да           | -            | путь к JSON-файлу кампании  |
| `--seed`   | нет          | `0`          | seed для начального поля    |

Команда загружает и проверяет конфигурацию кампании, создаёт начальное поле с двумя плитками по seed и печатает его:

```text
Moves: 0/40  Score: 0
     .     .     .     .
     .     .     .    1%
     .     .     .     .
     .     .     .    1%
```

Одинаковые конфигурация и seed всегда дают одинаковое поле.

### `simulate` - batch-симуляция и отчёт

```bash
make run ARGS="simulate --config campaign.json --players 10000 --strategy random --seed 42 --output report.md"
```

| Флаг         | Обязательный | По умолчанию | Описание                                                  |
| ------------ | ------------ | ------------ | --------------------------------------------------------- |
| `--config`   | да           | -            | путь к JSON-файлу кампании                                |
| `--players`  | нет          | `1000`       | число игроков (партий), целое `> 0`                       |
| `--strategy` | нет          | `random`     | стратегия выбора хода                                     |
| `--seed`     | нет          | `0`          | общий seed симуляции                                      |
| `--output`   | нет          | -            | файл Markdown-отчёта; без флага отчёт печатается в stdout |

С `--output` в консоль выводится короткий итог, а полный отчёт сохраняется в файл:

```text
Players: 10000  Max reward rate: 0.0926 / 0.2000  Status: OK
Report: report.md
```

Одинаковые конфигурация, seed, число игроков и стратегия всегда дают одинаковый отчёт.

### Ошибки

Некорректная конфигурация отклоняется до начала игры или симуляции с указанием поля и причины:

```text
error: load campaign: validate config: invalid campaign: movesLimit must be positive
```

Неверные флаги печатают ошибку и подсказку по командам, код завершения `1`:

```text
error: --players must be positive
error: unknown strategy "greedy"
error: --config is required
```

## Тесты

```bash
make test
make test-race
make check
```

## Конфигурация кампании

Пример - `campaign.json`

```json
{
  "movesLimit": 40,
  "spawnRules": [
    {"cashback": "1", "probability": "0.85"},
    {"cashback": "2", "probability": "0.15"}
  ],
  "maxRewardCashback": "32",
  "maxRewardAllowedRate": "0.20",
  "rewards": [
    {"minTile": "1", "name": "basic"},
    {"minTile": "4", "name": "bronze"},
    {"minTile": "8", "name": "silver"},
    {"minTile": "16", "name": "gold"},
    {"minTile": "32", "name": "platinum"}
  ]
}
```

| Поле                   | Описание                                                     | Ограничения                                       |
| ---------------------- | ------------------------------------------------------------ | ------------------------------------------------- |
| `movesLimit`           | лимит валидных ходов                                         | целое `> 0`                                       |
| `spawnRules`           | значения новых плиток и вероятность каждого                  | не пусто, каждая вероятность в `(0, 1]`, сумма ровно `1` |
| `maxRewardCashback`    | плитка, с которой игрок получает максимальную награду        | корректный cashback                               |
| `maxRewardAllowedRate` | допустимая доля игроков с максимальной наградой              | от `0` до `1`                                     |
| `rewards`              | таблица наград                                               | см. ниже                                          |

Cashback - целое число процентов строкой: `"1"`, `"16"`. Допустимы только положительные
степени двойки.

Вероятности и доли записаны строками и разбираются сразу в `big.Rat`, без `float64`.

`movesLimit = 40` подобран симуляцией: при `random` на 10000 игроках максимальную награду получают около 9%
игроков, это укладывается в лимит 20%. С `movesLimit = 100` её получают 92.72% игроков, и симуляция
возвращает `FAILED`.

### Таблица наград

`rewards` - список уровней, каждый уровень:

- `minTile` - максимальная плитка, начиная с которой игрок получает этот уровень
- `name` - название уровня, не пустое

Игрок получает уровень с наибольшим `minTile`, который не больше его максимальной плитки.
По таблице выше максимальная плитка `16%` даёт `gold`, `8%` - `silver`, `32%` и выше - `platinum`.

Чтобы уровень по любой максимальной плитке определялся однозначно, таблица проверяется при загрузке:

- `minTile` строго возрастает - подходящий уровень всегда ровно один
- нижний `minTile` не больше наименьшего значения из `spawnRules` - уровень есть у любой партии
- верхний `minTile` равен `maxRewardCashback` - верхний уровень и есть максимальная награда

## Структура

| Пакет                        | Назначение                                                                  |
| ---------------------------- | --------------------------------------------------------------------------- |
| `cmd/cashback-merge`         | точка входа                                                                 |
| `internal/cli`               | разбор аргументов и флагов, вывод через `io.Writer`, сохранение отчёта      |
| `internal/config`            | чтение JSON-файла и перевод в `domain.Campaign`                             |
| `internal/core/domain`       | кэшбэк, поле, состояние партии, сдвиг и слияние, кампания, её проверка и таблица наград |
| `internal/core/apperr`       | типизированные доменные ошибки, проверяются через `errors.As`               |
| `internal/features/game`     | начальное состояние, ход, появление плиток, стратегии, партия целиком       |
| `internal/features/simulation` | batch-симуляция, статистика, статус и Markdown-отчёт                     |

## Ход

`ApplyMove(state, direction)` в `internal/features/game` выполняет ход `UP`, `DOWN`, `LEFT` или `RIGHT`:

1. `Board.MergeBoardAfterMove` в `internal/core/domain` - чистый сдвиг и слияние без случайности: 
плитки сдвигаются к краю, соседние равные сливаются, плитка сливается не больше одного раза за ход:
`[1, 1, 2, _] -> [2, 2, _, _]`
2. если поле не изменилось - ход не засчитывается: лимит не расходуется, плитка не появляется, score не меняется
3. если изменилось - в случайной свободной клетке появляется одна плитка по `spawnRules`, `movesUsed` увеличивается, score растёт на сумму плиток, созданных слияниями: `[2, 2, 4, 4] -> [4, 8, _, _]` даёт `+12`

Партия завершается, когда израсходован `movesLimit` или на поле не осталось хода, который его меняет. Плитка 32% игру не завершает.

## Появление новой плитки

`spawnTile` в `internal/features/game` выбирает случайную свободную клетку и значение по `spawnRules`.
Значение выбирается без `float`: вероятности приводятся к общему знаменателю и превращаются в целые веса.
Для `0.85` и `0.15` это `17/20` и `3/20`: генератор выдаёт число от `0` до `19`, значения `0..16` дают `1%`,
`17..19` - `2%`.

## Награда

`Board.MaxTile` находит максимальную плитку на поле, `Campaign.RewardFor` - уровень награды по таблице `rewards`.
Игрок получил максимальную награду, если его максимальная плитка не меньше `maxRewardCashback` (`Campaign.IsMaxReward`).

## Симуляция

`simulation.Run` в `internal/features/simulation` играет `players` партий выбранной стратегией через `game.Play`
и собирает статистику: распределение максимальных плиток и наград, суммарный и максимальный score,
число игроков с максимальной наградой.

- стратегия `random` (`internal/features/game/strategy.go`) равновероятно выбирает одно из направлений, которые меняют поле
- seed каждой партии берётся по порядку из генератора, созданного по общему `--seed`, поэтому результат воспроизводим
- партии не печатаются, вывод только в итоговом отчёте

Статус считается по фактической доле игроков с максимальной наградой:

```text
actualMaxRewardRate <= maxRewardAllowedRate -> OK
actualMaxRewardRate >  maxRewardAllowedRate -> FAILED
```

Это эмпирическая проверка для конкретных конфигурации, seed, числа игроков и стратегии, а не математическая гарантия.

## Отчёт

`simulation.WriteMarkdown` формирует Markdown-отчёт. Плитки выводятся по возрастанию, уровни наград - в порядке таблицы, доли считаются через `big.Rat`. Полный пример - report.md

```markdown
## Распределение максимальных плиток

| Максимальная плитка | Игроков | Доля |
| --- | ---: | ---: |
| 4% | 1 | 0.01% |
| 8% | 422 | 4.22% |
| 16% | 8651 | 86.51% |
| 32% | 926 | 9.26% |

## Распределение наград

| Награда | Минимальная плитка | Игроков | Доля |
| --- | --- | ---: | ---: |
| basic | 1% | 0 | 0.00% |
| bronze | 4% | 1 | 0.01% |
| silver | 8% | 422 | 4.22% |
| gold | 16% | 8651 | 86.51% |
| platinum | 32% | 926 | 9.26% |

## Score

| Показатель | Значение |
| --- | ---: |
| Средний score | 123.38 |
| Максимальный score | 204 |

## Ограничение максимальной награды

| Показатель | Значение |
| --- | ---: |
| Игроков с максимальной наградой | 926 |
| Фактическая доля | 9.26% |
| Допустимая доля | 20.00% |
| Статус | **OK** |
```

## Бонусные задания

### Стратегия `greedy-score`

```bash
make run ARGS="simulate --config campaign.json --players 10000 --strategy greedy-score --seed 42"
```

Из ходов, которые меняют поле, выбирает ход с наибольшим числом очков за слияния, при равных очках - ход,
после которого больше свободных клеток, затем первый в порядке `UP`, `DOWN`, `LEFT`, `RIGHT`

Сравнение с `random` на одинаковых seed - compare-strategies.md:

```bash
make run ARGS="compare --config campaign.json --strategy random --strategy-b greedy-score --players 10000 --seed 42 --output compare-strategies.md"
```

### Генеративные тесты

`internal/features/bonus/merge_generative_test.go`, запускаются вместе со всеми тестами через `make test` или отдельно:

```bash
go test -v -run 'TestMergeGenerative|TestPlaySameSeedSameGame' ./internal/features/bonus
```

Что проверяют:

- на случайных полях от 2×2 до 8×8 сумма плиток не меняется при сдвиге и слиянии, одна плитка
  не объединяется дважды, score за ход равен сумме плиток, полученных слиянием
- повторный запуск партии с тем же seed даёт тот же результат

Поля генерируются с фиксированным seed, поэтому тесты стабильны

### Сравнение кампаний

```bash
make run ARGS="compare --config campaign.json --config-b campaign-100.json --players 10000 --seed 42 --output compare.md"
```

| Флаг           | Обязательный | По умолчанию | Описание                                   |
| -------------- | ------------ | ------------ | ------------------------------------------ |
| `--config`     | да           | -            | конфиг кампании A                          |
| `--config-b`   | нет          | `--config`   | конфиг кампании B                          |
| `--strategy`   | нет          | `random`     | стратегия варианта A                       |
| `--strategy-b` | нет          | `--strategy` | стратегия варианта B                       |
| `--players`    | нет          | `1000`       | число игроков в каждом варианте            |
| `--seed`       | нет          | `0`          | общий seed обоих вариантов                 |
| `--output`     | нет          | -            | файл Markdown-отчёта; без флага - в stdout |

Оба варианта играются на одинаковых seed, отчёт содержит средний и максимальный score, распределение наград,
долю максимальной награды и статус каждого варианта, пример - compare.md 

`campaign-100.json` - `campaign.json` с `movesLimit = 100`, отчёт симуляции для него:

```bash
make run ARGS="simulate --config campaign-100.json --players 10000 --strategy random --seed 42 --output report-100.md"
```

### Размер поля

Размер задаётся необязательным полем `boardSize` в конфиге кампании, допустимо от 2 до 8, без поля - 4.
`campaign-5x5.json` - `campaign.json` с `"boardSize": 5`:

```bash
make run ARGS="new --config campaign-5x5.json --seed 42"
make run ARGS="play --config campaign-5x5.json --seed 42"
make run ARGS="simulate --config campaign-5x5.json --players 10000 --strategy random --seed 42"
```

### Консольная игра

```bash
make run ARGS="play --config campaign.json --seed 42"
```

| Флаг       | Обязательный | По умолчанию | Описание                                  |
| ---------- | ------------ | ------------ | ----------------------------------------- |
| `--config` | да           | -            | путь к JSON-файлу кампании                |
| `--seed`   | нет          | `0`          | seed партии                               |
| `--save`   | нет          | -            | сохранить seed и ходы партии для `replay` |

Ходы: `w`/`up`, `s`/`down`, `a`/`left`, `d`/`right`, выход - `q`. После каждого хода выводятся поле,
оставшиеся ходы, score и текущая награда

### Воспроизведение партии

```bash
make run ARGS="play --config campaign.json --seed 42 --save game.json"
make run ARGS="replay --config campaign.json --input game.json"
```

`play --save` сохраняет seed и ходы, изменившие поле:

```json
{
  "seed": 42,
  "moves": ["UP", "LEFT", "DOWN"]
}
```

`replay` проигрывает партию заново и выводит состояние после каждого хода, промежуточные состояния и результат совпадают с исходной партией

# обычные запуски
OK: make run ARGS="simulate --config campaign.json --players 10000 --strategy random --seed 42 --output report.md"
FAIL: make run ARGS="simulate --config campaign-100.json --players 10000 --strategy random --seed 42 --output report-100.md"

UNCORRECT INP: make run ARGS="simulate --config campaign.json --players 0"

# стратегия greedy-score и сравнение с random на одинаковых seed

make run ARGS="simulate --config campaign.json --players 10000 --strategy greedy-score --seed 42"

make run ARGS="compare --config campaign.json --strategy random --strategy-b greedy-score --players 10000 --seed 42 --output compare-strategies.md"

# генеративные тесты
go test -v -run 'TestMergeGenerative|TestPlaySameSeedSameGame' ./internal/features/bonus

# сравнение кампаний
make run ARGS="compare --config campaign.json --config-b campaign-100.json --players 10000 --seed 42 --output compare.md"

# размер поля
make run ARGS="new --config campaign-5x5.json --seed 42"
make run ARGS="simulate --config campaign-5x5.json --players 10000 --strategy random --seed 42"

# консольная игра
make run ARGS="play --config campaign.json --seed 42"

# replay: сыграть пару ходов, выйти по q, затем воспроизвести
make run ARGS="play --config campaign.json --seed 42 --save game.json"
make run ARGS="replay --config campaign.json --input game.json"
rm game.json
