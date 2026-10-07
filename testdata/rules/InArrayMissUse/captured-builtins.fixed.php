<?php
namespace Lookup {
    function array_key_exists($k, $a) { return false; }

    $a = \array_key_exists($id, $map);
}

namespace Imported {
    use function Lookup\array_key_exists;

    $b = \array_key_exists($id, $map);
}

namespace Plain {
    $c = array_key_exists($id, $map);
}
