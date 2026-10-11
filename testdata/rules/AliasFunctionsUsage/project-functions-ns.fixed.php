<?php
namespace {
    function recode($text) { return strrev($text); }
}

namespace Reports {
    $a = recode($label);
    $b = rtrim($label);
}
