package memory

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"
)

func scoreMemory(item Memory, query RecallQuery, now time.Time) (float64, string) {
	reasons := make([]string, 0, 5)
	score := 0.0

	if item.Scope.Type == query.Scope.Type && item.Scope.ID == query.Scope.ID {
		score += 0.42
		reasons = append(reasons, "same scope")
	} else if item.Scope.Type == query.Scope.Type {
		score += 0.12
		reasons = append(reasons, "same scope type")
	}

	if len(query.Kinds) > 0 && slices.Contains(query.Kinds, item.Kind) {
		score += 0.12
		reasons = append(reasons, "kind matched")
	}

	tagMatches := overlap(item.Tags, normalizeList(query.Tags))
	if tagMatches > 0 {
		score += math.Min(0.18, float64(tagMatches)*0.06)
		reasons = append(reasons, fmt.Sprintf("%d tag match", tagMatches))
	}

	signalMatches := signalOverlap(item.Signals, normalizeMap(query.Signals))
	if signalMatches > 0 {
		score += math.Min(0.16, float64(signalMatches)*0.08)
		reasons = append(reasons, fmt.Sprintf("%d signal match", signalMatches))
	}

	textScore := textOverlap(item.Title+" "+item.Content, query.Text)
	if textScore > 0 {
		score += math.Min(0.18, textScore)
		reasons = append(reasons, "text matched")
	}

	score += item.Importance * 0.12
	score += confidenceWeight(item.Confidence)
	score += outcomeWeight(item.Outcome)
	score *= recencyWeight(item.UpdatedAt, now)

	if len(reasons) == 0 {
		return 0, ""
	}

	return clamp(score, 0, 1), strings.Join(reasons, ", ")
}

func consolidationScore(a Memory, b Memory) (float64, []string) {
	reasons := make([]string, 0, 4)
	score := 0.0

	if a.Scope.Type == b.Scope.Type && a.Scope.ID == b.Scope.ID {
		score += 0.35
		reasons = append(reasons, "same scope")
	}
	if a.Kind == b.Kind {
		score += 0.18
		reasons = append(reasons, "same kind")
	}

	tagMatches := overlap(a.Tags, b.Tags)
	if tagMatches > 0 {
		score += math.Min(0.22, float64(tagMatches)*0.07)
		reasons = append(reasons, "shared tags")
	}

	textScore := textOverlap(a.Title+" "+a.Content, b.Title+" "+b.Content)
	if textScore > 0.08 {
		score += math.Min(0.25, textScore)
		reasons = append(reasons, "similar wording")
	}

	return clamp(score, 0, 1), reasons
}

func confidenceWeight(confidence Confidence) float64 {
	switch confidence {
	case ConfidenceHigh:
		return 0.08
	case ConfidenceMedium:
		return 0.04
	default:
		return 0
	}
}

func outcomeWeight(outcome Outcome) float64 {
	switch outcome {
	case OutcomeEffective:
		return 0.08
	case OutcomeRejected, OutcomeExpired:
		return -0.1
	default:
		return 0
	}
}

func recencyWeight(updatedAt time.Time, now time.Time) float64 {
	age := now.Sub(updatedAt)
	if age <= 0 {
		return 1
	}
	if age < 24*time.Hour {
		return 1
	}
	if age < 30*24*time.Hour {
		return 0.92
	}
	if age < 180*24*time.Hour {
		return 0.78
	}
	return 0.62
}

func isExpired(item Memory, now time.Time) bool {
	return item.ExpiresAt != nil && !item.ExpiresAt.After(now)
}

func overlap(a []string, b []string) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	seen := make(map[string]struct{}, len(a))
	for _, value := range a {
		seen[strings.ToLower(strings.TrimSpace(value))] = struct{}{}
	}

	matches := 0
	for _, value := range b {
		if _, ok := seen[strings.ToLower(strings.TrimSpace(value))]; ok {
			matches++
		}
	}

	return matches
}

func signalOverlap(a map[string]string, b map[string]string) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	matches := 0
	for key, value := range b {
		if a[key] == value {
			matches++
		}
	}

	return matches
}

func textOverlap(a string, b string) float64 {
	left := tokenize(a)
	right := tokenize(b)
	if len(left) == 0 || len(right) == 0 {
		return 0
	}

	matches := overlap(left, right)
	return float64(matches) / float64(max(len(left), len(right)))
}

func tokenize(text string) []string {
	fields := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return r == ' ' || r == ',' || r == '.' || r == ';' || r == ':' || r == '/' || r == '\\' || r == '-' || r == '_' || r == '\n' || r == '\t'
	})

	return normalizeList(fields)
}

func normalizeList(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}

		seen[value] = struct{}{}
		result = append(result, value)
	}

	return result
}

func normalizeMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}

	result := make(map[string]string, len(values))
	for key, value := range values {
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.ToLower(strings.TrimSpace(value))
		if key == "" || value == "" {
			continue
		}
		result[key] = value
	}

	return result
}

func clamp(value float64, minValue float64, maxValue float64) float64 {
	return math.Max(minValue, math.Min(maxValue, value))
}
