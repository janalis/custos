package infer_test

import "testing"

func TestTraitCompositionInference(t *testing.T) {
	checkAnywhere(t, `<?php
 trait NumberTrait { function value(): int { return 1; } function unknown() { return 1; } }
 trait TextTrait { function value(): string { return ''; } }
 trait Composed { use TextTrait, NumberTrait { NumberTrait::value insteadof TextTrait; TextTrait::value as text; unknown as copied; } }
 class Receiver { use Composed; }
 class Conflict { use NumberTrait, TextTrait; }
 t('selected', (new Receiver())->value());
 t('alias', (new Receiver())->text());
 t('unknown', (new Receiver())->copied());
 t('conflict', (new Conflict())->value());
 `, map[string]string{"selected": "int", "alias": "string", "unknown": "?unknown", "conflict": "?unknown"})
}

func TestAbstractTraitImplementationInference(t *testing.T) {
	checkAnywhere(t, `<?php
 trait Requirement { abstract function value(): object; }
 trait Implementation { function value(): stdClass { return new stdClass; } }
 class Forward { use Requirement, Implementation; }
 class Reverse { use Implementation, Requirement; }
 class ParentClass { function value(): stdClass { return new stdClass; } }
 class Inherited extends ParentClass { use Requirement; }
 abstract class AbstractContract { use Requirement; }
 function inspect(AbstractContract $c) { t('contract',$c->value()); }
 t('forward', (new Forward())->value());
 t('reverse', (new Reverse())->value());
 t('inherited', (new Inherited())->value());
 `, map[string]string{"forward": "\\stdClass", "reverse": "\\stdClass", "inherited": "\\stdClass", "contract": "object"})
}
