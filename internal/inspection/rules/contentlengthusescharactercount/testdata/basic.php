<?php
function bad($body) { <warning descr="Measure response content length in bytes.">header('Content-Length: ' . mb_strlen($body, 'UTF-8'))</warning>; echo $body; }
