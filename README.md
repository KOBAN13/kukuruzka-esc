# Kukuruzka-ESC

`Kukuruzka-ESC` - экспериментальное ECS-ядро на Go.

Проект реализует базовые сущности Entity Component System: мир, сущности, компоненты, архетипы, запросы, системы, буфер команд и ресурсы. Основной код библиотеки находится в пакете [`esc_core`](./esc_core).

> Статус: проект находится в активной разработке. Публичный API уже намечен, но тестовое покрытие и демонстрационный пример пока не оформлены. Примеры ниже показывают предполагаемый стиль использования текущих типов.

## Возможности

- хранение сущностей через `Entity` с индексом и поколением;
- регистрация компонентов по Go-типам;
- архетипное хранение компонентов;
- создание, удаление и изменение сущностей через `Spawn`, `Despawn`, `Add`, `Remove`, `Set`;
- чтение компонентов через `Get`, `Has` и итераторы запросов;
- deferred-изменения через `CommandBuffer`;
- системный раннер со стадиями выполнения;
- проверка конфликтов доступа систем к компонентам;
- типизированное хранилище ресурсов;
- отладочные отчеты по архетипам, запросам и доступам.

## Требования

- Go `1.26` или новее, согласно [`go.mod`](./go.mod).

## Структура проекта

```text
.
├── esc_core/          # ECS-ядро
├── main.go            # текущая стартовая заглушка
├── go.mod             # Go-модуль
├── LICENSE            # MIT License
└── README.md
```

Ключевые файлы пакета `esc_core`:

- `world.go` - создание мира и операции над сущностями;
- `component.go` - компоненты, registry и ошибки;
- `query.go`, `iterator.go` - описание и выполнение запросов;
- `command.go`, `*_command_builder.go` - буфер команд;
- `runner.go`, `system.go` - запуск систем по стадиям;
- `resources.go` - глобальные ресурсы;
- `access.go` - модель доступа и поиск конфликтов;
- `bundle.go` - группировка компонентов;
- `debug.go` - структуры отладочных отчетов.

## Быстрый старт

Установите зависимости и проверьте сборку:

```bash
go test ./...
```

Запуск текущей стартовой программы:

```bash
go run .
```

## Минимальный пример API

```go
package main

import (
	"fmt"

	ecs "Kukuruzka-ESC/esc_core"
)

type Position struct {
	X float32
	Y float32
}

type Velocity struct {
	X float32
	Y float32
}

func main() {
	world := ecs.NewWorld(ecs.WithEntityCapacity(128))

	entity, err := ecs.Spawn(
		world,
		Position{X: 10, Y: 20},
		Velocity{X: 1, Y: 0},
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(entity)
}
```

## Основные сущности

### World

`World` хранит сущности, архетипы, registry компонентов и текущую фазу мутаций.

Создание мира:

```go
world := esc_core.NewWorld()
worldWithCapacity := esc_core.NewWorld(esc_core.WithEntityCapacity(1024))
```

### Components

Компонентом считается Go-структура. Пустые структуры могут использоваться как tag-компоненты.

```go
type Player struct{}

type Health struct {
	Value int
}
```

Токен компонента:

```go
token := esc_core.Component[Health]()
```

### Entity Operations

Основные операции над сущностями:

- `Spawn(world, components...)` - создать сущность с набором компонентов;
- `Despawn(world, entity)` - удалить сущность;
- `Add(world, entity, components...)` - добавить компоненты;
- `Remove(world, entity, componentTokens...)` - удалить компоненты;
- `Has[T](world, entity)` - проверить наличие компонента;
- `Get[T](world, entity)` - прочитать компонент;
- `Set[T](world, entity, value)` - заменить значение компонента;
- `IsAlive(world, entity)` - проверить, что сущность ещё существует.

### Queries

Запросы описывают, какие компоненты нужны для чтения или записи, и по каким компонентам сущности должны фильтроваться.

`QueryBuilder` поддерживает фильтры `With` и `Without`, а также декларацию доступа через `Read` и `Write`. После `Build` запрос можно обходить через `query.Iter()`, `Iterator.Next()` и `Iterator.Entity()`.

Для доступа к компонентам внутри итератора предусмотрены функции:

- `Read[T](it)` - чтение компонента;
- `Write[T](it)` - получение указателя на компонент для изменения.

### CommandBuffer

`CommandBuffer` нужен для отложенных изменений мира, например во время работы систем.

```go
commands := &esc_core.CommandBuffer{}

err := commands.Spawn().
	With(Position{}).
	With(Velocity{}).
	Commit()
if err != nil {
	panic(err)
}

err = commands.Apply(world)
commands.Clear()
```

### Runner и System

Система должна реализовать интерфейс `System`:

```go
type System interface {
	Name() string
	Stage() StageID
	Update(ctx *Context) error
	Access() AccessSet
	DebugQueries() []QueryDebugInfo
}
```

`Runner` выполняет системы по стадиям и применяет накопленные команды после каждой стадии:

```go
const UpdateStage esc_core.StageID = 1

runner := esc_core.NewRunner([]esc_core.StageID{UpdateStage})
runner.Add(mySystem)

ctx := &esc_core.Context{
	World:     world,
	Commands: &esc_core.CommandBuffer{},
	Resources: esc_core.NewResources(),
}

err := runner.ValidateAccess()
if err != nil {
	panic(err)
}

err = runner.Update(ctx)
```

## Ресурсы

Ресурсы используются для хранения глобального состояния, не привязанного к конкретной сущности.

Доступные операции:

- `NewResources()` - создать контейнер ресурсов;
- `PutResources[T](resources, value)` - сохранить ресурс;
- `GetResources[T](resources)` - получить ресурс;
- `RemoveResources[T](resources)` - удалить ресурс.

## Отладка

Доступные отладочные методы:

```go
world.DebugArchetypes()
runner.DebugAccess()
runner.DebugQueries()
```

Они помогают посмотреть текущие архетипы, доступы систем и состав запросов.

## Разработка

Полезные команды:

```bash
gofmt -w .
go test ./...
go run .
```

В текущем окружении команда `go test ./...` не была выполнена, потому что бинарник `go` недоступен.

## Лицензия

Проект распространяется под лицензией MIT. Подробности в [`LICENSE`](./LICENSE).
