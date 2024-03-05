package utilities

// Returns the minimum of two integers. Needed this for
// video streaming chunk calculations.
func Minimum(num1 int, num2 int) int {
	if num1 < num2 {
		return num1
	}

	return num2
}
