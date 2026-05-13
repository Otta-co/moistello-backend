package circle

import (
	"math"
	"math/rand"
)

func RandomOrder(seed int64, memberCount int) []int {
	order := make([]int, memberCount)
	for i := range order {
		order[i] = i
	}
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(order), func(i, j int) {
		order[i], order[j] = order[j], order[i]
	})
	return order
}

func AuctionWinner(bids map[string]float64) (string, float64) {
	var maxBid float64
	var winner string
	for user, bid := range bids {
		if bid > maxBid {
			maxBid = bid
			winner = user
		}
	}
	return winner, maxBid
}

func VoteTally(votes map[string]int) string {
	var maxVotes int
	var winner string
	for user, cnt := range votes {
		if cnt > maxVotes {
			maxVotes = cnt
			winner = user
		}
	}
	return winner
}

func CalculatePayout(contributionAmount float64, memberCount int, roundNumber int, latePenalties float64) float64 {
	totalPool := contributionAmount*float64(memberCount) + latePenalties
	precision := math.Pow(10, 7)
	return math.Round(totalPool*precision) / precision
}
