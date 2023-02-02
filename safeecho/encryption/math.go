package encryption

func phi(n *BigInt) *BigInt {
	result := n.copy()
	for i := fromInt(2); i.mul(i).compare(n) <= 0; i = i.next() {
		dv, mod := n.div(i)
		if mod.compare(fromInt(0)) == 0 {
			for mod.compare(fromInt(0)) == 0 {
				dv, mod = n.div(i)
				if mod.compare(fromInt(0)) == 0 {
					n = dv.copy()
				}
			}
			dv, _ = result.div(i)
			result = result.sub(dv)
		}
	}
	if n.compare(fromInt(1)) > 0 {
		dv, _ := result.div(n)
		result = result.sub(dv)
	}
	return result
}

func pow(x, y, m *BigInt) *BigInt {
	if y.compare(fromInt(0)) == 0 {
		return fromInt(1)
	}

	//fmt.Printf("%v %d\n", y.digits, y.digits[0])

	p := pow(x, y.half(), m)

	_, tempmod := p.mul(p).div(m)

	p = tempmod.copy()
	if y.even() {
		return p
	} else {
		_, tempmod := p.mul(x).div(m)
		return tempmod
	}
}

func modularInverse(a, m *BigInt) *BigInt {
	//println(m.String())
	//println(phi(m).prev().String())
	return pow(a.copy(), phi(m).prev(), m.copy())
}

func isPrime(x *BigInt) bool {
	if x.compare(fromInt(2)) < 0 {
		return false
	}
	if x.compare(fromInt(2)) == 0 {
		return true
	}
	if x.even() {
		return false
	}
	for i := fromInt(3); i.mul(i).compare(x) <= 0; i = i.next().next() {
		_, temp := x.div(i)
		if temp.compare(fromInt(0)) == 0 {
			return false
		}
	}
	return true
}

func nextPrime(x *BigInt) *BigInt {
	for {
		x = x.next()
		if isPrime(x) {
			return x
		}
	}
}

func gcd(a, b *BigInt) *BigInt {
	if b.compare(fromInt(0)) == 0 {
		return a
	}
	_, temp := a.div(b)
	return gcd(b, temp)
}
