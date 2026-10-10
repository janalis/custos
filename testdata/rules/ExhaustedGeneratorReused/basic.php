<?php
function values() { yield 1; } $g = values(); iterator_to_array($g); <error descr="Create a fresh generator before traversing again.">foreach ($g as $v) {}</error>
