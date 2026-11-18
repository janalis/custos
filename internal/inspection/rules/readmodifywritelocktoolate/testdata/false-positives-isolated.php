<?php
function good($source, $target) { $n = (int) file_get_contents($source); file_put_contents($target, $n + 1, LOCK_EX); }
