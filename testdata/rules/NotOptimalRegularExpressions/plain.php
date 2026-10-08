<?php
function plainDemo($path, $name, $tpl, $raw, $list) {
    $r = [];
    $r[] = <warning descr="Replace with 'false !== strpos($path, &quot;tmp&quot;)'.">preg_match('/tmp/', $path)</warning> && $name;
    $r[] = <warning descr="Replace with 'false !== stripos($path, &quot;cache-dir&quot;)'.">preg_match('#cache-dir#i', $path)</warning> && $name;
    $r[] = <warning descr="Replace with '0 === strpos($path, &quot;.&quot;)'.">preg_match('/^\./', $path)</warning> && $name;
    $r[] = <warning descr="Replace with '0 === stripos($name, &quot;img_&quot;)'.">preg_match('/^img_/i', $name)</warning> && $name;
    $r[] = <warning descr="Replace with '&quot;index&quot; === $name'.">preg_match('/^index$/D', $name)</warning> && $name;
    $r[] = <warning descr="Replace with '&quot;index&quot; !== $name'.">!preg_match('/^index$/D', $name)</warning>;
    $r[] = preg_match('/^index$/', $name); // also matches "index\n"
    $r[] = <warning descr="Replace with '0 !== strpos($path, &quot;tmp&quot;)'.">preg_match('/^tmp/', $path) == 0</warning>;
    $r[] = <warning descr="Replace with 'false === strpos($path, &quot;tmp&quot;)'.">0 === preg_match('/tmp/', $path)</warning>;
    $r[] = <warning descr="Replace with 'false !== strpos($path, &quot;tmp&quot;)'.">preg_match('/tmp/', $path) > 0</warning>;
    if (<warning descr="Replace with 'false !== strpos($path, &quot;tmp&quot;)'.">preg_match('/tmp/', $path)</warning> && $name) {}
    $r[] = preg_match('/^index$/i', $name);
    $r[] = preg_match('/tmp$/', $path);
    $r[] = preg_match('/tmp/', $path, $hit);
    $r[] = preg_match('/tmp\d/', $path);
    $r[] = preg_match('/tmp/', $path); // 1/0: a boolean test would change the value
    echo preg_match('/tmp/', $path);

    $r[] = <warning descr="Replace with 'str_replace(&quot;__NAME__&quot;, 'Name', $tpl)'.">preg_replace('/__NAME__/', 'Name', $tpl)</warning>;
    $r[] = <warning descr="Replace with 'str_ireplace(&quot;draft&quot;, '', $tpl)'.">preg_replace('/draft/i', '', $tpl)</warning>;
    $r[] = <warning descr="Replace with 'str_replace(&quot;*&quot;, 7, $tpl)'.">preg_replace('/\*/', 7, $tpl)</warning>;
    $r[] = preg_replace('/draft/', '', $tpl, 2);
    $r[] = preg_replace(['/draft/'], '', $tpl);
    $r[] = preg_replace('/^draft/', 'x', $tpl);

    $r[] = <warning descr="Replace with 'ltrim($raw, '0')'.">preg_replace('/^0+/', '', $raw)</warning>;
    $r[] = <warning descr="Replace with 'rtrim($raw, '/')'.">preg_replace('#/+$#D', '', $raw)</warning>;
    $r[] = preg_replace('#/+$#', '', $raw); // keeps a final newline rtrim would not
    $r[] = <warning descr="Replace with 'trim($raw, '-')'.">preg_replace('/^-*|-+$/D', '', $raw)</warning>;
    $r[] = <warning descr="Replace with 'trim($raw, &quot; \t\n\r\v\f&quot;)'.">preg_replace('/^\s+|\s+$/', "", $raw)</warning>;
    $r[] = <warning descr="Replace with 'rtrim($raw, &quot; \t\n\r\v\f&quot;)'.">preg_replace('/\s*$/', '', $raw)</warning>;
    $r[] = preg_replace('/^0+/u', '', $raw);
    $r[] = preg_replace('/^a+/i', '', $raw); // also strips A
    $r[] = preg_replace('/^0+/U', '', $raw); // lazy: strips one zero
    $r[] = <warning descr="Replace with 'ltrim($raw, '0')'.">preg_replace('/^0+/S', '', $raw)</warning>;
    $r[] = preg_replace('/^0+/', ' ', $raw);
    $r[] = preg_replace('/^.+/', '', $raw);
    $r[] = preg_replace('/^0+|1+$/', '', $raw);

    $r[] = <warning descr="Replace with 'explode(&quot;;&quot;, $list)'.">preg_split('/;/', $list)</warning>;
    $r[] = <warning descr="Replace with 'explode(&quot;=>&quot;, $list, 3)'.">preg_split('/=>/', $list, 3)</warning>;
    $r[] = preg_split('/\s*,\s*/', $list);
    $r[] = preg_split('/;/u', $list);
    $r[] = preg_split('/;/', $list, -1, PREG_SPLIT_NO_EMPTY);
    return $r;
}
function replacementsWithReferences(string $s, string $r, $db)
{
    // preg_replace() interprets $0, \0, ${1} and \\ in the replacement.
    $a = preg_replace('/abc/', '$0$0', $s);
    $b = preg_replace('/abc/', '\\0x', $s);
    $c = preg_replace('/abc/', $r, $s);
    $d = preg_replace('/__HANDLER__/i', "'" . $db->escape($r) . "'", $s);
    $e = <warning descr="Replace with 'str_replace(&quot;abc&quot;, (int) $r, $s)'.">preg_replace('/abc/', (int) $r, $s)</warning>;
    return [$a, $b, $c, $d, $e];
}
