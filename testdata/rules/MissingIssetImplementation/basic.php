<?php
namespace Acme;

/** @property string $documented */
class Profile
{
    public $nick;
    private $secret;
}

class Lazy
{
    public function __get($k) { return 1; }
    public function __isset($k) { return true; }
}

class LazyChild extends Lazy {}

trait Probing { public function __isset($k) { return false; } }
class Traited { use Probing; }

class Bag
{
    public function __get($k) { return null; }

    public function probe($unknown): array
    {
        $p = new Profile();
        $l = new LazyChild();
        $t = new Traited();
        $b = new Bag();
        $o = new \STDCLASS();
        $x = new \SimpleXMLElement('<a/>');
        $k = 'x';
        return [
            isset($p->nick),
            isset($p->secret),
            isset($p->documented),
            isset(<error descr="\Acme\Profile has no __isset(); this isset/empty check is always false.">$p->color</error>),
            isset($l->anything),
            isset($t->anything),
            isset(<error descr="\Acme\Bag has no __isset(); this isset/empty check is always false.">$b->color</error>, $b->{$k}),
            empty(<error descr="\Acme\Bag has no __isset(); this isset/empty check is always false.">$b->size</error>),
            empty($b->$k),
            isset($b?->color),
            isset($this->dynamic),
            isset($o->field),
            isset($x->node),
            isset($unknown->field),
            isset(Bag::${$k}),
            isset($k[$b->color]),
        ];
    }
}
