<?php
$t=Transliterator::create("Any-Latin");$s=$t->transliterate($text);<warning descr="Check transliteration before consuming text.">strlen($s)</warning>;
