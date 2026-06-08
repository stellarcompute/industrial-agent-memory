package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/fde-ai/industrial-agent-memory/memory"
)

func main() {
	ctx := context.Background()
	service := memory.NewService(memory.NewInMemoryStore())

	item, err := service.Remember(ctx, memory.Observation{
		Kind:    memory.KindKnownFailure,
		Scope:   memory.Scope{Type: memory.ScopeEquipment, ID: "press-01", Name: "冲压机 01"},
		Title:   "振动升高伴随温度上升",
		Content: "当冲压机振动和温度同时升高时，优先检查轴承润滑、模具偏载和最近一次换模记录。",
		Tags:    []string{"vibration", "temperature", "bearing", "press"},
		Signals: map[string]string{
			"alarm_code": "vib-high",
			"line":       "press-line-1",
		},
		Confidence: memory.ConfidenceHigh,
		Importance: 0.82,
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = service.MarkOutcome(ctx, item.ID, memory.OutcomeEffective)
	if err != nil {
		log.Fatal(err)
	}

	results, err := service.Recall(ctx, memory.RecallQuery{
		Scope: memory.Scope{Type: memory.ScopeEquipment, ID: "press-01"},
		Text:  "vibration temperature bearing",
		Tags:  []string{"vibration"},
		Signals: map[string]string{
			"alarm_code": "vib-high",
		},
		Limit: 3,
	})
	if err != nil {
		log.Fatal(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	fmt.Println("recall results:")
	if err := encoder.Encode(results); err != nil {
		log.Fatal(err)
	}
}
