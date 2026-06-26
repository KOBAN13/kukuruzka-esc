# Kukuruzka-ECS

`Kukuruzka-ECS` - экспериментальное ECS-ядро на Go.

Проект реализует базовые части Entity Component System: мир, сущности,
компоненты, архетипное хранение, запросы, системы, буфер команд и ресурсы.
Основной код находится в пакете [`ecs`](./ecs).

> Статус: активная разработка. API уже намечен, но проект пока не имеет
> тестового покрытия, релизных тегов и полноценного демонстрационного примера.
> Некоторые экспортированные части остаются черновыми и могут требовать правок
> перед использованием в реальном приложении. [`main.go`](./main.go) сейчас
> является стартовой IDE-заглушкой, а не примером использования ECS.

## Возможности

- сущности `Entity` с индексом и поколением;
- регистрация компонентов по Go-типам;
- хранение компонентов по архетипам;
- операции `Spawn`, `Despawn`, `Add`, `Remove`, `Has`, `Get`, `Set`;
- отложенные изменения мира через `CommandBuffer`;
- запросы через `Query`, `Iterator`, `Read` и `Write`;
- запуск систем по стадиям через `Runner`;
- декларация доступа систем к компонентам через `AccessSet`;
- контейнер ресурсов `Resources`;
- отладочные отчеты по архетипам, запросам и доступам.

## Требования

- Go `1.26` или новее, согласно [`go.mod`](./go.mod).

## Установка

Сейчас пакет не стоит подключать через прямой `go get`: путь модуля в
[`go.mod`](./go.mod) и URL текущего репозитория ещё не синхронизированы.

Надёжный вариант для локальной разработки - клонировать репозиторий и
подключить его через `replace` в проекте-потребителе:

```bash
git clone https://github.com/KOBAN13/kukuruzka-esc.git
```

В `go.mod` проекта-потребителя:

```go
require github.com/KOBAN13/Kukuruzka-ECS v0.0.0

replace github.com/KOBAN13/Kukuruzka-ECS => ../kukuruzka-esc
```

Импорт пакета остаётся таким же, как путь модуля:

```go
import ecs "github.com/KOBAN13/Kukuruzka-ECS/ecs"
```

После публикации репозитория по тому же пути, который указан в `go.mod`,
локальный `replace` можно будет убрать и заменить подключение на обычный
`go get`.

## Структура

```text
.
├── ecs/        # ECS-ядро
├── main.go     # текущая стартовая заглушка
├── go.mod      # Go-модуль
├── LICENSE     # MIT License
└── README.md
```

Ключевые файлы пакета `ecs`:

- [`world.go`](./ecs/world.go) - мир и основные операции над сущностями;
- [`entity.go`](./ecs/entity.go) - тип `Entity`;
- [`component.go`](./ecs/component.go) - компоненты, registry и ошибки;
- [`arhetype.go`](./ecs/arhetype.go), [`column.go`](./ecs/column.go) - архетипы и колонки компонентов;
- [`query.go`](./ecs/query.go), [`iterator.go`](./ecs/iterator.go) - запросы и итерация;
- [`command.go`](./ecs/command.go), [`*_command_builder.go`](./ecs) - буфер команд;
- [`runner.go`](./ecs/runner.go), [`system.go`](./ecs/system.go) - системы и стадии выполнения;
- [`access.go`](./ecs/access.go) - модель доступа и конфликты;
- [`resources.go`](./ecs/resources.go) - глобальные ресурсы;
- [`debug.go`](./ecs/debug.go) - структуры отладочных отчетов.

## Быстрый старт

Минимальный пример создания сущности:

```go
package main

import (
	"fmt"

	ecs "github.com/KOBAN13/Kukuruzka-ECS/ecs"
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

Проверка проекта:

```bash
go test ./...
```

Запуск текущей стартовой программы:

```bash
go run .
```

## Основной API

### World

`World` хранит сущности, архетипы, registry компонентов и текущую фазу
мутаций.

```go
world := ecs.NewWorld()
worldWithCapacity := ecs.NewWorld(ecs.WithEntityCapacity(1024))
```

### Components

Компонентом должен быть Go-тип со `struct`-kind. Пустые структуры можно
использовать как tag-компоненты.

```go
type Player struct{}

type Health struct {
	Value int
}

