<?php
namespace {
    function recode($text) { return strrev($text); }
}

namespace Reports {
    $a = recode($label);
    $b = <warning descr="Use 'rtrim(...)' instead of the alias 'chop(...)'.">chop</warning>($label);
}
