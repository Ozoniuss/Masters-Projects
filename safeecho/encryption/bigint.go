package encryption

import "strconv"

const base = 10

type BigInt struct {
	digits []int64
}

func zero() *BigInt {
	return &BigInt{
		digits: []int64{},
	}
}

func fromString(n string) *BigInt {
	res := &BigInt{
		digits: make([]int64, len(n)),
	}
	for i, c := range n {
		res.digits[len(n)-i-1] = int64(c - '0')
	}
	res.normalize()
	return res
}

func fromInt(n int64) *BigInt {
	return fromString(strconv.FormatInt(n, 10))
}

func (a *BigInt) toInt() int64 {
	x := int64(0)
	for i := len(a.digits) - 1; i >= 0; i-- {
		x *= base
		x += a.digits[i]
	}
	return x
}

func (a *BigInt) even() bool {
	return a.digits[0]%2 == 0
}

func (a *BigInt) normalize() {
	for len(a.digits) > 0 && a.digits[len(a.digits)-1] == 0 {
		a.digits = a.digits[:len(a.digits)-1]
	}
}

func (a *BigInt) copy() *BigInt {
	res := &BigInt{
		digits: make([]int64, len(a.digits)),
	}
	copy(res.digits, a.digits)
	return res
}

func (a *BigInt) prev() (result *BigInt) {
	result = a.copy()

	carry := int64(1)
	for i := 0; i < len(result.digits); i++ {
		result.digits[i] = (a.digits[i] - carry + base) % base
		if a.digits[i]-carry < 0 {
			carry = 1
		} else {
			carry = 0
		}
	}
	if carry == 1 {
		panic("underflow")
	}

	result.normalize()

	return
}

func (a *BigInt) next() (result *BigInt) {
	result = a.copy()

	carry := int64(1)
	for i := 0; i < len(result.digits); i++ {
		result.digits[i] = (a.digits[i] + carry) % base
		carry = (a.digits[i] + carry) / base
	}
	if carry == 1 {
		result.digits = append(result.digits, 1)
	}

	result.normalize()

	return
}

func (a *BigInt) half() *BigInt {
	result := &BigInt{
		digits: make([]int64, len(a.digits)),
	}

	carry := int64(0)
	for i := len(a.digits) - 1; i >= 0; i-- {
		result.digits[i] = (a.digits[i] + carry*base) / 2
		carry = (a.digits[i] + carry*base) % 2
	}

	result.normalize()

	return result
}

func (a *BigInt) String() string {
	result := ""
	for i := len(a.digits) - 1; i >= 0; i-- {
		result += strconv.Itoa(int(a.digits[i]))
	}
	if result == "" {
		result = "0"
	}
	return result
}

func (a *BigInt) compare(b *BigInt) int {
	if len(a.digits) > len(b.digits) {
		return 1
	}
	if len(a.digits) < len(b.digits) {
		return -1
	}
	for i := len(a.digits) - 1; i >= 0; i-- {
		if a.digits[i] > b.digits[i] {
			return 1
		}
		if a.digits[i] < b.digits[i] {
			return -1
		}
	}
	return 0
}

func (a *BigInt) add(b *BigInt) (result *BigInt) {
	result = &BigInt{
		digits: make([]int64, 0),
	}

	carry := int64(0)
	for i := 0; i < len(a.digits) && i < len(b.digits); i++ {
		result.digits = append(result.digits, (a.digits[i]+b.digits[i]+carry)%base)
		carry = (a.digits[i] + b.digits[i] + carry) / base
	}
	for i := len(a.digits); i < len(b.digits); i++ {
		result.digits = append(result.digits, (b.digits[i]+carry)%base)
		carry = (b.digits[i] + carry) / base
	}
	for i := len(b.digits); i < len(a.digits); i++ {
		result.digits = append(result.digits, (a.digits[i]+carry)%base)
		carry = (a.digits[i] + carry) / base
	}
	if carry != 0 {
		result.digits = append(result.digits, carry)
	}

	return
}

func (a *BigInt) sub(b *BigInt) (result *BigInt) {
	result = &BigInt{
		digits: make([]int64, 0),
	}

	carry := int64(0)
	for i := 0; i < len(b.digits); i++ {
		result.digits = append(result.digits, (a.digits[i]-b.digits[i]-carry+base)%base)
		if a.digits[i]-b.digits[i]-carry < 0 {
			carry = 1
		} else {
			carry = 0
		}
	}
	for i := len(b.digits); i < len(a.digits); i++ {
		result.digits = append(result.digits, (a.digits[i]-carry+base)%base)
		if a.digits[i]-carry < 0 {
			carry = 1
		} else {
			carry = 0
		}
	}
	if carry == 1 {
		panic("underflow")
	}

	result.normalize()

	return
}

func (a *BigInt) mul(b *BigInt) (result *BigInt) {
	result = &BigInt{
		digits: make([]int64, len(a.digits)+len(b.digits)-1),
	}

	for i := 0; i < len(a.digits); i++ {
		for j := 0; j < len(b.digits); j++ {
			result.digits[i+j] += a.digits[i] * b.digits[j]
		}
	}

	carry := int64(0)
	for i := 0; i < len(result.digits); i++ {
		result.digits[i] += carry
		carry = result.digits[i] / base
		result.digits[i] %= base
	}

	if carry != 0 {
		result.digits = append(result.digits, carry)
	}

	return
}

// divide a by b using binary search, return quotient and remainder
func (a *BigInt) div(b *BigInt) (q, r *BigInt) {
	if a.compare(b) < 0 {
		return zero(), a
	}

	// binary search
	l, r := zero(), a
	ans := l.copy()
	for l.compare(r) <= 0 {
		temp := l.add(r)
		m := temp.half()
		if m.mul(b).compare(a) <= 0 {
			ans = m.copy()
			l = m.next()
		} else {
			r = m.prev()
		}
	}

	return ans, a.sub(ans.mul(b))
}
