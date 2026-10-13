<?php
namespace Formats {
    const DATE_ATOM = 'mine';

    $a = X\DATE_ISO8601;
    $b = \DATE_ATOM;
    $c = \DATE_ATOM;
}

namespace Own {
    const DATE_ISO8601 = 'mine';

    $d = DATE_ISO8601;
    $e = \DATE_ATOM;
}

namespace Imported {
    use const DATE_ISO8601 as ISO;
    use const Other\DATE_ISO8601;

    $f = DATE_ATOM;
    $g = DATE_ISO8601;
}
