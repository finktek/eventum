package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/finktek/eventum"
	"github.com/finktek/eventum/eventstore"
	"github.com/finktek/eventum/examples/todo/domain"
	subscriptions2 "github.com/finktek/eventum/mongodb"
	"github.com/finktek/eventum/subscriptions"
	"github.com/gofrs/uuid"
	"log"
	"time"
)

func main()  {
	eventStore, _ := eventstore.NewEventStoreDbClient("esdb://localhost:2113?tls=false")
	todoAggregateStore, _:= finkgoes.NewAggregateStore(eventStore, &domain.TodoList{})
	commandHandler := &domain.TodoCommandHandler{AggregateStore: todoAggregateStore}

	cmdId := uuid.Must(uuid.NewV4())
	cmd := finkgoes.NewCommand(&domain.CreateTodoList{Id: cmdId, Title: "My Todo"})

	err := commandHandler.Handle(context.Background(), cmd)
	if err != nil {
		fmt.Println(err)
	}
	cmd2ItemId := uuid.Must(uuid.NewV4())
	cmd2 := finkgoes.NewCommand(&domain.AddTodoItem{Id: cmdId, TodoItemID: cmd2ItemId, Description: "Do something good" })
	err = commandHandler.Handle(context.Background(), cmd2)
	if err != nil {
		fmt.Println(err)
	}

	cmd3ItemId := uuid.Must(uuid.NewV4())
	cmd3 := finkgoes.NewCommand(&domain.AddTodoItem{Id: cmdId, TodoItemID: cmd3ItemId, Description: "Do nothing for the rest of the day" })
	err = commandHandler.Handle(context.Background(), cmd3)
	if err != nil {
		fmt.Println(err)
	}

	cmd33ItemId := uuid.Must(uuid.NewV4())
	cmd33 := finkgoes.NewCommand(&domain.AddTodoItem{Id: cmdId, TodoItemID: cmd33ItemId, Description: "Sleep" })
	err = commandHandler.Handle(context.Background(), cmd33)
	if err != nil {
		fmt.Println(err)
	}


	cmd4 := finkgoes.NewCommand(&domain.MarkItemAsDone{Id: cmdId, TodoItemID: cmd3ItemId})
	err = commandHandler.Handle(context.Background(), cmd4)
	if err != nil {
		fmt.Println(err)
	}

	cmd5 := finkgoes.NewCommand(&domain.MarkItemAsDone{Id: cmdId, TodoItemID: cmd2ItemId})
	err = commandHandler.Handle(context.Background(), cmd5)
	if err != nil {
		fmt.Println(err)
	}

	todo := domain.InitTodoList(cmdId.String())
	if err := todoAggregateStore.Load(context.Background(), todo, cmdId.String()); err != nil {
		log.Println("error loading todo list: ", err)
	}

	todoData, _ := json.Marshal(&todo)

	log.Println("TODO ", string(todoData))

	var subscription subscriptions.SubscriptionService
	var checkpointStore subscriptions.CheckpointStore
	checkpointStore, _ = subscriptions2.NewMongoDbCheckpointStore()
	subscription, _ = eventstore.NewAllStreamSubscription("esdb://localhost:2113?tls=false", "list-suub", checkpointStore)
	subscription.AddHandler(ListEventHandler{})
	time.Sleep(time.Second * 1)
	subscription.Start(context.Background())
	time.Sleep(time.Second * 100)

}

type ListEventHandler struct {
}

func (h ListEventHandler) Handle(event finkgoes.Event) {
	switch e := event.GetData().(type) {
	case *domain.TodoListCreated:
		log.Println("CREATED", e.ID)
	case *domain.TodoItemAdded:
		log.Println("ADDED", e.ID)
	case *domain.TodoItemDone:
		log.Println("DONE ", e.ID)
	}
}
