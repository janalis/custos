<?php
namespace Lookup {
    function array_key_exists($k, $a) { return false; }

    $a = <warning descr="Look the key up directly with '\array_key_exists($id, $map)'.">in_array($id, array_keys($map))</warning>;
}

namespace Imported {
    use function Lookup\array_key_exists;

    $b = <warning descr="Look the key up directly with '\array_key_exists($id, $map)'.">in_array($id, array_keys($map))</warning>;
}

namespace Plain {
    $c = <warning descr="Look the key up directly with 'array_key_exists($id, $map)'.">in_array($id, array_keys($map))</warning>;
}
