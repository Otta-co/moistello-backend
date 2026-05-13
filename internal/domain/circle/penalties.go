package circle

import "math"

func CalculateLateFee(contributionAmount float64, lateFeePercent float64) float64 {
	fee := contributionAmount * lateFeePercent / 100.0
	precision := math.Pow(10, 7)
	return math.Round(fee*precision) / precision
}

func CalculateEarlyExitPenalty(totalContributed float64, collateralAmount float64, strikes int) float64 {
	if totalContributed <= 0 {
		return collateralAmount
	}
	basePenalty := collateralAmount * float64(strikes+1) / 10.0
	if basePenalty > totalContributed {
		basePenalty = totalContributed
	}
	precision := math.Pow(10, 7)
	return math.Round(basePenalty*precision) / precision
}

func ShouldRemoveMember(strikes int, maxStrikes int) bool {
	return strikes >= maxStrikes
}

func ApplyStrikes(member *CircleMember, penaltyType string) int {
	add := 1
	if penaltyType == "severe" {
		add = 3
	}
	return add
}
