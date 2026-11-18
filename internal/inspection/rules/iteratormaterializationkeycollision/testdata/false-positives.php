<?php
function rows() { yield 'k' => 2; yield 'k' => 5; } $a = iterator_to_array(rows(), false);
