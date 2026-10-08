<?php
// A subclass instance may carry the property (or __isset()).
class AppException extends Exception
{
    public $errorcode = '';
}

function faultCode($fault)
{
    if ($fault instanceof Exception) {
        return isset($fault->errorcode) ? $fault->errorcode : null;
    }
    return null;
}

function tagName(DOMNode $node): string
{
    return isset($node->tagName) ? $node->tagName : $node->nodeName;
}

class Base {}
class Magic extends Base { public function __isset($n) { return true; } }
class Mid extends Base {}
#[\AllowDynamicProperties]
class Dyn extends Mid {}
class Leaf extends Mid {}
class Other extends Leaf {}
final class Closed {}

function probes(Base $b, Leaf $l, Closed $c)
{
    return [
        isset($b->anything),
        isset(<error descr="\Leaf has no __isset(); this isset/empty check is always false.">$l->anything</error>),
        isset(<error descr="\Closed has no __isset(); this isset/empty check is always false.">$c->anything</error>),
    ];
}
