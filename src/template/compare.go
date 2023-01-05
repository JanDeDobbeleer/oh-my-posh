package template

func interFaceToInt(e interface{}) int {
	if val, OK := e.(int); OK {
		return val
	}
	if val, OK := e.(float64); OK {
		return int(val)
	}
	if val, OK := e.(int64); OK {
		return int(val)
	}
	return 0
}

func interfaceToFloat64(e interface{}) float64 {
	if val, OK := e.(float64); OK {
		return val
	}
	if val, OK := e.(int); OK {
		return float64(val)
	}
	if val, OK := e.(int64); OK {
		return float64(val)
	}
	return 0
}

func gt(e1, e2 interface{}) bool {
	if val, OK := e1.(int); OK {
		return val > interFaceToInt(e2)
	}
	if val, OK := e1.(float64); OK {
		return val > interfaceToFloat64(e2)
	}
	return false
}

func lt(e1, e2 interface{}) bool {
	return gt(e2, e1)
}
