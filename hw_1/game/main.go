package main

import "strings"

type Item struct {
	Name string
}

type Room struct {
	Name        string
	Description func() string
	Items       map[string]*Item
	Exits       []string
}

type Player struct {
	CurrentRoom string
	HasBag      bool
	Inventory   map[string]*Item
}

var (
	rooms      map[string]*Room
	player     *Player
	doorOpened bool
)

func main() {

	/*
		в этой функции можно ничего не писать,
		но тогда у вас не будет работать через go run main.go
		очень круто будет сделать построчный ввод команд тут, хотя это и не требуется по заданию
	*/

	initGame()

	_ = handleCommand("some command")
}

func initGame() {
	rooms = make(map[string]*Room)
	player = &Player{
		CurrentRoom: "кухня",
		HasBag:      false,
		Inventory:   make(map[string]*Item),
	}
	doorOpened = false
	rooms["кухня"] = &Room{
		Name:  "кухня",
		Items: map[string]*Item{"чай": {Name: "чай"}},
		Exits: []string{"коридор"},
		Description: func() string {
			if player.HasBag {
				return "ты находишься на кухне, на столе: чай, надо идти в универ. можно пройти - коридор"
			}
			return "ты находишься на кухне, на столе: чай, надо собрать рюкзак и идти в универ. можно пройти - коридор"
		},
	}
	rooms["коридор"] = &Room{
		Name:  "коридор",
		Items: make(map[string]*Item),
		Exits: []string{"кухня", "комната", "улица"},
		Description: func() string {
			if player.CurrentRoom == "кухня" && len(rooms["кухня"].Items) == 1 && player.HasBag {
				return "кухня, ничего интересного. можно пройти - коридор"
			}
			return "ничего интересного. можно пройти - кухня, комната, улица"
		},
	}

	rooms["комната"] = &Room{
		Name: "комната",
		Items: map[string]*Item{
			"ключи":     {Name: "ключи"},
			"конспекты": {Name: "конспекты"},
			"рюкзак":    {Name: "рюкзак"},
		},
		Exits: []string{"коридор"},
		Description: func() string {
			r := rooms["комната"]
			_, hasKeys := r.Items["ключи"]
			_, hasNotes := r.Items["конспекты"]
			_, hasBag := r.Items["рюкзак"]

			if !hasKeys && !hasNotes && !hasBag {
				return "пустая комната. можно пройти - коридор"
			}

			var onTable []string
			if hasKeys {
				onTable = append(onTable, "ключи")
			}
			if hasNotes {
				onTable = append(onTable, "конспекты")
			}

			var result []string
			if len(onTable) > 0 {
				result = append(result, "на столе: "+strings.Join(onTable, ", "))
			}
			if hasBag {
				result = append(result, "на стуле: рюкзак")
			}

			return strings.Join(result, ", ") + ". можно пройти - коридор"
		},
	}
	rooms["улица"] = &Room{
		Name:  "улица",
		Items: make(map[string]*Item),
		Exits: []string{"домой"},
		Description: func() string {
			return "на улице весна. можно пройти - домой"
		},
	}
}

func handleCommand(command string) string {
	args := strings.Split(command, " ")
	if len(args) == 0 || args[0] == "" {
		return "неизвестная команда"
	}
	verb := args[0]
	currentRoom := rooms[player.CurrentRoom]

	switch verb {
	case "осмотреться":
		return currentRoom.Description()

	case "идти":
		if len(args) < 2 {
			return "куда идти?"
		}
		target := args[1]

		canGo := false
		for _, exit := range currentRoom.Exits {
			if exit == target {
				canGo = true
				break
			}
		}
		if !canGo {
			return "нет пути в " + target
		}

		if player.CurrentRoom == "коридор" && target == "улица" && !doorOpened {
			return "дверь закрыта"
		}

		player.CurrentRoom = target
		if target == "комната" {
			return "ты в своей комнате. можно пройти - коридор"
		}
		if target == "кухня" && player.HasBag {
			return "кухня, ничего интересного. можно пройти - коридор"
		}

		return rooms[player.CurrentRoom].Description()

	case "надеть":
		if len(args) < 2 {
			return "что надеть?"
		}
		target := args[1]
		if target == "рюкзак" {
			if _, exists := currentRoom.Items["рюкзак"]; exists {
				player.HasBag = true
				delete(currentRoom.Items, "рюкзак")
				return "вы надели: рюкзак"
			}
		}
		return "нет такого"

	case "взять":
		if len(args) < 2 {
			return "что взять?"
		}
		target := args[1]
		item, exists := currentRoom.Items[target]
		if !exists {
			return "нет такого"
		}
		if !player.HasBag {
			return "некуда класть"
		}

		delete(currentRoom.Items, target)
		player.Inventory[target] = item
		return "предмет добавлен в инвентарь: " + target

	case "применить":
		if len(args) < 3 {
			return "неизвестная команда"
		}
		itemNeeded := args[1]
		objectTarget := args[2]
		if _, hasItem := player.Inventory[itemNeeded]; !hasItem {
			return "нет предмета в инвентаре - " + itemNeeded
		}
		if objectTarget == "дверь" {
			if itemNeeded == "ключи" && player.CurrentRoom == "коридор" {
				doorOpened = true
				return "дверь открыта"
			}
			return "не к чему применить"
		}

		if objectTarget == "шкаф" {
			return "не к чему применить"
		}

		return "не к чему применить"

	default:
		return "неизвестная команда"
	}
}
