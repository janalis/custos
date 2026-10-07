<?php
$total = <warning descr="Use 'count(...)' instead of the alias 'sizeof(...)'.">sizeof</warning>($rows);
$csv   = \<warning descr="Use 'implode(...)' instead of the alias 'join(...)'.">join</warning>(';', $cells);
$ok    = <warning descr="Use 'is_int(...)' instead of the alias 'is_long(...)'.">is_long</warning>($port) && <warning descr="Use 'is_writable(...)' instead of the alias 'is_writeable(...)'.">is_writeable</warning>($dir);
<warning descr="'magic_quotes_runtime(...)' is a legacy alias (deprecated 5.3, removed 7.0); stop relying on it.">magic_quotes_runtime</warning>(0);
