package index

import "custos/internal/phpver"

// addEnumMembers describes members PHP supplies for every enum. It runs
// after extraction so a permissively parsed explicit declaration wins.
func addEnumMembers(c *Class, backing string) {
	if c.Avail.From < phpver.PHP81 {
		c.Avail.From = phpver.PHP81
	}
	c.Interfaces = append(c.Interfaces, "UnitEnum")
	c.Props["name"] = &Property{Name: "name", Class: c.FQN, Type: "string", Readonly: true, Builtin: true}
	add := func(name, ret, doc string, params []Param) {
		if c.Methods[name] == nil {
			c.Methods[name] = &Method{
				Name: name, Class: c.FQN, Static: true, Return: ret, DocReturn: doc,
				Params: params, Builtin: true, Avail: Avail{From: phpver.PHP81},
			}
		}
	}
	typ := `\` + c.FQN
	add("cases", "array", "list<"+typ+">", nil)
	if backing == "" {
		return
	}
	c.Interfaces = append(c.Interfaces, "BackedEnum")
	c.Props["value"] = &Property{Name: "value", Class: c.FQN, Type: backing, Readonly: true, Builtin: true}
	params := []Param{{Name: "value", Type: "int|string"}}
	add("from", typ, "", params)
	add("tryfrom", typ+"|null", "", params)
}
