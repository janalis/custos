<?php
// Reading the requested size, or passing an expression built from it, keeps the proof.
<warning descr="Return no more than the requested byte count.">curl_setopt($ch, CURLOPT_READFUNCTION, function ($c, $h, $limit) { error_log('asked ' . $limit); $copy = $limit; $copy--; return str_repeat('x', $limit + 1); })</warning>;
