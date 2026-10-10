<?php
$g=(function(){yield 1;return 2;})(); $g->next(); $g->getReturn();
