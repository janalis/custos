<?php
$header = fread($stream, 4); <warning descr="Read the complete record before decoding it.">unpack('Nsize', $header)</warning>;
