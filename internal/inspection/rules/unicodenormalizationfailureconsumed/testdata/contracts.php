<?php
$s = Normalizer::normalize($input); <warning descr="Check normalization before consuming text.">strlen($s)</warning>;
