<?php
if ($x !== false) { a(); } else { b(); }
if (false == $x) { a(); } else { b(); }
if (false === $untyped) { a(); } else { b(); }
if (!$x): a(); else: b(); endif;
if (!$x) { a(); } else if ($y) { b(); }
