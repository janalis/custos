<?php
function sites($p, $x)
{
    $r[] = <warning descr="Replace with 'false === strpos($p, &quot;tmp&quot;)'.">preg_match('/tmp/', $p) < 1</warning>;
    $r[] = <warning descr="Replace with 'false === strpos($p, &quot;tmp&quot;)'.">preg_match('/tmp/', $p) != 1</warning>;
    $r[] = <warning descr="Replace with 'false === strpos($p, &quot;tmp&quot;)'.">1 > preg_match('/tmp/', $p)</warning>;
    $r[] = <warning descr="Replace with 'false === strpos($p, &quot;tmp&quot;)'.">preg_match('/tmp/', $p) <= 0</warning>;
    $r[] = <warning descr="Replace with 'false !== strpos($p, &quot;tmp&quot;)'.">preg_match('/tmp/', $p) >= 1</warning>;
    $r[] = <warning descr="Replace with 'false !== strpos($p, &quot;tmp&quot;)'.">0 < preg_match('/tmp/', $p)</warning>;
    $r[] = <warning descr="Replace with 'false !== strpos($p, &quot;tmp&quot;)'.">preg_match('/tmp/', $p) === 1</warning>;
    $r[] = <warning descr="Replace with 'false === strpos($p, &quot;tmp&quot;)'.">preg_match('/tmp/', $p) !== 0x1</warning>;
    $r[] = <warning descr="Replace with 'false === strpos($p, &quot;tmp&quot;)'.">preg_match('/tmp/', $p) < 0.5</warning>;
    $r[] = <warning descr="Replace with 'false !== strpos($p, &quot;tmp&quot;)'.">1 <= preg_match('/tmp/', $p)</warning>;
    $r[] = <warning descr="Replace with 'false === strpos($p, &quot;tmp&quot;)'.">0 >= preg_match('/tmp/', $p)</warning>;
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
    $r[] = 'n' . <warning descr="Replace with '&quot;a&quot; !== $p'.">!preg_match('/^a$/D', $p)</warning>;
    $r[] = <warning descr="Replace with 'false !== strpos($p, &quot;abc&quot;)'.">preg_match('/abc/', subject: $p)</warning> && $p;
    $r[] = preg_match('/^a/m', $p);
    $r[] = preg_match('/abc/A', $p);
    $r[] = <warning descr="Replace with 'str_replace(&quot;abc&quot;, 'x', $p)'.">preg_replace('/abc/', 'x', $p)</warning> == 0;
    return $r;
}

function splits($p)
{
    $r[] = <warning descr="Replace with 'explode(&quot;;&quot;, $p)'.">preg_split('/[;]/', $p)</warning>;
    $r[] = <warning descr="Replace with 'explode(&quot;.&quot;, $p)'.">preg_split('/[.]/', $p)</warning>;
    $r[] = <warning descr="Replace with 'explode(&quot;\&quot;&quot;, $p)'.">preg_split('/"/', $p)</warning>;
    $r[] = <warning descr="Replace with 'explode(&quot;,&quot;, $p, 3)'.">preg_split('/,/', $p, 3)</warning>;
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
