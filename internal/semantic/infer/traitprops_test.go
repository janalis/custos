package infer_test

import "testing"

// Inside a trait, `$this->p` takes the type the trait declares for p.
func TestTraitOwnProperties(t *testing.T) {
	checkWith(t, nil, `<?php
class Dto { public function name(): string { return ''; } }
trait Inner { protected ?Dto $inner = null; }
/** @property Dto $magic */
trait HasDto {
    use Inner;
    private Dto $dto;
    /** @var list<Dto> */
    private array $all = [];
    private ?self $parent = null;
    /** @var HasDto|null */
    private $named;
    private static Dto $shared;
    public $untyped;
    public function run() {
        t('dto', $this->dto);
        t('call', $this->dto->name());
        t('doc', $this->all);
        t('inherited', $this->inner);
        t('self', $this->parent);
        t('named', $this->named);
        t('magic', $this->magic);
        t('undeclared', $this->nothing);
        t('untyped', $this->untyped);
        t('dynamic', $this->{'dto'});
        $fn = function () { t('closure', $this->dto); };
    }
}
class User { use HasDto; public function f($o) { t('other', $o->dto); } }
`, map[string]string{
		"dto": `\Dto`, "call": "string", "doc": `\Dto[]`, "inherited": `\Dto|null`,
		"self": "?unknown", "named": "?unknown", "magic": "?unknown", "undeclared": "?unknown", "untyped": "?unknown",
		"dynamic": "?unknown", "closure": `\Dto`, "other": "?unknown",
	})
}