healthToken := ecs.Component[Health]()
```

### Entity Operations

- `Spawn(world, components...)` - создать сущность;
- `SpawnBundle(world, bundle)` - создать сущность из bundle;
- `Despawn(world, entity)` - удалить сущность;
- `Add(world, entity, components...)` - добавить компоненты;
- `Remove(world, entity, componentTokens...)` - удалить компоненты;
- `Has[T](world, entity)` - проверить наличие компонента;
- `Get[T](world, entity)` - получить копию компонента;
- `GetWrite[T](world, entity)` - получить компонент для записи;
- `Set[T](world, entity, value)` - заменить компонент;
- `IsAlive(world, entity)` - проверить, что сущность жива.

Пример изменения компонента:

```go
err := ecs.Set(world, entity, Position{X: 15, Y: 20})
if err != nil {
	panic(err)
}
```

### CommandBuffer

`CommandBuffer` накапливает изменения и применяет их позже. Это нужно для
систем, потому что прямые мутации мира во время `MutationRunningSystem`
отклоняются.

```go
commands := &ecs.CommandBuffer{}

err := commands.Spawn().
	With(Position{}).
	With(Velocity{}).
	Commit()
if err != nil {
	panic(err)
}

if err := commands.Apply(world); err != nil {
	panic(err)
}
commands.Clear()
```

Доступные команды:

- `commands.Spawn().With(component).Commit()`;
- `commands.Add(entity).With(component).Commit()`;
- `commands.Remove(entity).Component(ecs.Component[T]()).Commit()`;
- `commands.Despawn(entity)`;
- `commands.Apply(world)`;
- `commands.Clear()`.

### Queries

Запрос создаётся через `NewQuery(world, name)`. Builder содержит методы:

- `With(componentToken)` - сущность должна иметь компонент;
- `Without(componentToken)` - сущность не должна иметь компонент;
- `Read(componentToken)` - компонент читается системой;
- `Write(componentToken)` - компонент изменяется системой;
- `Build()` - собрать `Query`.

Итератор запроса предоставляет `Next()` и `Entity()`, а доступ к компонентам
предполагается через `Read[T](it)` и `Write[T](it)`.

Целевой стиль использования:

```go
query, err := ecs.NewQuery(world, "movement").
	With(ecs.Component[Position]()).
	With(ecs.Component[Velocity]()).
	Read(ecs.Component[Velocity]()).
	Write(ecs.Component[Position]()).
	Build()
if err != nil {
	panic(err)
}

for it := query.Iter(); it.Next(); {
	entity := it.Entity()
	_ = entity

	velocity, err := ecs.Read[Velocity](it)
	if err != nil {
		panic(err)
	}

	position, err := ecs.Write[Position](it)
	if err != nil {
		panic(err)
	}

	position.X += velocity.X
	position.Y += velocity.Y
}
```

### Runner и System

Система должна реализовать интерфейс:

```go
type System interface {
	Name() string
	Stage() StageID
	Update(ctx *Context) error
	Access() AccessSet
	DebugQueries() []QueryDebugInfo
}
```

`Runner` выполняет системы по стадиям и применяет команды после каждой стадии:

```go
const UpdateStage ecs.StageID = 1

runner := ecs.NewRunner([]ecs.StageID{UpdateStage})
runner.Add(mySystem)

ctx := &ecs.Context{
	World:     world,
	Commands:  &ecs.CommandBuffer{},
	Resources: ecs.NewResources(),
}

if err := runner.ValidateAccess(); err != nil {
	panic(err)
}

if err := runner.Update(ctx); err != nil {
	panic(err)
}
```

### Resources

`Resources` предназначен для глобального состояния, не привязанного к
конкретной сущности. Контракт хранения ресурсов в текущей реализации ещё
нужно стабилизировать, поэтому этот API стоит считать черновым.

Доступные операции:

- `NewResources()` - создать контейнер;
- `PutResources[T](resources, value)` - сохранить ресурс;
- `GetResources[T](resources)` - получить ресурс;
- `RemoveResources[T](resources)` - удалить ресурс.

## Отладка

Отладочные методы:

```go
world.DebugArchetypes()
runner.DebugAccess()
runner.DebugQueries()
```

Они возвращают текстовые отчеты по архетипам, доступам систем и запросам.

## Разработка

Полезные команды:

```bash
gofmt -w .
go test ./...
go run .
```

В текущем окружении обновления README команда `go test ./...` не запускалась:
бинарник `go` недоступен в `PATH`.

## Лицензия

Проект распространяется под лицензией MIT. Подробности в [`LICENSE`](./LICENSE).
