<?php
$s=transliterator_transliterate($transform,$input); <warning descr="Check transliteration before consuming text.">strlen($s)</warning>;
