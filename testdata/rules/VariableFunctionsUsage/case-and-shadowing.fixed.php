<?php
namespace App {
    function call_user_func($f, ...$args) { return null; }

    call_user_func('trim', $to);
}

namespace {
    trim($to);
    call_user_func($hook, $to);
}
