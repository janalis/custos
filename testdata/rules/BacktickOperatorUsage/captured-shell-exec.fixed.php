<?php
namespace Ops {
    function shell_exec($cmd) { return ''; }

    $a = \shell_exec("uptime");
}

namespace Imported {
    use function Ops\shell_exec;

    $b = \shell_exec("whoami");
}
