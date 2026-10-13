<?php
namespace Formats {
    const DATE_ATOM = 'mine';

    $a = X\DATE_ISO8601;
    $b = <error descr="DATE_ISO8601 is not ISO-8601 compliant; use DATE_ATOM.">DATE_ISO8601</error>;
    $c = <error descr="DATE_ISO8601 is not ISO-8601 compliant; use DATE_ATOM.">\DATE_ISO8601</error>;
}

namespace Own {
    const DATE_ISO8601 = 'mine';

    $d = DATE_ISO8601;
    $e = <error descr="DATE_ISO8601 is not ISO-8601 compliant; use DATE_ATOM.">\DATE_ISO8601</error>;
}

namespace Imported {
    use const DATE_ISO8601 as ISO;
    use const Other\DATE_ISO8601;

    $f = <error descr="DATE_ISO8601 is not ISO-8601 compliant; use DATE_ATOM.">ISO</error>;
    $g = DATE_ISO8601;
}
