package listUtils

// InList[T comparable]
//
//	@Description: 通用的用于判断是否在列表中
//	@param list
//	@param value
//	@return bool
func InList[T comparable](list []T, value T) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
