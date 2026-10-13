<?php
$g = (function() { yield 7; yield 11; })(); $g->next(); <error descr="Create another generator instead of rewinding it.">$g->rewind()</error>;
