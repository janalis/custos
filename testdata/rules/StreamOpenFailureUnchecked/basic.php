<?php
function bad($path) { $h = fopen($path, 'rb'); <warning descr="Check that the stream opened successfully.">fread($h, 12)</warning>; }
