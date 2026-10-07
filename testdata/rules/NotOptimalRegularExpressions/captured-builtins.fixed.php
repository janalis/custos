<?php
namespace Strings {
    function strpos($h, $n) { return 0; }
    function str_replace($s, $r, $t) { return ''; }
    use function Other\ltrim;
    use function Other\explode;

    $r = [];
    $r[] = false !== \strpos($path, "tmp");
    $r[] = \str_replace("__NAME__", $name, $tpl);
    $r[] = \ltrim($raw, '0');
    $r[] = rtrim($raw, '/');
    $r[] = \explode(",", $list);
}
