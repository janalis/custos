<?php
function differentBody($body) { header('Content-Length: '.mb_strlen($body)); echo 'other'; }
function changedBody($body) { header('Content-Length: '.mb_strlen($body)); $body = 'new'; echo $body; }
function multipleChunks($body) { header('Content-Length: '.mb_strlen($body)); echo $body, 'tail'; }
function compressedBody($body) { ob_start('compress'); header('Content-Length: '.mb_strlen($body)); echo $body; }
header('Content-Length: '.mb_strlen('ascii')); echo 'ascii';
