<?php
function settings(array $env) {
    $base = <weak_warning descr="Avoid relying on the value returned by an included file.">require __DIR__ . '/base.php'</weak_warning>;
    merge(<weak_warning descr="Avoid relying on the value returned by an included file.">include_once 'extra.php'</weak_warning>);
    $env['db'] = <weak_warning descr="Avoid relying on the value returned by an included file.">require_once('db.php')</weak_warning>;
    if (true !== <weak_warning descr="Avoid relying on the value returned by an included file.">include 'optional.php'</weak_warning>) {
        log_missing();
    }
    $list = [<weak_warning descr="Avoid relying on the value returned by an included file.">include 'a.php'</weak_warning>];
    $x = (<weak_warning descr="Avoid relying on the value returned by an included file.">require 'b.php'</weak_warning>);
    $y = @<weak_warning descr="Avoid relying on the value returned by an included file.">include 'c.php'</weak_warning>;
    return <weak_warning descr="Avoid relying on the value returned by an included file.">include('defaults.php')</weak_warning>;
}
