package tygo

import "strings"

func (g *PackageGenerator) IsEnumStruct(name string) bool {
	for _, b := range g.conf.EnumStructs {
		if strings.EqualFold(b, name) {
			return true
		}
	}
	return false
}
