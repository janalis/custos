<?php
namespace Ops {
    function shell_exec($cmd) { return ''; }

    $a = <warning descr="Run the command through shell_exec() instead of backticks.">`uptime`</warning>;
}

namespace Imported {
    use function Ops\shell_exec;

    $b = <warning descr="Run the command through shell_exec() instead of backticks.">`whoami`</warning>;
}
