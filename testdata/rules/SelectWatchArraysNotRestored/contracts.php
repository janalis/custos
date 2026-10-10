<?php
$read=[$a,$b]; $write=null; $except=null; while ($running) { <warning descr="Restore stream watch arrays before each selection.">stream_select($read,$write,$except,1)</warning>; }
