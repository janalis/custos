<?php
namespace Lib;

use Vendor\Missing\Collection;

class Paged extends Collection {}
class Folders extends Paged {}
class Plain {}

function first(Folders $f, Plain $p)
{
    $a = $f[0]; // the missing ancestor may implement ArrayAccess
    $b = <error descr="'$p' does not support offset access (types: \Lib\Plain).">$p[0]</error>;
    return [$a, $b];
}
