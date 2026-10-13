<?php
namespace Lookup {
    function array_key_exists($k, $a) { return false; }

    $a = in_array($id, array_keys($map));
}

namespace Imported {
    use function Lookup\array_key_exists;

    $b = in_array($id, array_keys($map));
}

namespace Plain {
    $c = in_array($id, array_keys($map));
}
