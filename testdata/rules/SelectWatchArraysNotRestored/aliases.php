<?php
use function stream_select as builtinCall13;
$read=[$a,$b]; $write=null; $except=null; while ($running) { <warning descr="Restore stream watch arrays before each selection.">builtinCall13($read,$write,$except,1)</warning>; }
