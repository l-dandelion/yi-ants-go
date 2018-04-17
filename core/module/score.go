package module

/*
 * function for calculating the score of a module
 */
type CalculateScore func(counts Counts) uint64

/*
 * simple function for calculating the score of a module
 */
func CalculateScoreSimple(counts Counts) uint64 {
	return counts.CalledCount +
		counts.AcceptedCount<<1 +
		counts.CompletedCount<<2 +
		counts.HandlingNumber<<4
}

/*
 * set score of a module
 * return true if updated
 */
func SetScore(module Module) bool {
	calculator := module.ScoreCalculator()
	if calculator == nil {
		calculator = CalculateScoreSimple
	}
	newScore := calculator(module.Counts())
	if newScore == module.Score() {
		return false
	}
	module.SetScore(newScore)
	return true
}
