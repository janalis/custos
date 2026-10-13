<?php

function removedBeforeBothHeaders()
{
    header_remove();
    header('Set-Cookie: theme=dark');
    <warning descr="Append distinct cookie headers instead of replacing them.">header('set-cookie:lang=fr', true)</warning>;
}
