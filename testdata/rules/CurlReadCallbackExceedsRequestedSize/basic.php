<?php
<warning descr="Return no more than the requested byte count.">curl_setopt($ch, CURLOPT_READFUNCTION, fn($ch, $h, $limit) => str_repeat('x', $limit + 2))</warning>;
