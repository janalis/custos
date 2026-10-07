<?php
function plainDemo($path, $name, $tpl, $raw, $list) {
    $r = [];
    $r[] = false !== strpos($path, "tmp");
    $r[] = false !== stripos($path, "cache-dir");
    $r[] = 0 === strpos($path, ".");
    $r[] = 0 === stripos($name, "img_");
    $r[] = "index" === $name;
    $r[] = "index" !== $name;
    $r[] = preg_match('/^index$/', $name); // also matches "index\n"
    $r[] = 0 !== strpos($path, "tmp");
    $r[] = false === strpos($path, "tmp");
    $r[] = false !== strpos($path, "tmp");
    if (false !== strpos($path, "tmp") && $name) {}
    $r[] = preg_match('/^index$/i', $name);
    $r[] = preg_match('/tmp$/', $path);
    $r[] = preg_match('/tmp/', $path, $hit);
    $r[] = preg_match('/tmp\d/', $path);

    $r[] = str_replace("__NAME__", $name, $tpl);
    $r[] = str_ireplace("draft", '', $tpl);
    $r[] = str_replace("*", $name, $tpl);
    $r[] = preg_replace('/draft/', '', $tpl, 2);
    $r[] = preg_replace(['/draft/'], '', $tpl);
    $r[] = preg_replace('/^draft/', 'x', $tpl);

    $r[] = ltrim($raw, '0');
    $r[] = rtrim($raw, '/');
    $r[] = preg_replace('#/+$#', '', $raw); // keeps a final newline rtrim would not
    $r[] = trim($raw, '-');
    $r[] = trim($raw, " \t\n\r\v\f");
    $r[] = rtrim($raw, " \t\n\r\v\f");
    $r[] = preg_replace('/^0+/u', '', $raw);
    $r[] = preg_replace('/^a+/i', '', $raw); // also strips A
    $r[] = preg_replace('/^0+/U', '', $raw); // lazy: strips one zero
    $r[] = ltrim($raw, '0');
    $r[] = preg_replace('/^0+/', ' ', $raw);
    $r[] = preg_replace('/^.+/', '', $raw);
    $r[] = preg_replace('/^0+|1+$/', '', $raw);

    $r[] = explode(";", $list);
    $r[] = explode("=>", $list, 3);
    $r[] = preg_split('/\s*,\s*/', $list);
    $r[] = preg_split('/;/u', $list);
    $r[] = preg_split('/;/', $list, -1, PREG_SPLIT_NO_EMPTY);
    return $r;
}
