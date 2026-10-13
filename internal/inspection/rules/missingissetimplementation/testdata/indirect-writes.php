<?php
class Sack
{
}

class Lazy
{
    public function __get($k) { return null; }
}

class Sink
{
    public function __construct(&$out, $in = null) { $out = 1; }
    public function fill(&$out) { $out = 1; }
    public static function make($in, &...$outs) {}
}

function by_value($in) {}
function no_params() {}
function named_out($in, &$out) { $out = 1; }

function indirect_writes(Sack $b, Sink $s, array $rows): array
{
    $b->list[] = 1;
    $b->map['k']['j'] = 1;
    $r = &$b->ref;
    $r = 1;
    preg_match('/a/', 'a', $b->match);
    foreach ($rows as $b->key => $b->last) {
    }
    [, $b->first, [$b->nested]] = $rows;
    list($b->second) = $rows;
    $b->count++;
    $s->fill($b->filled);
    new Sink($b->built);
    Sink::make(1, $rows, $b->tail);
    named_out(out: $b->named, in: 1);
    legacy_out(&$b->calltime);

    by_value($b->read);
    by_value(in: $b->readNamed);
    by_value(other: $b->unknownName);
    no_params($b->extraArg);
    unresolved($b->passed);
    new Sink($rows, $b->ctorRead);
    new Missing($b->noCtor);
    $copy = $b->source;
    $first = $b->element[0];

    return [
        isset($b->list, $b->map, $b->ref, $b->match, $b->key, $b->last),
        isset($b->first, $b->nested, $b->second, $b->count),
        isset($b->filled, $b->built, $b->tail, $b->named, $b->calltime),
        isset(<error descr="\Sack has no __isset(); this isset/empty check is always false.">$b->read</error>),
        isset(<error descr="\Sack has no __isset(); this isset/empty check is always false.">$b->readNamed</error>),
        isset(<error descr="\Sack has no __isset(); this isset/empty check is always false.">$b->unknownName</error>),
        isset(<error descr="\Sack has no __isset(); this isset/empty check is always false.">$b->extraArg</error>),
        isset(<error descr="\Sack has no __isset(); this isset/empty check is always false.">$b->passed</error>),
        isset(<error descr="\Sack has no __isset(); this isset/empty check is always false.">$b->ctorRead</error>),
        isset(<error descr="\Sack has no __isset(); this isset/empty check is always false.">$b->noCtor</error>),
        isset(<error descr="\Sack has no __isset(); this isset/empty check is always false.">$b->source</error>),
        isset(<error descr="\Sack has no __isset(); this isset/empty check is always false.">$b->element</error>),
    ];
}

// With __get() and no __set(), a write through an element access or a
// reference goes to the value __get() returned: no property is created.
function intercepted_writes(Lazy $l, Sink $s, array $rows): array
{
    $l->direct = 1;
    [$l->slot] = $rows;
    $l->count++;
    foreach ($rows as $l->each) {
    }
    $l->list[] = 1;
    $l->map['k'] = 1;
    $r = &$l->ref;
    $r = 1;
    $s->fill($l->filled);
    $l->bumped[0]++;
    [$l->cell[0]] = $rows;

    return [
        isset($l->direct, $l->slot, $l->count, $l->each),
        isset(<error descr="\Lazy has no __isset(); this isset/empty check is always false.">$l->list</error>),
        isset(<error descr="\Lazy has no __isset(); this isset/empty check is always false.">$l->map</error>),
        isset(<error descr="\Lazy has no __isset(); this isset/empty check is always false.">$l->ref</error>),
        isset(<error descr="\Lazy has no __isset(); this isset/empty check is always false.">$l->filled</error>),
        isset(<error descr="\Lazy has no __isset(); this isset/empty check is always false.">$l->bumped</error>),
        isset(<error descr="\Lazy has no __isset(); this isset/empty check is always false.">$l->cell</error>),
    ];
}
