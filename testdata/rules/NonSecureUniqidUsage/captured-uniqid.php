<?php
namespace Ids {
    function uniqid($prefix = '') { return $prefix; }

    $a = array_map(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'uniqid'</error>, ['a']);
    $b = uniqid('own');
}

namespace Plain {
    $c = array_map(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'uniqid'</error>, ['b']);
}
