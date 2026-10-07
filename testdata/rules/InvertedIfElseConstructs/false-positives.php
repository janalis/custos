<?php
if ($x !== false) { a(); } else { b(); }
if (false == $x) { a(); } else { b(); }
if (false === $untyped) { a(); } else { b(); }
if (!$x): a(); else: b(); endif;
if (!$x) { a(); } else if ($y) { b(); }
if (-$x) { a(); } else { b(); }
if ($x === true) { a(); } else { b(); }
if (!$x): a(); else { b(); } endif;

function emptyElse($ok)
{
    if (!$ok) {
        echo 'failed';
    } else {
        // nothing to do
    }
    if (!$ok) {
        echo 'failed';
    } else {}
}
