<?php
namespace App {
    function str_replace($s, $r, $subject) { return $subject; }

    function slug($title) {
        $title = str_replace('&', 'and', $title);
        return str_replace(' ', '-', $title);
    }
}

namespace {
    function slug($title) {
        $title = Str_Replace('&', 'and', $title);
        return <warning descr="Fold this str_replace() into the preceding one on the same variable.">STR_REPLACE(' ', '-', $title)</warning>;
    }
}
