<?php
function byteEncoding($body) { header('Content-Length: ' . mb_strlen($body, '8bit')); echo $body; }
function singleByteEncoding($body) { header('Content-Length: ' . mb_strlen($body, 'ISO-8859-1')); echo $body; }
function trailingTemplate($body) { header('Content-Length: ' . mb_strlen($body)); echo $body; ?>
<!-- footer -->
<?php }
function repeatedBody($body) { header('Content-Length: ' . mb_strlen($body)); foreach ([1, 2] as $_) { echo $body; } }
