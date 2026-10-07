<?php
namespace App {
    function call_user_func($f, ...$args) { return null; }

    call_user_func('trim', $to);
}

namespace {
    <weak_warning descr="Call it directly: 'trim($to)'.">Call_User_Func('trim', $to)</weak_warning>;
    <weak_warning descr="Pass the arguments inline: 'call_user_func($hook, $to)'.">CALL_USER_FUNC_ARRAY($hook, [$to])</weak_warning>;
}
