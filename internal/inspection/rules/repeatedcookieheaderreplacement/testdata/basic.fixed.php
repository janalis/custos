<?php
header('Set-Cookie: a=1'); header('Set-Cookie: b=2', false);

function unrelatedHeaderRemoved()
{
    header('Set-Cookie: theme=dark; Path=/');
    header_remove('X-Powered-By');
    header('Set-Cookie: lang=fr; Path=/', false);
}
