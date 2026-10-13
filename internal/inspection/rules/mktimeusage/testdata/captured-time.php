<?php
namespace Clock {
    function time() { return 0; }

    $a = <warning descr="Call time() instead; mktime()/gmmktime() without arguments is deprecated.">mktime()</warning>;
}

namespace Imported {
    use function Clock\time;

    $b = <warning descr="Call time() instead; mktime()/gmmktime() without arguments is deprecated.">gmmktime()</warning>;
}

namespace Plain {
    $c = <warning descr="Call time() instead; mktime()/gmmktime() without arguments is deprecated.">mktime()</warning>;
    $d = <warning descr="Call time() instead; mktime()/gmmktime() without arguments is deprecated.">\mktime()</warning>;
}
