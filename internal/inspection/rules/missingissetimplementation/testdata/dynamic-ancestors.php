<?php
class Row extends \stdClass {}
class XmlNode extends \SimpleXMLElement {}
class Filter extends \Vendor\Missing\InputFilter {}
class Plain {}

function probe(Row $r, XmlNode $x, Filter $f, Plain $p): array
{
    return [
        isset($r->title),
        isset($x->child),
        isset($f->blockedTags),
        isset(<error descr="\Plain has no __isset(); this isset/empty check is always false.">$p->title</error>),
    ];
}
