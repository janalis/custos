<?php
namespace Stats {
    function count($items) { return 0; }

    $a = \count($rows);
    $b = \count($rows);
}

namespace Imported {
    use function Stats\count;

    $c = \count($rows);
    $d = implode(',', $rows);
}
