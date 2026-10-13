<?php
$g=(function(){yield 1;return 2;})(); <warning descr="Finish the generator before reading its return value.">$g->getReturn()</warning>;
