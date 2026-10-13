<?php
use Normalizer as BuiltinClass7;
$s = BuiltinClass7::normalize($input); <warning descr="Check normalization before consuming text.">strlen($s)</warning>;
