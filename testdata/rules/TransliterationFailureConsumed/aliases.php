<?php
use function transliterator_transliterate as builtinCall1;
$s=builtinCall1($transform,$input); <warning descr="Check transliteration before consuming text.">strlen($s)</warning>;
