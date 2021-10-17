package sorter

import "strings"

func skipPackageAlias(pkg string) string {
	values := strings.Split(pkg, " ")
	if len(values) > 1 {
		return strings.Trim(values[1], `"`)
	}

	return strings.Trim(pkg, `"`)
}

func isPackageWithAlias(pkg string) bool {
	values := strings.Split(pkg, " ")

	return len(values) > 1
}

func OrderContainsAlias(order []string) bool {
	for _, o := range order {
		if o == AliasedPkg {
			return true
		}
	}
	return false
}
