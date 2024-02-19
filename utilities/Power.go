package utilities

func Power(x *int, y *int) int {
	number := *x

	for i := 0; i < *y; i++ {
		number *= number
	}

	return number
}
