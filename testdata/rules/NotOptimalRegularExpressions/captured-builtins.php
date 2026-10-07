<?php
namespace Strings {
    function strpos($h, $n) { return 0; }
    function str_replace($s, $r, $t) { return ''; }
    use function Other\ltrim;
    use function Other\explode;

    $r = [];
    $r[] = <warning descr="Replace with 'false !== \strpos($path, &quot;tmp&quot;)'.">preg_match('/tmp/', $path)</warning>;
    $r[] = <warning descr="Replace with '\str_replace(&quot;__NAME__&quot;, $name, $tpl)'.">preg_replace('/__NAME__/', $name, $tpl)</warning>;
    $r[] = <warning descr="Replace with '\ltrim($raw, '0')'.">preg_replace('/^0+/', '', $raw)</warning>;
    $r[] = <warning descr="Replace with 'rtrim($raw, '/')'.">preg_replace('#/+$#', '', $raw)</warning>;
    $r[] = <warning descr="Replace with '\explode(&quot;,&quot;, $list)'.">preg_split('/,/', $list)</warning>;
}
