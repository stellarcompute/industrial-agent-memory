package memory

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"
)

type Clock func() time.Time

type Service struct {
	store Store
	now   Clock
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
		now:   time.Now,
	}
}

func (s *Service) WithClock(clock Clock) *Service {
	s.now = clock
	return s
}

func (s *Service) Remember(ctx context.Context, observation Observation) (Memory, error) {
	if err := validateObservation(observation); err != nil {
		return Memory{}, err
	}

	now := s.now().UTC()
	item := Memory{
		ID:         newID("mem"),
		Kind:       observation.Kind,
		Scope:      observation.Scope,
		Title:      strings.TrimSpace(observation.Title),
		Content:    strings.TrimSpace(observation.Content),
		Tags:       normalizeList(observation.Tags),
		Signals:    normalizeMap(observation.Signals),
		Evidence:   observation.Evidence,
		Confidence: defaultConfidence(observation.Confidence),
		Outcome:    OutcomeUnknown,
		Importance: clamp(observation.Importance, 0, 1),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if observation.TTL > 0 {
		expiresAt := now.Add(observation.TTL)
		item.ExpiresAt = &expiresAt
	}

	if item.Importance == 0 {
		item.Importance = 0.5
	}

	if err := s.store.Save(ctx, item); err != nil {
		return Memory{}, err
	}

	return item, nil
}

func (s *Service) Recall(ctx context.Context, query RecallQuery) ([]RecallResult, error) {
	now := query.Now
	if now.IsZero() {
		now = s.now().UTC()
	}
	if query.Limit <= 0 {
		query.Limit = 8
	}

	items, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]RecallResult, 0, len(items))
	for _, item := range items {
		if isExpired(item, now) {
			continue
		}
		score, reason := scoreMemory(item, query, now)
		if score <= 0 {
			continue
		}
		if !query.IncludeWeak && score < 0.22 {
			continue
		}

		results = append(results, RecallResult{
			Memory: item,
			Score:  score,
			Reason: reason,
		})
	}

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Memory.UpdatedAt.After(results[j].Memory.UpdatedAt)
		}
		return results[i].Score > results[j].Score
	})

	if len(results) > query.Limit {
		results = results[:query.Limit]
	}

	return results, nil
}

func (s *Service) MarkOutcome(ctx context.Context, id string, outcome Outcome) (Memory, error) {
	if outcome == "" {
		return Memory{}, errors.New("outcome is required")
	}

	item, err := s.store.Get(ctx, id)
	if err != nil {
		return Memory{}, err
	}

	item.Outcome = outcome
	item.UpdatedAt = s.now().UTC()
	if outcome == OutcomeEffective {
		item.Confidence = ConfidenceHigh
		item.Importance = clamp(item.Importance+0.15, 0, 1)
	}
	if outcome == OutcomeRejected {
		item.Importance = clamp(item.Importance-0.25, 0, 1)
	}

	if err := s.store.Save(ctx, item); err != nil {
		return Memory{}, err
	}

	return item, nil
}

func (s *Service) ForgetExpired(ctx context.Context) (int, error) {
	now := s.now().UTC()
	items, err := s.store.List(ctx)
	if err != nil {
		return 0, err
	}

	deleted := 0
	for _, item := range items {
		if !isExpired(item, now) {
			continue
		}
		if err := s.store.Delete(ctx, item.ID); err != nil {
			return deleted, err
		}
		deleted++
	}

	return deleted, nil
}

func (s *Service) ConsolidationCandidates(ctx context.Context, threshold float64) ([]ConsolidationCandidate, error) {
	if threshold <= 0 {
		threshold = 0.68
	}

	items, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}

	candidates := make([]ConsolidationCandidate, 0)
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			score, reasons := consolidationScore(items[i], items[j])
			if score < threshold {
				continue
			}

			candidates = append(candidates, ConsolidationCandidate{
				PrimaryID:   items[i].ID,
				DuplicateID: items[j].ID,
				Reasons:     reasons,
				Score:       score,
			})
		}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	return candidates, nil
}

func validateObservation(observation Observation) error {
	if observation.Scope.Type == "" || strings.TrimSpace(observation.Scope.ID) == "" {
		return errors.New("scope type and id are required")
	}
	if strings.TrimSpace(observation.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(observation.Content) == "" {
		return errors.New("content is required")
	}
	if observation.Kind == "" {
		return errors.New("kind is required")
	}

	return nil
}

func defaultConfidence(value Confidence) Confidence {
	if value == "" {
		return ConfidenceMedium
	}
	return value
}

func newID(prefix string) string {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return prefix + "-" + hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}

	return prefix + "-" + hex.EncodeToString(bytes[:])
}
