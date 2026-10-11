<?php
// The list keys 0 and 1 loosely equal the case literals, so the key works.
function digit($c) { switch ($c) { case '0': return 'zero'; case '1': return 'one'; } }
echo digit(array_rand(['0', '1']));
echo digit(array_rand(['1', '0']));
function decimal($c) { switch ($c) { case '0.0': return 'zero'; case ' 1': return 'one'; } }
echo decimal(array_rand(['0.0', ' 1']));
