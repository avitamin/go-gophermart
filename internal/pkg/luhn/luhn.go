// Package luhn provides Luhn algorithm validation for order numbers.
package luhn

// Valid checks if a number string is valid according to the Luhn algorithm.
func Valid(number string) bool {
	if len(number) == 0 {
		return false
	}

	var sum int
	isSecond := false

	for i := len(number) - 1; i >= 0; i-- {
		d := int(number[i] - '0')
		if d < 0 || d > 9 {
			return false
		}

		if isSecond {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}

		sum += d
		isSecond = !isSecond
	}

	return sum%10 == 0
}
