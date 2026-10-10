package infer_test

import "testing"

// Static properties narrow like `$this->prop`, and lose the fact on any
// call or write in between.
func TestStaticPropertyNarrowing(t *testing.T) {
	checkAnywhere(t, `<?php
class W {}
class S {
    /** @var W[] */
    private static array $stack = [];
    private static ?W $cur = null;
    public static function f() {
        if (!empty(self::$stack)) {
            $w = array_pop(self::$stack);
            t('pop', $w);
        }
        if (!empty(static::$stack)) { t('static', static::$stack); }
        if (self::$cur === null) { return; }
        t('cur', self::$cur);
        self::reset();
        t('afterCall', self::$cur);
        if (S::$cur !== null) { t('named', S::$cur); }
        self::$cur = new W();
        t('written', self::$cur);
    }
    public static function reset(): void {}
}
`, map[string]string{
		"pop": `\W`, "static": `non-empty \W[]`, "cur": `\W`, "afterCall": `\W`, // as for $this->p, an early-exit guard survives calls "named": `\W`, "written": `\W`,
	})
}

// Properties of plain variables narrow like `$this->p` (conditions,
// continue/return guards), and lose the fact when the variable changes.
func TestVariablePropertyNarrowing(t *testing.T) {
	checkAnywhere(t, `<?php
class Ref {
    public string|array $value = '';
    public ?Ref $next = null;
    /** @var array<string, int|string> */
    public array $list = [];
}
class H {
    private ?Ref $p = null;
    public function f(array $m, Ref $r, Ref $s) {
        foreach ($m as $ref) {
            /** @var Ref $ref */
            if (!is_string($ref->value)) continue;
            t('continue', str_replace('_', '-', $ref->value));
        }
        foreach ($m as $x) {
            if ($this->p === null) { continue; }
            t('thisProp', $this->p);
        }
        if ($r->next !== null) {
            t('cond', $r->next);
            $r = new Ref();
            t('reassigned', $r->next);
        }
        if ($s->next === null) { return; }
        t('guard', $s->next);
        $s = new Ref();
        t('afterReset', $s->next);
        if (isset($r->list['k'])) { t('dim', $r->list['k']); $r = new Ref(); t('dimReset', $r->list); }
        if (is_int($s->list['n'])) { $s = $r; $v = $s->list['n']; t('dimAfter', $v); }
        if ($r->next instanceof Ref && is_string($r->next->value)) {
            t('chain', $r->next->value);
            $r->next = new Ref();
            t('chainReset', $r->next->value);
        }
        if (is_string($this->p->value)) { t('thisChain', $this->p->value); }
    }
}
`, map[string]string{
		"continue": "string", "thisProp": `\Ref`, "cond": `\Ref`, "reassigned": `\Ref|null`, "guard": `\Ref`,
		"afterReset": `\Ref|null`, "dim": "int|string", "dimReset": "int[]|string[]", "dimAfter": "int|string",
		"chain": "string", "chainReset": "array|string", "thisChain": "string",
	})
}

// `$isObject = is_object($r); if ($isObject) { … }` narrows $r, unless
// either variable changed in between.
func TestBooleanAliasNarrowing(t *testing.T) {
	checkAnywhere(t, `<?php
function f(object|string|null $r, ?int $n, $c) {
    if (($isObject = is_object($r)) && $c) {}
    if ($isObject) { t('true', $r); } else { t('false', $r); }
    $has = $n !== null;
    if (!$has) { return; }
    t('guard', $n);
    $ok = is_int($n);
    $n = $c ? null : 1;
    if ($ok) { t('changed', $n); }
    $flag = is_string($r);
    if ($c) { $flag = true; }
    if ($flag) { t('twoDefs', $r); }
    $copy = $has;
    if ($copy) { t('aliasOfAlias', $n); }
    $call = strlen('x');
    if ($call) { t('notBool', $r); }
    if ($undefined) { t('undefinedAlias', $r); }
    t('replaceUnknown', str_replace('a', 'b', $c));
}
`, map[string]string{
		"true": "object", "false": "null|string", "guard": "int", "changed": "int|null",
		"twoDefs": "null|object|string", "aliasOfAlias": "int|null", "notBool": "null|object|string",
		"undefinedAlias": "null|object|string", "replaceUnknown": "?unknown",
	})
}
