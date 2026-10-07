<?php
namespace App {
    function date($f, $t = null) { return ''; }
    function time() { return 0; }

    echo date('Y', \time());
    echo \date('Y', time());
}

namespace {
    echo Date('Y');
}
