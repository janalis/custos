package syntax

import phpversion "custos/internal/php/version"

// MinVersion returns the minimum supported PHP version for selected modern
// syntax features. Other nodes return the engine's baseline version. This is
// feature metadata; it does not change parsing or emit compatibility reports.
func MinVersion(n Node) phpversion.Version {
	switch x := n.(type) {
	case *PropertyHook:
		return phpversion.PHP84
	case *ClassConst:
		if x.Type != nil {
			return phpversion.PHP83
		}
	case *Property:
		if asymmetricVisibility(x.Modifiers) {
			return phpversion.PHP84
		}
	case *Param:
		if asymmetricVisibility(x.Modifiers) {
			return phpversion.PHP84
		}
	case *New:
		switch p := x.Parent().(type) {
		case *FuncCall:
			if p.Name == x {
				return phpversion.PHP84
			}
		case *MethodCall:
			if p.Var == x {
				return phpversion.PHP84
			}
		case *PropertyFetch:
			if p.Var == x {
				return phpversion.PHP84
			}
		case *ArrayDimFetch:
			if p.Var == x {
				return phpversion.PHP84
			}
		case *StaticCall:
			if p.Class == x {
				return phpversion.PHP84
			}
		case *StaticPropertyFetch:
			if p.Class == x {
				return phpversion.PHP84
			}
		case *ClassConstFetch:
			if p.Class == x {
				return phpversion.PHP84
			}
		}
	}
	return phpversion.Min
}

func asymmetricVisibility(mods Modifiers) bool {
	return mods.Has(TPrivateSet) || mods.Has(TProtectedSet) || mods.Has(TPublicSet)
}
