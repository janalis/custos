<?php
function values() { yield 1; } $g = values(); iterator_to_array($g); foreach (values() as $v) {}
