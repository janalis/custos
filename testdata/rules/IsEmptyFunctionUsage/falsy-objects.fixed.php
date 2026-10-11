<?php
interface Named {}
interface Plain {}
class Feed extends SimpleXMLElement implements Named {}
class Tag implements Plain {}

function check(?SimpleXMLElement $x, ?Feed $f, ?GMP $n, ?DateTime $d, ?MissingThing $m)
{
    // Empty XML elements and GMP zero are empty: no null comparison.
    if (empty($x) || empty($f) || empty($n)) {
        return 1;
    }
    if ($d === null) {
        return 2;
    }
    return $m === null ? 3 : 0;
}

// A zero BcMath\Number is empty, and so is whatever an interface typed value
// holds when a falsy class (or a class extending one) implements it.
function supertypes(?\BcMath\Number $num, ?\Stringable $s, ?\Countable $c, \Traversable $t, ?\Iterator $i, ?Named $named, ?Plain $plain, ?Tag $tag)
{
    return [
        empty($num), empty($s), empty($c), !empty($t), empty($i), empty($named),
        $plain === null,
        $tag !== null,
    ];
}

function make(): ?DateTime { return null; }

// The replacement must not fuse with the keyword before it.
function glued(?DateTime $d) { return make() !== null; }
function spaced(?DateTime $d) { return$d !== null; }

// A variable bound by reference can change behind its declared type.
function aliased(?DateTime $a, ?DateTime $b, ?DateTime $c, ?DateTime $d)
{
    $r = &$a;
    $r = '';
    $b = &$r;
    $list = [&$c, $d];
    return [empty($a), empty($b), empty($c), $d === null, $list];
}
