package memory

import (
	"context"
	"testing"
	"time"
)

func TestRecallPrioritizesSameEquipmentMemory(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	service := NewService(NewInMemoryStore()).WithClock(func() time.Time { return now })

	_, err := service.Remember(ctx, Observation{
		Kind:    KindKnownFailure,
		Scope:   Scope{Type: ScopeEquipment, ID: "press-01"},
		Title:   "冲压机振动升高",
		Content: "振动超过 7mm/s 且温度升高时，通常先检查轴承润滑和模具偏载。",
		Tags:    []string{"vibration", "bearing", "press"},
		Signals: map[string]string{
			"alarm_code": "vib-high",
		},
		Confidence: ConfidenceHigh,
		Importance: 0.8,
	})
	if err != nil {
		t.Fatalf("remember: %v", err)
	}

	_, err = service.Remember(ctx, Observation{
		Kind:       KindProcedure,
		Scope:      Scope{Type: ScopeEquipment, ID: "oven-02"},
		Title:      "固化炉温度波动",
		Content:    "固化炉温控波动时检查热电偶和风道。",
		Tags:       []string{"temperature", "oven"},
		Confidence: ConfidenceMedium,
		Importance: 0.6,
	})
	if err != nil {
		t.Fatalf("remember unrelated: %v", err)
	}

	results, err := service.Recall(ctx, RecallQuery{
		Scope: Scope{Type: ScopeEquipment, ID: "press-01"},
		Text:  "vibration temperature bearing",
		Tags:  []string{"vibration"},
		Signals: map[string]string{
			"alarm_code": "vib-high",
		},
		Limit: 3,
		Now:   now,
	})
	if err != nil {
		t.Fatalf("recall: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected recall result")
	}
	if results[0].Memory.Scope.ID != "press-01" {
		t.Fatalf("expected press memory first, got %s", results[0].Memory.Scope.ID)
	}
}

func TestMarkOutcomeStrengthensEffectiveMemory(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewInMemoryStore())

	item, err := service.Remember(ctx, Observation{
		Kind:       KindMaintenance,
		Scope:      Scope{Type: ScopeEquipment, ID: "pump-01"},
		Title:      "泵体异响",
		Content:    "更换联轴器后异响消失。",
		Importance: 0.4,
	})
	if err != nil {
		t.Fatalf("remember: %v", err)
	}

	updated, err := service.MarkOutcome(ctx, item.ID, OutcomeEffective)
	if err != nil {
		t.Fatalf("mark outcome: %v", err)
	}

	if updated.Confidence != ConfidenceHigh {
		t.Fatalf("expected high confidence, got %s", updated.Confidence)
	}
	if updated.Importance <= item.Importance {
		t.Fatalf("expected importance to increase")
	}
}

func TestForgetExpiredDeletesOldMemory(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	service := NewService(NewInMemoryStore()).WithClock(func() time.Time { return now })

	_, err := service.Remember(ctx, Observation{
		Kind:    KindObservation,
		Scope:   Scope{Type: ScopeLine, ID: "line-01"},
		Title:   "临时观察",
		Content: "这是一条只保留一小时的班次观察。",
		TTL:     time.Hour,
	})
	if err != nil {
		t.Fatalf("remember: %v", err)
	}

	service.WithClock(func() time.Time { return now.Add(2 * time.Hour) })
	deleted, err := service.ForgetExpired(ctx)
	if err != nil {
		t.Fatalf("forget expired: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 deleted memory, got %d", deleted)
	}
}

func TestConsolidationCandidatesFindSimilarMemories(t *testing.T) {
	ctx := context.Background()
	service := NewService(NewInMemoryStore())

	for _, title := range []string{"轴承润滑不足导致振动", "振动升高与轴承润滑不足有关"} {
		_, err := service.Remember(ctx, Observation{
			Kind:    KindKnownFailure,
			Scope:   Scope{Type: ScopeEquipment, ID: "press-01"},
			Title:   title,
			Content: "处理方式是补充润滑并复核轴承状态。",
			Tags:    []string{"bearing", "vibration"},
		})
		if err != nil {
			t.Fatalf("remember: %v", err)
		}
	}

	candidates, err := service.ConsolidationCandidates(ctx, 0.55)
	if err != nil {
		t.Fatalf("consolidation candidates: %v", err)
	}
	if len(candidates) == 0 {
		t.Fatal("expected consolidation candidate")
	}
}
