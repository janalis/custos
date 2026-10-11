<?php
header('Set-Cookie: a=1'); <warning descr="Append distinct cookie headers instead of replacing them.">header('Set-Cookie: b=2')</warning>;

function unrelatedHeaderRemoved()
{
    header('Set-Cookie: theme=dark; Path=/');
    header_remove('X-Powered-By');
    <warning descr="Append distinct cookie headers instead of replacing them.">header('Set-Cookie: lang=fr; Path=/')</warning>;
}
