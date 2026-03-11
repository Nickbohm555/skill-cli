package overlap

const (
	descriptionWeight    = 0.35
	inScopeJaccardWeight = 0.30
	outOfScopeWeight     = 0.20
	commandOverlapWeight = 0.15

	mediumOverlapThreshold = 0.40
	highOverlapThreshold   = 0.70
)

func ClassifyScore(score float64) OverlapSeverity {
	switch {
	case score >= highOverlapThreshold:
		return SeverityHigh
	case score >= mediumOverlapThreshold:
		return SeverityMedium
	case score > 0:
		return SeverityLow
	default:
		return SeverityNone
	}
}
