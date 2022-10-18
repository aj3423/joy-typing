package mode

import (
	"strconv"
	"strings"
)

var mapNum = map[string]int{
	"zero":      0,
	"one":       1,
	"two":       2,
	"three":     3,
	"four":      4,
	"five":      5,
	"six":       6,
	"seven":     7,
	"eight":     8,
	"nine":      9,
	"ten":       10,
	"eleven":    11,
	"twelve":    12,
	"thirteen":  13,
	"fourteen":  14,
	"fifteen":   15,
	"sixteen":   16,
	"seventeen": 17,
	"eighteen":  18,
	"nineteen":  19,
	"twenty":    20,
	"thirty":    30,
	"forty":     40,
	"fifty":     50,
	"sixty":     60,
	"seventy":   70,
	"eighty":    80,
	"ninety":    90,
}
var mapMultiplier = map[string]int{
	"hundred":  100,
	"thousand": 1000,
}

// `numberUntil` scans from beginning, find all words that
// represents 1 number and stop
// Return Value:
//  0. the result number(if succeeded)
//  1. the cost, how many words were used to get that number
//  2. success, `true` if a number is found from the beginning
//
// ["haha", "seven", "two"]
// Returns: nil, 1, false ---- because the first string "haha" is not number
//
// ["seven", "two", "thousand", "n", "six", "hundred", "n", "five"]
// Returns: 7, 1, true  ---- because "seven" is number and "seven two" is not, cost 1 string, succeeded
//
// ["two", "thousand", "n", "six", "hundred", "n", "five", "haha"]
// Returns: 2605, 7, true ---- "n" is considered as "and", added up, costs 7 strings, and succeeded
func numberUntil(words []string) (int, int, bool) {
	skip := 0 // always increases by 1 when each time `pop1()`
	pop1 := func() {
		words = words[1:]
		skip += 1
	}
	first := func() string {
		// cast to lower case, so it can handle "TWO THOUSAND"
		return strings.ToLower(words[0])
	}
	empty := func() bool { return len(words) == 0 }

	if empty() {
		return 0, 0, false
	}

	if v0, ok := mapNum[first()]; ok { // is number
		pop1()

		if empty() {
			return v0, skip, true
		}
		if mul, ok := mapMultiplier[first()]; ok { // is multiplier, e.g. "two thousand"
			pop1()

			v0 *= mul
			if empty() { // no more words left
				return v0, skip, true
			} else { // more words
				first_ := first()

				if mul2, ok := mapMultiplier[first()]; ok { // is multiplier again, e.g. "two thousand hundred"
					v0 *= mul2
					pop1()

					if empty() { // no more words left
						return v0, skip, true
					}
					first_ = first()
				}
				if first_ == "n" || first_ == "and" {
					pop1()
					if empty() {
						return v0, skip, true
					}
				}
				v1, cost, ok := numberUntil(words[0:])
				// only add them up if the latter number less than the multiplier
				// so only "two hundred n twenty" -> 220
				// but NOT "two hundred n one hundred" -> 300
				if ok && v1 < mul {
					return v0 + v1, cost + skip, true
				} else {
					return v0, skip, true
				}
			}

		} else {
			switch v0 {
			case 20, 30, 40, 50, 60, 70, 80, 90: // e.g.: twenty one
				v1, cost, ok := numberUntil(words[0:1])
				if ok && v1 < 10 {
					return v0 + v1, 1 + cost, true
				} else {
					return v0, 1, true
				}
			default:
				return v0, 1, true
			}
		}
	}
	return 0, 1, false
}

func replaceNumbers(words []string) []string {

	ret := []string{}
	for len(words) > 0 {
		value, cost, ok := numberUntil(words)
		if !ok {
			// if it's not number, simply append this element
			ret = append(ret, words[0])
			words = words[cost:]
		} else {
			// if it is number, it `cost` some elements in arr,
			// skip those costs and append the number to result
			ret = append(ret, strconv.Itoa(value))
			words = words[cost:]
		}

	}
	return ret
}
