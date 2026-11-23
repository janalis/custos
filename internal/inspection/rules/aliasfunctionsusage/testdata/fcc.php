<?php
// A first-class callable refers to the alias as a call does.
$count = <warning descr="Use 'count(...)' instead of the alias 'sizeof(...)'.">sizeof</warning>(...);
$glue = <warning descr="Use 'implode(...)' instead of the alias 'join(...)'.">join</warning>(...);
