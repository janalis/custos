<?php
function sites($p, $x)
{
    $r[] = false === strpos($p, "tmp");
    $r[] = false === strpos($p, "tmp");
    $r[] = false === strpos($p, "tmp");
    $r[] = false === strpos($p, "tmp");
    $r[] = false !== strpos($p, "tmp");
    $r[] = false !== strpos($p, "tmp");
    $r[] = false !== strpos($p, "tmp");
    $r[] = false === strpos($p, "tmp");
    $r[] = false === strpos($p, "tmp");
    $r[] = false !== strpos($p, "tmp");
    $r[] = false === strpos($p, "tmp");
    $r[] = preg_match('/tmp/', $p) < 1e999;
    $r[] = preg_match('/tmp/', $p) == -1;
    $r[] = preg_match('/tmp/', $p) == 2;
    $r[] = preg_match('/tmp/', $p) === 1.0;
    $r[] = preg_match('/tmp/', $p) == -$x;
    $r[] = preg_match('/tmp/', $p) == '1';
    $r[] = preg_match('/tmp/', $p) == 08;
    $r[] = (preg_match('/tmp/', $p)) === 0;
    $r[] = preg_match('/^a$/', $p) + 1;
    $r[] = @preg_match('/tmp/', $p);
    $r[] = preg_match('/tmp/', $p) instanceof Countable;
    $r[] = 'n' . ("a" !== $p);
    $r[] = false !== strpos($p, "abc") && $p;
    $r[] = preg_match('/^a/m', $p);
    $r[] = preg_match('/abc/A', $p);
    $r[] = str_replace("abc", 'x', $p) == 0;
    return $r;
}

function splits($p)
{
    $r[] = explode(";", $p);
    $r[] = explode(".", $p);
    $r[] = explode("\"", $p);
    $r[] = explode(",", $p, 3);
    $r[] = preg_split('/|/', $p);
    $r[] = preg_split('/[^]/', $p);
    $r[] = preg_split('/\//', $p);
    $r[] = preg_split('/,/', $p, -1);
    $r[] = preg_split('/,/', $p, $limit);
    $r[] = preg_split('/+x/', $p);
    $r[] = preg_replace('/^\+/', '', $p);
    $r[] = preg_replace("/^'+/", '', $p);
    return $r;
}
